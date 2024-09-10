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
)

type BackstageServiceInterface interface {
	entity.BackstageInterface
}

type BackstageService struct {
	Azure AzureServiceInterface
	Amqp  mq.AMQPServiceInterface
	Cache cache.CacheInterface
}

const backstagePrefix = "backstage"

func NewBackstageService(azure AzureServiceInterface, mq mq.AMQPServiceInterface, cache cache.CacheInterface) BackstageServiceInterface {

	return &BackstageService{
		Azure: azure,
		Amqp:  mq,
		Cache: cache,
	}
}

func (b *BackstageService) TriggerSyncProvider(ctx context.Context, trigger *entity.Trigger) ([]entity.KindReource, error) {
	var err error = nil
	if trigger.Provider == "azure" {
		return b.azureTriggerSyncProvider(ctx, trigger)

	} else if trigger.Provider == "aws" {

	} else {
		return nil, errors.New("provider not found")
	}
	return nil, err
}

func (b *BackstageService) azureTriggerSyncProvider(ctx context.Context, trigger *entity.Trigger) ([]entity.KindReource, error) {
	var err error = nil
	response := []entity.KindReource{}

	if trigger.Account != "" {
		rsg, err := b.Azure.ListResourcesByResourceGroup(ctx, trigger.Account)
		if err != nil {
			return response, err
		}

		// var sub *entity.AzureSubscription

		for _, r := range rsg {
			if r.ID != nil {
				parseID := b.parseResourceID(*r.ID)
				resultRsg, err := b.Azure.FilterResources(ctx, parseID["resourcegroups"])
				if err != nil {
					return response, err
				}

				resultSubs, err := b.Azure.GetSubscription(ctx, "", "ea9f2737-3006-4d2f-b375-177c70866743")
				if err != nil {
					return response, err
				}

				dependsSubs := b.parseToTemplate(resultSubs, "subscriptions")
				if !b.contains(response, *dependsSubs) {
					response = append(response, *dependsSubs)
				}
				dependsGroup := b.parseToTemplate(resultRsg[0], "resourcegroups")
				dependsGroup.Spec.DependsOn = append(dependsGroup.Spec.DependsOn, fmt.Sprintf("resource:%s", dependsSubs.Metadata.Name))
				if !b.contains(response, *dependsGroup) {
					response = append(response, *dependsGroup)
				}
				resource := b.parseToTemplate(r, "resources")
				if resource != nil {
					resource.Spec.DependsOn = append(resource.Spec.DependsOn, fmt.Sprintf("resource:%s", dependsGroup.Metadata.Name))
					if !b.contains(response, *resource) {
						response = append(response, *resource)
					}
				}
			}
		}

	} else if trigger.Tag.Key != "" && trigger.Tag.Value != "" {
		_, err = b.Azure.ListResourcesByTag(ctx, trigger.Tag.Key, trigger.Tag.Value)
		if err != nil {
			return nil, err
		}

	} else {
		rsg, err := b.Azure.ListResources(ctx)
		if err != nil {
			return response, err
		}

		// var sub *entity.AzureSubscription
		for _, r := range rsg {
			if r.ID != nil {
				parseID := b.parseResourceID(*r.ID)
				resultRsg, err := b.Azure.FilterResources(ctx, parseID["resourcegroups"])
				if err != nil {
					return response, err
				}

				resultSubs, err := b.Azure.GetSubscription(ctx, "", "ea9f2737-3006-4d2f-b375-177c70866743")
				if err != nil {
					return response, err
				}

				dependsSubs := b.parseToTemplate(resultSubs, "subscriptions")
				if !b.contains(response, *dependsSubs) {
					response = append(response, *dependsSubs)
				}
				dependsGroup := b.parseToTemplate(resultRsg[0], "resourcegroups")
				dependsGroup.Spec.DependsOn = append(dependsGroup.Spec.DependsOn, fmt.Sprintf("resource:%s", dependsSubs.Metadata.Name))
				if !b.contains(response, *dependsGroup) {
					response = append(response, *dependsGroup)
				}
				resource := b.parseToTemplate(r, "resources")
				if resource != nil {
					resource.Spec.DependsOn = append(resource.Spec.DependsOn, fmt.Sprintf("resource:%s", dependsGroup.Metadata.Name))
					if !b.contains(response, *resource) {
						response = append(response, *resource)
					}
				}
			}
		}
	}

	b.publishResourcesToAMQP(ctx, response)

	return response, err
}

func (b *BackstageService) parseResourceID(id string) map[string]string {

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

func (b *BackstageService) parseToTemplate(resource interface{}, resourceType string) *entity.KindReource {
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

func (b *BackstageService) parseRelationship(ctx context.Context, response []entity.KindReource, rsg []*armresources.GenericResourceExpanded) ([]entity.KindReource, error) {

	for _, r := range rsg {
		if r.ID != nil {
			parseID := b.parseResourceID(*r.ID)
			resultRsg, err := b.Azure.FilterResources(ctx, parseID["resourcegroups"])
			if err != nil {
				return response, err
			}

			resultSubs, err := b.Azure.GetSubscription(ctx, "", "ea9f2737-3006-4d2f-b375-177c70866743")
			if err != nil {
				return response, err
			}

			dependsSubs := b.parseToTemplate(resultSubs, "subscriptions")
			if !b.contains(response, *dependsSubs) {
				response = append(response, *dependsSubs)
			}
			dependsGroup := b.parseToTemplate(resultRsg[0], "resourcegroups")
			dependsGroup.Spec.DependsOn = append(dependsGroup.Spec.DependsOn, fmt.Sprintf("resource:%s", dependsSubs.Metadata.Name))
			if !b.contains(response, *dependsGroup) {
				response = append(response, *dependsGroup)
			}
			resource := b.parseToTemplate(r, "resources")
			if resource != nil {
				resource.Spec.DependsOn = append(resource.Spec.DependsOn, fmt.Sprintf("resource:%s", dependsGroup.Metadata.Name))
				if !b.contains(response, *resource) {
					response = append(response, *resource)
				}
			}
		}
	}
	return response, nil
}

func (b *BackstageService) contains(slice []entity.KindReource, item entity.KindReource) bool {
	for _, v := range slice {
		if (v.Metadata.Name == item.Metadata.Name) && (v.Metadata.Namespace == item.Metadata.Namespace) && (v.Spec.Type == item.Spec.Type) {
			return true
		}
	}
	return false
}

func (b *BackstageService) publishResourcesToAMQP(ctx context.Context, data interface{}) error {

	dataConvertToByte, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = b.Amqp.Publish(ctx, mq.DataAMQP{
		ContentType: "application/json",
		Exchange:    "collector",
		RouteKey:    "backstage",
		Queue:       "manifests",
		Body:        dataConvertToByte,
	})

	return err
}

func (b *BackstageService) GetAllKinds(ctx context.Context, search entity.FilterKind) ([]entity.KindReource, error) {
	var data []entity.KindReource

	queryPrefix := fmt.Sprintf("%s_get_all_kinds%s%s%s", backstagePrefix, search.Namespace, search.Kind, search.Name)

	result, _ := b.Cache.Get(ctx, queryPrefix)
	if result != nil {
		err := json.Unmarshal(result, &data)
		if err != nil {
			return nil, err
		}
		go b.TriggerSyncProvider(ctx, &entity.Trigger{
			Provider: "azure",
		})

		return data, nil
	}

	objs, err := b.TriggerSyncProvider(ctx, &entity.Trigger{
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

	filter, err := b.filterKinds(objs, search)
	serializedData, err := json.Marshal(filter)
	go b.Cache.Set(ctx, queryPrefix, serializedData, b.Cache.TTL(time.Second))
	return filter, err

}

func (b *BackstageService) filterKinds(request []entity.KindReource, filter entity.FilterKind) ([]entity.KindReource, error) {

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
