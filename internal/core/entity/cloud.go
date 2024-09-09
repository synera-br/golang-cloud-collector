package entity

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type AzureProviderInterface interface {
	GetSubscription(ctx context.Context, name string, id string) (*armsubscriptions.Subscription, error)
	ListResourcesByTag(ctx context.Context, tagKey, tagValue string) ([]*armresources.GenericResourceExpanded, error)
	ListResourcesByResourceGroup(ctx context.Context, rsg string) ([]*armresources.GenericResourceExpanded, error)
	ListResources(ctx context.Context) ([]*armresources.GenericResourceExpanded, error)
	FilterResources(ctx context.Context, name ...string) ([]*armresources.ResourceGroup, error)
}

type AzureProvider struct {
	Subscription      string `json:"subsription" binding:"required" mapstructure:"subscription_id"`
	ApplicationID     string `json:"application_id" binding:"required" mapstructure:"application_id"`
	ApplicationSecret string `json:"application_sceret" binding:"required" mapstructure:"application_secret"`
	Tenant            string `json:"tenant" binding:"required" mapstructure:"tenant_id"`
}

type AzureSubscription struct {
	Name                string `json:"name" binding:"required"`
	ID                  string `json:"id" binding:"required"`
	TenantID            string `json:"tenant_id" binding:"required"`
	AuthorizationSource string `json:"authorization_source" binding:"required"`
	State               string `json:"state" binding:"required"`
	Type                string `json:"type" binding:"required"`
}

func (a *AzureProvider) Validate() error {

	var err error = nil
	if a.Subscription == "" {
		err = errors.New("the Azure subscription cannot be empty")
	}

	if a.ApplicationID == "" {
		err = errors.New("the Azure application ID cannot be empty")
	}

	if a.ApplicationSecret == "" {
		err = errors.New("the Azure application secret cannot be empty")
	}

	if a.Tenant == "" {
		err = errors.New("the Azure tenant cannot be empty")
	}

	return err
}

func (a *AzureProvider) IsEmpty() bool {
	return a.Subscription == "" &&
		a.ApplicationID == "" &&
		a.ApplicationSecret == "" &&
		a.Tenant == ""
}
