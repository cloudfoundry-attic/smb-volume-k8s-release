package handlers

import (
	"bytes"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/smb-broker/store"
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain"
	"github.com/pivotal-cf/brokerapi/domain/apiresponses"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"net/http"
	"os"
)

const MountConfigKey = "name"
const MountBindOptionKey = "mount"
const DefaultMountPath = "/home/vcap/data/"
const ServiceID = "123"
const PlanID = "plan-id"

func BrokerHandler(serviceInstanceStore store.ServiceInstanceStore, pv corev1.PersistentVolumeInterface, pvc corev1.PersistentVolumeClaimInterface) (http.Handler, error) {
	if serviceInstanceStore == nil {
		return nil, errors.New("missing a Service Instance Store")
	}
	router := mux.NewRouter()
	logger := lager.NewLogger("smb-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	brokerapi.AttachRoutes(router, smbServiceBroker{
		Store:                 serviceInstanceStore,
		PersistentVolume:      pv,
		PersistentVolumeClaim: pvc,
	}, logger)
	return router, nil
}

type smbServiceBroker struct {
	Store                 store.ServiceInstanceStore
	PersistentVolume      corev1.PersistentVolumeInterface
	PersistentVolumeClaim corev1.PersistentVolumeClaimInterface
}

func (s smbServiceBroker) Services(ctx context.Context) ([]domain.Service, error) {
	return []domain.Service{{
		ID:                   ServiceID,
		Name:                 "SMB",
		Description:          "SMB for K8s",
		Bindable:             true,
		InstancesRetrievable: true,
		BindingsRetrievable:  true,
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

func (s smbServiceBroker) Provision(ctx context.Context, instanceID string, details domain.ProvisionDetails, asyncAllowed bool) (domain.ProvisionedServiceSpec, error) {
	var serviceInstanceParameters map[string]interface{}

	if details.RawParameters != nil {
		var decoder = json.NewDecoder(bytes.NewBuffer(details.RawParameters))
		err := decoder.Decode(&serviceInstanceParameters)
		if err != nil {
			return domain.ProvisionedServiceSpec{}, errors.New("unable to decode service instance parameters")
		}
	}

	err := s.Store.Add(instanceID, store.ServiceInstance{
		ServiceID:  details.ServiceID,
		PlanID:     details.PlanID,
		Parameters: serviceInstanceParameters,
	})
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	_, err = s.PersistentVolumeClaim.Create(&v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: instanceID,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			VolumeName:  instanceID,
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse("1M")},
			},
		},
	})
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	_, err = s.PersistentVolume.Create(&v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: instanceID,
		},
		Spec: v1.PersistentVolumeSpec{
			StorageClassName: "standard",
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			Capacity:         v1.ResourceList{v1.ResourceStorage: resource.MustParse("100M")},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/tmp/",
				},
			},
		},
	})
	return domain.ProvisionedServiceSpec{}, err
}

func (s smbServiceBroker) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, asyncAllowed bool) (domain.DeprovisionServiceSpec, error) {
	err := s.PersistentVolumeClaim.Delete(instanceID, &metav1.DeleteOptions{})
	if err != nil {
		return domain.DeprovisionServiceSpec{}, err
	}

	err = s.PersistentVolume.Delete(instanceID, &metav1.DeleteOptions{})
	if err != nil {
		return domain.DeprovisionServiceSpec{}, err
	}

	s.Store.Remove(instanceID)

	return domain.DeprovisionServiceSpec{}, nil
}

func (s smbServiceBroker) GetInstance(ctx context.Context, instanceID string) (domain.GetInstanceDetailsSpec, error) {
	retrievedServiceInstance, found := s.Store.Get(instanceID)
	if !found {
		return domain.GetInstanceDetailsSpec{}, apiresponses.NewFailureResponse(errors.New("unable to find service instance"), 404, "")
	}

	parametersInstanceDetailsMap := map[string]interface{}{}
	for key, val := range retrievedServiceInstance.Parameters {
		parametersInstanceDetailsMap[key] = val
	}

	return domain.GetInstanceDetailsSpec{
		ServiceID:  retrievedServiceInstance.ServiceID,
		PlanID:     retrievedServiceInstance.PlanID,
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

	_, found := s.Store.Get(instanceID)
	if !found {
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
	_, found := s.Store.Get(instanceID)
	if !found {
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
