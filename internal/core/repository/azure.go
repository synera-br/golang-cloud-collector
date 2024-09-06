package repository

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/synera-br/golang-cloud-collector/internal/core/entity"
)

type AzureRepository struct {
	Provider entity.AzureProviderInterface
}

func NewAzureRepository(provider *entity.AzureProvider) entity.AzureProviderInterface {
	return &AzureRepository{}
}

func (a AzureRepository) Connection(provider *entity.AzureProvider) error {

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		// TODO: handle error
	}
	// Azure SDK Resource Management clients accept the credential as a parameter.
	// The client will authenticate with the credential as necessary.
	client, err := armsubscription.NewSubscriptionsClient(cred, nil)
	if err != nil {
		// TODO: handle error
	}
	_, err = client.Get(context.TODO(), subscriptionID, nil)
	if err != nil {
		// TODO: handle error
	}

	return nil
}
