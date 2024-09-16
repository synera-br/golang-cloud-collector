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
	"github.com/synera-br/golang-cloud-collector/pkg/otelpkg"
)

type AzureServiceInterface interface {
	entity.AzureProviderInterface
}

type AzureService struct {
	Repository entity.AzureProviderInterface
	Cache      cache.CacheInterface
	Tracer     *otelpkg.OtelPkgInstrument
}

const azurePrefix = "azure"

func NewAzureService(provider *entity.AzureProviderInterface, cc *cache.CacheInterface, otl *otelpkg.OtelPkgInstrument) (AzureServiceInterface, error) {

	return &AzureService{
		Repository: *provider,
		Cache:      *cc,
		Tracer:     otl,
	}, nil

}

func (s *AzureService) GetSubscription(ctx context.Context, name string, id string) (*armsubscriptions.Subscription, error) {
	ctxSpan, span := s.Tracer.Tracer.Start(ctx, "AzureService.GetSubscription")
	defer span.End()

	var data armsubscriptions.Subscription
	result, _ := s.Cache.Get(ctxSpan, fmt.Sprintf("%s_%s", azurePrefix, name))
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		go s.getSubscriptionFromRepository(ctxSpan, name, id)
		return &data, nil
	}

	return s.getSubscriptionFromRepository(ctxSpan, name, id)
}

func (s *AzureService) getSubscriptionFromRepository(ctx context.Context, name string, id string) (*armsubscriptions.Subscription, error) {
	ctxSpan, span := s.Tracer.Tracer.Start(ctx, "AzureService.getSubscriptionFromRepository")
	defer span.End()

	v, err := s.Repository.GetSubscription(ctxSpan, name, id)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	serializedData, err := json.Marshal(v)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if v != nil {
		s.Cache.Set(ctxSpan, fmt.Sprintf("%s_%s", azurePrefix, name), serializedData, s.Cache.TTL(time.Second))
	}

	return v, nil
}

func (s *AzureService) ListResourcesByTag(ctx context.Context, tagKey, tagValue string) ([]*armresources.GenericResourceExpanded, error) {
	ctxSpan, span := s.Tracer.Tracer.Start(ctx, "AzureService.ListResourcesByTag")
	defer span.End()

	var data []*armresources.GenericResourceExpanded

	queryPrefix := fmt.Sprintf("%s_key_%s_value_%s", azurePrefix, tagKey, tagValue)
	result, _ := s.Cache.Get(ctxSpan, queryPrefix)
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		go s.listResourcesByTagFromRepository(ctxSpan, tagKey, tagValue)
		return data, nil
	}
	return s.listResourcesByTagFromRepository(ctxSpan, tagKey, tagValue)
}

func (s *AzureService) listResourcesByTagFromRepository(ctx context.Context, tagKey, tagValue string) ([]*armresources.GenericResourceExpanded, error) {
	ctxSpan, span := s.Tracer.Tracer.Start(ctx, "AzureService.listResourcesByTagFromRepository")
	defer span.End()

	v, err := s.Repository.ListResourcesByTag(ctxSpan, tagKey, tagValue)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	serializedData, err := json.Marshal(v)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if len(v) > 0 {
		s.Cache.Set(ctxSpan, fmt.Sprintf("%s_key_%s_value_%s", azurePrefix, tagKey, tagValue), serializedData, s.Cache.TTL(time.Second))
	}
	return v, nil

}

func (s *AzureService) ListResourcesByResourceGroup(ctx context.Context, name string) ([]*armresources.GenericResourceExpanded, error) {
	ctxSpan, span := s.Tracer.Tracer.Start(ctx, "AzureService.ListResourcesByResourceGroup")
	defer span.End()

	var data []*armresources.GenericResourceExpanded
	result, _ := s.Cache.Get(ctxSpan, fmt.Sprintf("%s_%s", azurePrefix, name))
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		go s.listResourcesByResourceGroupFromRepository(ctxSpan, name)
		return data, nil
	}
	return s.listResourcesByResourceGroupFromRepository(ctxSpan, name)
}

func (s *AzureService) listResourcesByResourceGroupFromRepository(ctx context.Context, name string) ([]*armresources.GenericResourceExpanded, error) {
	ctxSpan, span := s.Tracer.Tracer.Start(ctx, "AzureService.listResourcesByResourceGroupFromRepository")
	defer span.End()

	v, err := s.Repository.ListResourcesByResourceGroup(ctxSpan, name)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	serializedData, err := json.Marshal(v)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if len(v) > 0 {
		s.Cache.Set(ctxSpan, fmt.Sprintf("%s_%s", azurePrefix, name), serializedData, s.Cache.TTL(time.Second))
	}

	return v, nil
}

func (s *AzureService) ListResources(ctx context.Context) ([]*armresources.GenericResourceExpanded, error) {
	ctxSpan, span := s.Tracer.Tracer.Start(ctx, "AzureService.ListResources")
	defer span.End()

	var data []*armresources.GenericResourceExpanded
	result, _ := s.Cache.Get(ctxSpan, fmt.Sprintf("%s_all_resources", azurePrefix))
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		go s.listResourcesFromRepository(ctxSpan)
		return data, nil
	}
	return s.listResourcesFromRepository(ctxSpan)
}

func (s *AzureService) listResourcesFromRepository(ctx context.Context) ([]*armresources.GenericResourceExpanded, error) {
	ctxSpan, span := s.Tracer.Tracer.Start(ctx, "AzureService.listResourcesFromRepository")
	defer span.End()

	v, err := s.Repository.ListResources(ctxSpan)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	serializedData, err := json.Marshal(v)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if len(v) > 0 {
		s.Cache.Set(ctxSpan, fmt.Sprintf("%s_all_resources", azurePrefix), serializedData, s.Cache.TTL(time.Second))
	}

	return v, nil
}

func (s *AzureService) FilterResourcesByResourceGroup(ctx context.Context, name string) ([]*armresources.GenericResourceExpanded, error) {
	ctxSpan, span := s.Tracer.Tracer.Start(ctx, "AzureService.FilterResourcesByResourceGroup")
	defer span.End()

	var data []*armresources.GenericResourceExpanded
	result, _ := s.Cache.Get(ctxSpan, fmt.Sprintf("%s_rsg_%s", azurePrefix, name))
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		go s.filterResourcesByResourceGroupFromRepository(ctxSpan, name)
		return data, nil
	}

	return s.filterResourcesByResourceGroupFromRepository(ctxSpan, name)
}

func (s *AzureService) filterResourcesByResourceGroupFromRepository(ctx context.Context, name string) ([]*armresources.GenericResourceExpanded, error) {
	ctxSpan, span := s.Tracer.Tracer.Start(ctx, "AzureService.filterResourcesByResourceGroupFromRepository")
	defer span.End()

	v, err := s.Repository.ListResources(ctxSpan)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	serializedData, err := json.Marshal(v)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if len(v) > 0 {
		s.Cache.Set(ctxSpan, fmt.Sprintf("%s_rsg_%s", azurePrefix, name), serializedData, s.Cache.TTL(time.Second))
	}

	return v, nil
}

func (s *AzureService) FilterResources(ctx context.Context, name ...string) ([]*armresources.ResourceGroup, error) {
	ctxSpan, span := s.Tracer.Tracer.Start(ctx, "AzureService.FilterResources")
	defer span.End()

	var data []*armresources.ResourceGroup
	var result []byte
	if len(name) > 0 && name[0] != "" {
		result, _ = s.Cache.Get(ctxSpan, fmt.Sprintf("%s_azure_filter_rsg_%s", azurePrefix, name[0]))
	} else {
		result, _ = s.Cache.Get(ctxSpan, fmt.Sprintf("%s_filter_tags", azurePrefix))
	}

	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		go s.filterResourcesFromRepository(ctxSpan, name...)
		return data, nil
	}
	return s.filterResourcesFromRepository(ctxSpan, name...)

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
