package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/mux"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/auth"
	"github.com/pivotal-cf/brokerapi/domain"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const MountConfigKey = "name"
const MountBindOptionKey = "mount"
const DefaultMountPath = "/home/vcap/data/"
const ServiceID = "6cb45412-8161-44ec-b462-e3fd08f55448"
const PlanID = "e805eb41-4fb4-485a-9066-b0edf57b90b3"
const DefaultUID = "1000"
const DefaultGID = "1000"
const UncPathRegex = `^[\\\/]{2,}[^\\\/]+[\\\/]+[^\\\/]+`

func BrokerHandler(namespace string, pv corev1.PersistentVolumeInterface, pvc corev1.PersistentVolumeClaimInterface, secret corev1.SecretInterface, username string, password string, logger lager.Logger) (http.Handler, error) {
	router := mux.NewRouter()
	router.Use(auth.NewWrapper(username, password).Wrap)

	brokerapi.AttachRoutes(router, smbServiceBroker{
		PersistentVolume:      pv,
		PersistentVolumeClaim: pvc,
		Secret:                secret,
		Namespace:             namespace,
		Logger:                logger,
	}, logger)
	return router, nil
}

type smbServiceBroker struct {
	PersistentVolume      corev1.PersistentVolumeInterface
	PersistentVolumeClaim corev1.PersistentVolumeClaimInterface
	Secret                corev1.SecretInterface
	Namespace             string
	Logger                lager.Logger
}

func (s smbServiceBroker) Services(ctx context.Context) ([]domain.Service, error) {
	return []domain.Service{{
		ID:                   ServiceID,
		Name:                 "SMB",
		Description:          "SMB for K8s",
		Bindable:             true,
		InstancesRetrievable: true,
		BindingsRetrievable:  false,
		Tags:                 []string{"pivotal", "smb", "volume-services"},
		PlanUpdatable:        false,
		Plans: []domain.ServicePlan{
			{
				Description: "The only SMB Plan",
				ID:          PlanID,
				Name:        "Existing",
				Metadata: &domain.ServicePlanMetadata{
					DisplayName: "SMB",
				},
			},
		},
		Requires:        []domain.RequiredPermission{"volume_mount"},
		Metadata:        &domain.ServiceMetadata{},
		DashboardClient: nil,
	}}, nil
}

func (s smbServiceBroker) Provision(ctx context.Context, instanceID string, details domain.ProvisionDetails, asyncAllowed bool) (_ domain.ProvisionedServiceSpec, err error) {
	defer func() {
		if err != nil {
			s.cleanupResourcesCreatedByProvision(ctx, instanceID)
		}
	}()

	var serviceInstanceParameters map[string]interface{}

	if details.RawParameters != nil {
		var decoder = json.NewDecoder(bytes.NewBuffer(details.RawParameters))
		err := decoder.Decode(&serviceInstanceParameters)
		if err != nil {
			return domain.ProvisionedServiceSpec{}, errors.New("unable to decode service instance parameters")
		}
	}

	username, err := getAttribute(serviceInstanceParameters, "username")
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}
	password, err := getAttribute(serviceInstanceParameters, "password")
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}
	share, err := getAttribute(serviceInstanceParameters, "share")
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	re := regexp.MustCompile(UncPathRegex)
	match := re.MatchString(share)

	if match == false {
		return domain.ProvisionedServiceSpec{}, invalidParametersResponse("share must be a UNC path")
	}

	storageClass := ""

	_, err = s.PersistentVolumeClaim.Create(
		ctx,
		&v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: instanceID,
			},
			Spec: v1.PersistentVolumeClaimSpec{
				StorageClassName: &storageClass,
				VolumeName:       instanceID,
				AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse("1M")},
				},
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	_, err = s.Secret.Create(
		ctx,
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: instanceID,
			},
			StringData: map[string]string{
				"username": username,
				"password": password,
			},
		},
		metav1.CreateOptions{})
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	volumeAttributesMap := map[string]string{"share": share, "uid": DefaultUID, "gid": DefaultGID}
	if vers, found := serviceInstanceParameters["vers"]; found {
		var ok bool
		if volumeAttributesMap["vers"], ok = vers.(string); !ok {
			resp := apiresponses.NewFailureResponse(errors.New("'vers' configuration value must be a string"), http.StatusBadRequest, "")
			return domain.ProvisionedServiceSpec{}, resp
		}
	}

	_, err = s.PersistentVolume.Create(
		ctx,
		&v1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: instanceID,
			},
			Spec: v1.PersistentVolumeSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
				Capacity:    v1.ResourceList{v1.ResourceStorage: resource.MustParse("100M")},
				PersistentVolumeSource: v1.PersistentVolumeSource{
					CSI: &v1.CSIPersistentVolumeSource{
						Driver:           "org.cloudfoundry.smb",
						VolumeHandle:     "volume-handle",
						VolumeAttributes: volumeAttributesMap,
						NodePublishSecretRef: &v1.SecretReference{
							Name:      instanceID,
							Namespace: s.Namespace,
						},
					},
				},
			},
		},
		metav1.CreateOptions{})

	return domain.ProvisionedServiceSpec{}, err
}

func (s smbServiceBroker) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, asyncAllowed bool) (domain.DeprovisionServiceSpec, error) {
	err := s.Secret.Delete(ctx, instanceID, metav1.DeleteOptions{})
	if err != nil {
		return domain.DeprovisionServiceSpec{}, err
	}

	err = s.PersistentVolumeClaim.Delete(ctx, instanceID, metav1.DeleteOptions{})
	if err != nil {
		return domain.DeprovisionServiceSpec{}, err
	}

	err = s.PersistentVolume.Delete(ctx, instanceID, metav1.DeleteOptions{})
	if err != nil {
		return domain.DeprovisionServiceSpec{}, err
	}

	return domain.DeprovisionServiceSpec{}, nil
}

func (s smbServiceBroker) GetInstance(ctx context.Context, instanceID string) (domain.GetInstanceDetailsSpec, error) {
	pv, err := s.PersistentVolume.Get(ctx, instanceID, metav1.GetOptions{})
	if err != nil {
		return domain.GetInstanceDetailsSpec{}, apiresponses.NewFailureResponse(errors.New("unable to find service instance"), 404, "")
	}

	secret, err := s.Secret.Get(ctx, instanceID, metav1.GetOptions{})
	if err != nil {
		return domain.GetInstanceDetailsSpec{}, apiresponses.NewFailureResponse(errors.New("unable to establish username"), 404, "")
	}
	username := secret.Data["username"]

	parametersInstanceDetailsMap := map[string]interface{}{}
	parametersInstanceDetailsMap["share"] = pv.Spec.PersistentVolumeSource.CSI.VolumeAttributes["share"]
	parametersInstanceDetailsMap["username"] = string(username)

	return domain.GetInstanceDetailsSpec{
		ServiceID:  ServiceID,
		PlanID:     PlanID,
		Parameters: parametersInstanceDetailsMap,
	}, nil
}

func (s smbServiceBroker) Update(ctx context.Context, instanceID string, details domain.UpdateDetails, asyncAllowed bool) (domain.UpdateServiceSpec, error) {
	panic("implement me")
}

func (s smbServiceBroker) LastOperation(ctx context.Context, instanceID string, details domain.PollDetails) (domain.LastOperation, error) {
	panic("implement me")
}

func (s smbServiceBroker) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, asyncAllowed bool) (domain.Binding, error) {

	_, err := s.PersistentVolume.Get(ctx, instanceID, metav1.GetOptions{})
	if err != nil {
		return domain.Binding{}, apiresponses.ErrInstanceDoesNotExist
	}

	mountConfig := map[string]interface{}{
		MountConfigKey: instanceID,
	}

	var bindOpts = map[string]string{
		MountBindOptionKey: DefaultMountPath,
	}

	if len(details.RawParameters) > 0 {
		err := json.Unmarshal(details.RawParameters, &bindOpts)
		if err != nil {
			return domain.Binding{}, apiresponses.ErrRawParamsInvalid
		}
	}

	device := domain.SharedDevice{
		VolumeId:    bindingID,
		MountConfig: mountConfig,
	}
	volumeMount := domain.VolumeMount{
		Driver:       "smb",
		ContainerDir: bindOpts[MountBindOptionKey],
		Mode:         "rw",
		DeviceType:   "shared",
		Device:       device,
	}

	var volumeMounts []domain.VolumeMount
	volumeMounts = append(volumeMounts, volumeMount)
	binding := domain.Binding{
		Credentials:  struct{}{}, // if nil, cloud controller chokes on response
		VolumeMounts: volumeMounts,
	}
	return binding, nil
}

func (s smbServiceBroker) Unbind(ctx context.Context, instanceID, bindingID string, details domain.UnbindDetails, asyncAllowed bool) (domain.UnbindSpec, error) {
	_, err := s.PersistentVolume.Get(ctx, instanceID, metav1.GetOptions{})
	if err != nil {
		return domain.UnbindSpec{}, apiresponses.ErrInstanceDoesNotExist
	}

	return domain.UnbindSpec{}, nil
}

func (s smbServiceBroker) GetBinding(ctx context.Context, instanceID, bindingID string) (domain.GetBindingSpec, error) {
	panic("implement me")
}

func (s smbServiceBroker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details domain.PollDetails) (domain.LastOperation, error) {
	panic("implement me")
}

func invalidParametersResponse(description string) *apiresponses.FailureResponse {
	return apiresponses.NewFailureResponse(errors.New(description), 400, "invalid-parameters")
}

func (s smbServiceBroker) containsKey(serviceInstanceParameters map[string]interface{}, key string) bool {
	_, found := serviceInstanceParameters[key]
	return found
}

func getAttribute(source map[string]interface{}, key string) (string, error) {
	if valueFromSource, found := source[key]; found {
		if val, ok := valueFromSource.(string); !ok {
			return "", invalidParametersResponse(fmt.Sprintf("%s must be a string value", key))
		} else {
			return val, nil
		}
	}
	return "", invalidParametersResponse("share, username and password must be provided")
}

func (s smbServiceBroker) cleanupResourcesCreatedByProvision(ctx context.Context, instanceID string) {
	e := s.PersistentVolumeClaim.Delete(ctx, instanceID, metav1.DeleteOptions{})
	if e != nil {
		s.Logger.Error("handlers.cleanupResourcesCreatedByProvision-failed", e)
	}
	e = s.PersistentVolume.Delete(ctx, instanceID, metav1.DeleteOptions{})
	if e != nil {
		s.Logger.Error("handlers.cleanupResourcesCreatedByProvision-failed", e)
	}
	e = s.Secret.Delete(ctx, instanceID, metav1.DeleteOptions{})
	if e != nil {
		s.Logger.Error("handlers.cleanupResourcesCreatedByProvision-failed", e)
	}
}

