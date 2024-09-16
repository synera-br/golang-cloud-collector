package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/synera-br/golang-cloud-collector/internal/core/entity"
	"github.com/synera-br/golang-cloud-collector/pkg/cache"
	"github.com/synera-br/golang-cloud-collector/pkg/mq"
	"github.com/synera-br/golang-cloud-collector/pkg/otelpkg"
)

type BackstageServiceInterface interface {
	entity.BackstageInterface
}

type BackstageService struct {
	Azure  AzureServiceInterface
	Amqp   mq.AMQPServiceInterface
	Cache  cache.CacheInterface
	Tracer *otelpkg.OtelPkgInstrument
}

const backstagePrefix = "backstage"

func NewBackstageService(azure AzureServiceInterface, mq mq.AMQPServiceInterface, cache cache.CacheInterface, otl *otelpkg.OtelPkgInstrument) BackstageServiceInterface {

	return &BackstageService{
		Azure:  azure,
		Amqp:   mq,
		Cache:  cache,
		Tracer: otl,
	}
}

func (b *BackstageService) TriggerSyncProvider(ctx context.Context, trigger *entity.Trigger) ([]entity.KindReource, error) {
	ctxSpan, span := b.Tracer.Tracer.Start(ctx, "BackstageService.TriggerSyncProvider")
	defer span.End()

	var err error = nil
	if trigger.Provider == "azure" {
		return b.azureTriggerSyncProvider(ctxSpan, trigger)

	} else if trigger.Provider == "aws" {

	} else {
		span.RecordError(errors.New("provider not found"))
		return nil, errors.New("provider not found")
	}
	return nil, err
}

func (b *BackstageService) TriggerSyncProviderTemp(ctx context.Context, trigger *entity.Trigger) ([]entity.KindReource, error) {
	ctxSpan, span := b.Tracer.Tracer.Start(ctx, "BackstageService.TriggerSyncProvider")
	defer span.End()

	var err error = nil
	var getItems []*armresources.GenericResourceExpanded
	response := []entity.KindReource{}

	parseItems := make(map[string][]*armresources.GenericResourceExpanded)

	//
	if trigger.TargetResource.ResourceName != "" && trigger.TargetResource.ResourceType != "" {

	} else if trigger.TargetTags.Key != "" && trigger.TargetTags.Value != "" {

	} else {
		getItems, err = b.Azure.ListResources(ctxSpan)
		if err != nil {
			return response, err
		}
	}

	for _, item := range getItems {
		if item.Tags != nil && item.Tags["owner"] != nil {
			tag := *item.Tags["owner"]
			parseItems[tag] = append(parseItems[tag], item)
		}
	}

	return nil, err
}

func (b *BackstageService) azureTriggerSyncProvider(ctx context.Context, trigger *entity.Trigger) ([]entity.KindReource, error) {
	ctxSpan, span := b.Tracer.Tracer.Start(ctx, "BackstageService.azureTriggerSyncProvider")
	defer span.End()

	var err error = nil
	var resources []*armresources.GenericResourceExpanded

	if trigger.TargetResource.ResourceName != "" {
		resources, err = b.Azure.ListResourcesByResourceGroup(ctxSpan, trigger.TargetResource.ResourceName)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
	} else if trigger.TargetTags.Key != "" && trigger.TargetTags.Value != "" {
		resources, err = b.Azure.ListResourcesByTag(ctxSpan, trigger.TargetTags.Key, trigger.TargetTags.Value)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
	} else {
		resources, err = b.Azure.ListResources(ctxSpan)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
	}

	response, err := b.parseRelationship(ctxSpan, resources)
	if err != nil {
		return nil, err
	}

	b.publishResourcesToAMQP(ctxSpan, response)

	return response, err
}

func (b *BackstageService) parseResourceID(ctx context.Context, id string) map[string]string {
	_, span := b.Tracer.Tracer.Start(ctx, "BackstageService.parseResourceID")
	defer span.End()

	result := make(map[string]string)
	if id == "" {
		return result
	}

	// Remove leading slash if needed
	if strings.HasPrefix(id, "/") {
		id = strings.TrimPrefix(id, "/")
	}

	// Split the string by '/'
	parts := strings.Split(id, "/")

	for i, part := range parts {
		result[strings.ToLower(fmt.Sprintf("%d", i))] = strings.ToLower(part)
	}

	return result
}

func (b *BackstageService) parseToTemplate(ctx context.Context, resource interface{}, resourceType string) *entity.KindReource {
	_, span := b.Tracer.Tracer.Start(ctx, "BackstageService.parseToTemplate")
	defer span.End()

	result := entity.KindReource{}
	result.Metadata.Labels = make(map[string]string)
	result.Metadata.Annotations = make(map[string]string)

	if resourceType == "subscriptions" {
		rsc, ok := resource.(*armsubscriptions.Subscription)
		if ok {
			result.Metadata.Name = *rsc.DisplayName
			result.Spec.Type = "subscription"
			result.Metadata.Annotations["subscription_state"] = string(*rsc.State)
			result.Metadata.Annotations["subscription_quota"] = string(*rsc.SubscriptionPolicies.QuotaID)
			result.Metadata.Annotations["subscription_limit"] = string(*rsc.SubscriptionPolicies.SpendingLimit)

			for k, v := range rsc.Tags {
				if v != nil {
					result.Metadata.Labels[k] = *v
				}
			}
			owner, exists := rsc.Tags["owner"]
			if !exists {
				return nil
			}
			result.Spec.Owner = *owner

			system, exists := rsc.Tags["system"]
			if !exists {
				return nil
			}
			result.Spec.System = *system
		}

	} else if resourceType == "resourcegroups" {

		rsg, isRSG := resource.(*armresources.ResourceGroup)

		if isRSG {
			result.Metadata.Name = *rsg.Name
			result.Spec.Type = strings.ToLower(strings.Split(*rsg.Type, "/")[1])
			result.Metadata.Annotations["resource_family"] = strings.ToLower(strings.Split(*rsg.Type, "/")[0])
			result.Metadata.Annotations["resource_type"] = strings.ToLower(strings.Split(*rsg.Type, "/")[1])
			result.Metadata.Annotations["resource_state"] = strings.ToLower(*rsg.Properties.ProvisioningState)

			for k, v := range rsg.Tags {
				if v != nil {
					result.Metadata.Labels[k] = *v
				}
			}
			owner, exists := rsg.Tags["owner"]
			if !exists {
				return nil
			}
			result.Spec.Owner = *owner

			system, exists := rsg.Tags["system"]
			if !exists {
				return nil
			}
			result.Spec.System = *system

		}
	} else {
		rsc, isRSC := resource.(*armresources.GenericResourceExpanded)
		if isRSC {
			result.Metadata.Name = *rsc.Name
			result.Spec.Type = strings.ToLower(strings.Split(*rsc.Type, "/")[1])
			result.Metadata.Annotations["resource_family"] = strings.ToLower(strings.Split(*rsc.Type, "/")[0])
			result.Metadata.Annotations["resource_type"] = strings.ToLower(strings.Split(*rsc.Type, "/")[1])
			if rsc.ProvisioningState != nil {
				result.Metadata.Annotations["resource_state"] = *rsc.ProvisioningState
			}

			for k, v := range rsc.Tags {
				if v != nil {
					result.Metadata.Labels[k] = *v
				}
			}
			owner, exists := rsc.Tags["owner"]
			if exists {
				// return nil
				result.Spec.Owner = *owner
			}

			system, exists := rsc.Tags["system"]
			if exists {
				// return nil
				result.Spec.System = *system
			}
		}
	}

	result.Validate()
	return &result
}

func (b *BackstageService) parseRelationship(ctx context.Context, resources []*armresources.GenericResourceExpanded) ([]entity.KindReource, error) {
	ctxSpan, span := b.Tracer.Tracer.Start(ctx, "BackstageService.azureTriggerSyncProvider")
	defer span.End()

	var response []entity.KindReource

	for _, r := range resources {
		if r.ID != nil {
			parseID := b.parseResourceID(ctxSpan, *r.ID)
			resultRsg, err := b.Azure.FilterResources(ctxSpan, parseID["resourcegroups"])
			if err != nil {
				span.RecordError(err)
				return nil, err
			}

			resultSubs, err := b.Azure.GetSubscription(ctxSpan, "", "")
			if err != nil {
				span.RecordError(err)
				return nil, err
			}

			dependsSubs := b.parseToTemplate(ctxSpan, resultSubs, "subscriptions")
			if !b.contains(ctxSpan, response, *dependsSubs) {
				response = append(response, *dependsSubs)
			}
			dependsGroup := b.parseToTemplate(ctxSpan, resultRsg[0], "resourcegroups")
			dependsGroup.Spec.DependsOn = append(dependsGroup.Spec.DependsOn, fmt.Sprintf("resource:%s", dependsSubs.Metadata.Name))
			if !b.contains(ctxSpan, response, *dependsGroup) {
				response = append(response, *dependsGroup)
			}
			resource := b.parseToTemplate(ctxSpan, r, "resources")
			if resource != nil {
				resource.Spec.DependsOn = append(resource.Spec.DependsOn, fmt.Sprintf("resource:%s", dependsGroup.Metadata.Name))
				if !b.contains(ctxSpan, response, *resource) {
					response = append(response, *resource)
				}
			}
		}
	}
	return response, nil
}

func (b *BackstageService) contains(ctx context.Context, slice []entity.KindReource, item entity.KindReource) bool {
	_, span := b.Tracer.Tracer.Start(ctx, "BackstageService.contains")
	defer span.End()

	for _, v := range slice {
		if (v.Metadata.Name == item.Metadata.Name) && (v.Metadata.Namespace == item.Metadata.Namespace) && (v.Spec.Type == item.Spec.Type) {
			return true
		}
	}
	return false
}

func (b *BackstageService) publishResourcesToAMQP(ctx context.Context, data interface{}) error {
	ctxSpan, span := b.Tracer.Tracer.Start(ctx, "BackstageService.contains")
	defer span.End()

	dataConvertToByte, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = b.Amqp.Publish(ctxSpan, mq.DataAMQP{
		ContentType: "application/json",
		Exchange:    "collector",
		RouteKey:    "backstage",
		Queue:       "manifests",
		Body:        dataConvertToByte,
	})

	return err
}

func (b *BackstageService) GetAllKinds(ctx context.Context, search entity.FilterKind) ([]entity.KindReource, error) {
	ctxSpan, span := b.Tracer.Tracer.Start(ctx, "BackstageService.GetAllKinds")
	defer span.End()

	var data []entity.KindReource

	queryPrefix := fmt.Sprintf("%s_get_all_kinds%s%s%s", backstagePrefix, search.Namespace, search.Kind, search.Name)

	result, _ := b.Cache.Get(ctxSpan, queryPrefix)
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			return nil, err
		}
		go b.TriggerSyncProvider(ctxSpan, &entity.Trigger{
			Provider: "azure",
		})

		return data, nil
	}

	objs, err := b.TriggerSyncProvider(ctxSpan, &entity.Trigger{
		Provider: "azure",
	})
	if err != nil {
		return nil, err
	}

	if search.Name == "" && search.Namespace == "" && search.Kind == "" {
		serializedData, err := json.Marshal(objs)
		go b.Cache.Set(ctx, queryPrefix, serializedData, b.Cache.TTL(time.Second))
		return objs, err
	}

	filter, _ := b.filterKinds(ctxSpan, objs, search)
	serializedData, err := json.Marshal(filter)
	go b.Cache.Set(ctx, queryPrefix, serializedData, b.Cache.TTL(time.Second))
	return filter, err
}

func (b *BackstageService) filterKinds(ctx context.Context, request []entity.KindReource, filter entity.FilterKind) ([]entity.KindReource, error) {
	_, span := b.Tracer.Tracer.Start(ctx, "BackstageService.filterKinds")
	defer span.End()

	var response []entity.KindReource
	var err error
	if filter.Name == "" && filter.Kind == "" && filter.Namespace == "" {
		return nil, errors.New("filter requires at least one non-empty field")
	}
	for _, req := range request {

		if (filter.Name == "" || strings.EqualFold(req.Metadata.Name, filter.Name)) &&
			(filter.Kind == "" || strings.EqualFold(req.Kind, filter.Kind)) &&
			(filter.Namespace == "" || strings.EqualFold(req.Metadata.Namespace, filter.Namespace)) {
			response = append(response, req)
		}

	}

	return response, err
}
