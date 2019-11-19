package main

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
	"net/http"
	"os"
)

const ServiceID = "123"
const PlanID = "plan-id"

func BrokerHandler(serviceInstanceStore store.ServiceInstanceStore) (http.Handler, error) {
	if serviceInstanceStore == nil {
		return nil, errors.New("missing a Service Instance Store")
	}
	router := mux.NewRouter()
	logger := lager.NewLogger("smb-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	brokerapi.AttachRoutes(router, smbServiceBroker{
		Store: serviceInstanceStore,
	}, logger)
	return router, nil
}

type smbServiceBroker struct {
	Store store.ServiceInstanceStore
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
		Requires:        []domain.RequiredPermission{},
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
		Parameters: serviceInstanceParameters,
	})
	return domain.ProvisionedServiceSpec{}, err
}

func (s smbServiceBroker) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, asyncAllowed bool) (domain.DeprovisionServiceSpec, error) {
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
	panic("implement me")
}

func (s smbServiceBroker) Unbind(ctx context.Context, instanceID, bindingID string, details domain.UnbindDetails, asyncAllowed bool) (domain.UnbindSpec, error) {
	panic("implement me")
}

func (s smbServiceBroker) GetBinding(ctx context.Context, instanceID, bindingID string) (domain.GetBindingSpec, error) {
	panic("implement me")
}

func (s smbServiceBroker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details domain.PollDetails) (domain.LastOperation, error) {
	panic("implement me")
}
