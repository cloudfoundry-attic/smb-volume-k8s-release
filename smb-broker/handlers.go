package main

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/smb-broker/store"
	"context"
	"github.com/gorilla/mux"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-cf/brokerapi/domain"
	"net/http"
	"os"
)

const ServiceID = "123"
const PlanID = "plan-id"

func BrokerHandler(store ...store.ServiceInstanceStore) http.Handler {
	router := mux.NewRouter()
	logger := lager.NewLogger("smb-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))

	brokerapi.AttachRoutes(router, SMBServiceBroker{
		Store: store,
	}, logger)
	return router
}

type SMBServiceBroker struct {
	Store []store.ServiceInstanceStore
}

func (S SMBServiceBroker) Services(ctx context.Context) ([]domain.Service, error) {
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

func (S SMBServiceBroker) Provision(ctx context.Context, instanceID string, details domain.ProvisionDetails, asyncAllowed bool) (domain.ProvisionedServiceSpec, error) {
	return domain.ProvisionedServiceSpec{}, nil
}

func (S SMBServiceBroker) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, asyncAllowed bool) (domain.DeprovisionServiceSpec, error) {
	panic("implement me")
}

func (s SMBServiceBroker) GetInstance(ctx context.Context, instanceID string) (domain.GetInstanceDetailsSpec, error) {
	var get store.ServiceInstance

	if s.Store != nil {
		get = s.Store[0].Get(instanceID)
	}

	parametersInstanceDetailsMap := map[string]interface{}{}
	for key, val := range get.Parameters {
		parametersInstanceDetailsMap[key] = val
	}

	return domain.GetInstanceDetailsSpec{
		ServiceID:  get.ServiceID,
		PlanID:     get.PlanID,
		Parameters: parametersInstanceDetailsMap,
	}, nil
}

func (S SMBServiceBroker) Update(ctx context.Context, instanceID string, details domain.UpdateDetails, asyncAllowed bool) (domain.UpdateServiceSpec, error) {
	panic("implement me")
}

func (S SMBServiceBroker) LastOperation(ctx context.Context, instanceID string, details domain.PollDetails) (domain.LastOperation, error) {
	panic("implement me")
}

func (S SMBServiceBroker) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, asyncAllowed bool) (domain.Binding, error) {
	panic("implement me")
}

func (S SMBServiceBroker) Unbind(ctx context.Context, instanceID, bindingID string, details domain.UnbindDetails, asyncAllowed bool) (domain.UnbindSpec, error) {
	panic("implement me")
}

func (S SMBServiceBroker) GetBinding(ctx context.Context, instanceID, bindingID string) (domain.GetBindingSpec, error) {
	panic("implement me")
}

func (S SMBServiceBroker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details domain.PollDetails) (domain.LastOperation, error) {
	panic("implement me")
}
