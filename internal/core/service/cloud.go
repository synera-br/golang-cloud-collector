package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/synera-br/golang-cloud-collector/internal/core/entity"
	"github.com/synera-br/golang-cloud-collector/pkg/cache"
)

type AzureServiceInterface interface {
	entity.AzureProviderInterface
}

type AzureService struct {
	Repository entity.AzureProviderInterface
	Cache      cache.CacheInterface
}

const azurePrefix = "azure"

func NewAzureService(provider *entity.AzureProviderInterface, cc *cache.CacheInterface) (AzureServiceInterface, error) {

	return &AzureService{
		Repository: *provider,
		Cache:      *cc,
	}, nil

}

func (s *AzureService) GetSubscription(ctx context.Context, name string, id string) (*armsubscriptions.Subscription, error) {
	var data armsubscriptions.Subscription
	result, _ := s.Cache.Get(ctx, fmt.Sprintf("%s_%s", azurePrefix, name))
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			return nil, err
		}
		go s.getSubscriptionFromRepository(ctx, name, id)
		return &data, nil
	}

	return s.getSubscriptionFromRepository(ctx, name, id)
}

func (s *AzureService) getSubscriptionFromRepository(ctx context.Context, name string, id string) (*armsubscriptions.Subscription, error) {

	v, err := s.Repository.GetSubscription(ctx, name, id)
	if err != nil {
		return nil, err
	}

	serializedData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	if v != nil {
		s.Cache.Set(ctx, fmt.Sprintf("%s_%s", azurePrefix, name), serializedData, s.Cache.TTL(time.Second))
	}

	return v, nil
}

func (s *AzureService) ListResourcesByTag(ctx context.Context, tagKey, tagValue string) ([]*armresources.GenericResourceExpanded, error) {
	var data []*armresources.GenericResourceExpanded

	queryPrefix := fmt.Sprintf("%s_key_%s_value_%s", azurePrefix, tagKey, tagValue)
	result, _ := s.Cache.Get(ctx, queryPrefix)
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			return nil, err
		}
		go s.listResourcesByTagFromRepository(context.Background(), tagKey, tagValue)
		return data, nil
	}
	return s.listResourcesByTagFromRepository(context.Background(), tagKey, tagValue)
}

func (s *AzureService) listResourcesByTagFromRepository(ctx context.Context, tagKey, tagValue string) ([]*armresources.GenericResourceExpanded, error) {

	v, err := s.Repository.ListResourcesByTag(context.Background(), tagKey, tagValue)
	if err != nil {
		return nil, err
	}

	serializedData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	if len(v) > 0 {
		s.Cache.Set(ctx, fmt.Sprintf("%s_key_%s_value_%s", azurePrefix, tagKey, tagValue), serializedData, s.Cache.TTL(time.Second))
	}
	return v, nil

}

func (s *AzureService) ListResourcesByResourceGroup(ctx context.Context, name string) ([]*armresources.GenericResourceExpanded, error) {
	var data []*armresources.GenericResourceExpanded
	result, _ := s.Cache.Get(ctx, fmt.Sprintf("%s_%s", azurePrefix, name))
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			return nil, err
		}

		go s.listResourcesByResourceGroupFromRepository(ctx, name)
		return data, nil
	}
	return s.listResourcesByResourceGroupFromRepository(ctx, name)
}

func (s *AzureService) listResourcesByResourceGroupFromRepository(ctx context.Context, name string) ([]*armresources.GenericResourceExpanded, error) {
	v, err := s.Repository.ListResourcesByResourceGroup(context.Background(), name)
	if err != nil {
		return nil, err
	}

	serializedData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	if len(v) > 0 {
		s.Cache.Set(ctx, fmt.Sprintf("%s_%s", azurePrefix, name), serializedData, s.Cache.TTL(time.Second))
	}

	return v, nil
}

func (s *AzureService) ListResources(ctx context.Context) ([]*armresources.GenericResourceExpanded, error) {
	var data []*armresources.GenericResourceExpanded
	result, _ := s.Cache.Get(ctx, fmt.Sprintf("%s_all_resources", azurePrefix))
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			return nil, err
		}
		go s.listResourcesFromRepository(ctx)
		return data, nil
	}
	return s.listResourcesFromRepository(ctx)
}

func (s *AzureService) listResourcesFromRepository(ctx context.Context) ([]*armresources.GenericResourceExpanded, error) {

	v, err := s.Repository.ListResources(context.Background())
	if err != nil {
		return nil, err
	}
	serializedData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	if len(v) > 0 {
		s.Cache.Set(ctx, fmt.Sprintf("%s_all_resources", azurePrefix), serializedData, s.Cache.TTL(time.Second))
	}

	return v, nil
}

func (s *AzureService) FilterResourcesByResourceGroup(ctx context.Context, name string) ([]*armresources.GenericResourceExpanded, error) {
	var data []*armresources.GenericResourceExpanded
	result, _ := s.Cache.Get(ctx, fmt.Sprintf("%s_rsg_%s", azurePrefix, name))
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			return nil, err
		}
		go s.filterResourcesByResourceGroupFromRepository(ctx, name)
		return data, nil
	}

	return s.filterResourcesByResourceGroupFromRepository(ctx, name)
}

func (s *AzureService) filterResourcesByResourceGroupFromRepository(ctx context.Context, name string) ([]*armresources.GenericResourceExpanded, error) {
	v, err := s.Repository.ListResources(context.Background())
	if err != nil {
		return nil, err
	}
	serializedData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	if len(v) > 0 {
		s.Cache.Set(ctx, fmt.Sprintf("%s_rsg_%s", azurePrefix, name), serializedData, s.Cache.TTL(time.Second))
	}

	return v, nil
}

func (s *AzureService) FilterResources(ctx context.Context, name ...string) ([]*armresources.ResourceGroup, error) {
	var data []*armresources.ResourceGroup
	var result []byte
	if len(name) > 0 && name[0] != "" {
		result, _ = s.Cache.Get(ctx, fmt.Sprintf("%s_azure_filter_rsg_%s", azurePrefix, name[0]))
	} else {
		result, _ = s.Cache.Get(ctx, fmt.Sprintf("%s_filter_tags", azurePrefix))
	}

	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			return nil, err
		}
		go s.filterResourcesFromRepository(ctx, name...)
		return data, nil
	}
	return s.filterResourcesFromRepository(ctx, name...)

}

func (s *AzureService) filterResourcesFromRepository(ctx context.Context, name ...string) ([]*armresources.ResourceGroup, error) {

	v, err := s.Repository.FilterResources(context.Background(), name...)
	if err != nil {
		return nil, err
	}
	serializedData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(v) > 0 {
		if len(name) > 0 && name[0] != "" {
			s.Cache.Set(ctx, fmt.Sprintf("%s_azure_filter_rsg_%s", azurePrefix, name[0]), serializedData, s.Cache.TTL(time.Second))
		} else {
			s.Cache.Set(ctx, fmt.Sprintf("%s_filter_tags", azurePrefix), serializedData, s.Cache.TTL(time.Second))
		}
	}

	return v, nil

}
