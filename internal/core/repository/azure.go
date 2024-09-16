package repository

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/synera-br/golang-cloud-collector/internal/core/entity"
	"github.com/synera-br/golang-cloud-collector/pkg/otelpkg"
)

type AzureRepository struct {
	Provider     entity.AzureProvider
	Tracer       *otelpkg.OtelPkgInstrument
	Client       *armresources.Client
	Subscription entity.AzureSubscription
	Credential   *azidentity.ClientSecretCredential
}

func NewAzureRepository(provider *entity.AzureProvider, otl *otelpkg.OtelPkgInstrument) (entity.AzureProviderInterface, error) {

	p := &AzureRepository{
		Provider: *provider,
		Tracer:   otl,
	}

	ctxSpan, span := p.Tracer.Tracer.Start(context.Background(), "AzureRepository.NewAzureRepository")
	defer span.End()

	if err := p.Connection(ctxSpan, &p.Provider); err != nil {
		return nil, err
	}
	return p, nil
}

func (a *AzureRepository) Connection(ctx context.Context, provider *entity.AzureProvider) error {

	ctxSpan, span := a.Tracer.Tracer.Start(context.Background(), "AzureRepository.Connection")
	defer span.End()

	var err error

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Create a custom HTTP client with the custom transport
	httpClient := &http.Client{
		Transport: transport,
	}

	// Set up the Azure identity client options to use the custom HTTP client
	credOptions := &azidentity.ClientSecretCredentialOptions{
		ClientOptions: policy.ClientOptions{
			Transport: httpClient,
		},
	}

	a.Credential, err = azidentity.NewClientSecretCredential(provider.Tenant, provider.ApplicationID, provider.ApplicationSecret, &azidentity.ClientSecretCredentialOptions{
		ClientOptions: credOptions.ClientOptions,
	})
	if err != nil {
		return fmt.Errorf("failed to create client secret credential: %w", err)
	}

	a.GetSubscription(ctxSpan, a.Provider.Subscription, "")

	client, err := armresources.NewClient(a.Provider.Subscription, a.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create client connection: %w", err)
	}
	// token, err := getAccessToken(cred)

	a.Client = client

	return nil
}

func (a *AzureRepository) ListResources(ctx context.Context) ([]*armresources.GenericResourceExpanded, error) {
	ctxSpan, span := a.Tracer.Tracer.Start(ctx, "AzureRepository.ListResources")
	defer span.End()

	pager := a.Client.NewListPager(&armresources.ClientListOptions{})

	var result []*armresources.GenericResourceExpanded = nil
	for pager.More() {
		resp, err := pager.NextPage(ctxSpan)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to list resources: %w", err)
		}
		result = resp.Value
	}

	return result, nil
}

func (a *AzureRepository) ListResourcesByTag(ctx context.Context, tagKey, tagValue string) ([]*armresources.GenericResourceExpanded, error) {
	ctxSpan, span := a.Tracer.Tracer.Start(ctx, "AzureRepository.ListResourcesByTag")
	defer span.End()

	filter := fmt.Sprintf("tagName eq '%s' and tagValue eq '%s'", tagKey, tagValue)

	pager := a.Client.NewListPager(&armresources.ClientListOptions{
		Filter: &filter,
	})

	var result []*armresources.GenericResourceExpanded = nil
	for pager.More() {
		resp, err := pager.NextPage(ctxSpan)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to list resources: %w", err)
		}
		result = resp.Value
	}

	return result, nil
}

func (a *AzureRepository) ListResourcesByResourceGroup(ctx context.Context, rsg string) ([]*armresources.GenericResourceExpanded, error) {
	ctxSpan, span := a.Tracer.Tracer.Start(ctx, "AzureRepository.ListResourcesByResourceGroup")
	defer span.End()

	pager := a.Client.NewListByResourceGroupPager(rsg, &armresources.ClientListByResourceGroupOptions{})

	var result []*armresources.GenericResourceExpanded = nil
	for pager.More() {
		resp, err := pager.NextPage(ctxSpan)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to list resources: %w", err)
		}
		result = resp.Value
	}

	return result, nil
}

// func (a *AzureRepository) ListResourcesByID(ctx context.Context, rsg string) ([]*armresources.GenericResourceExpanded, error) {

// 	pager := a.Client.GetByID(context.Background(), rsg)

// 	var result []*armresources.GenericResourceExpanded = nil
// 	for pager.More() {
// 		resp, err := pager.NextPage(ctx)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to list resources: %w", err)
// 		}
// 		result = resp.Value
// 	}

// 	return result, nil
// }

func getAccessToken(cred *azidentity.ClientSecretCredential) (string, error) {
	// const tenantInfoURL string = "https://graph.microsoft.com/v1.0/organization"
	// // Faz a requisição para a Microsoft Graph API
	// req, err := http.NewRequest("GET", tenantInfoURL, nil)
	// if err != nil {
	// 	log.Fatalf("falha ao criar requisição: %v", err)
	// }

	// req.Header.Add("Authorization", "Bearer "+token)
	// resp, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	log.Fatalf("falha ao fazer a requisição: %v", err)
	// }
	// defer resp.Body.Close()

	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatalf("falha ao ler a resposta: %v", err)
	// }
	// var result map[string]interface{}
	// if err := json.Unmarshal(body, &result); err != nil {
	// 	log.Fatalf("falha ao parsear JSON: %v", err)
	// }

	// Imprime o nome do tenant
	// organizations := result["value"].([]interface{})
	// for _, org := range organizations {
	// 	orgDetails := org.(map[string]interface{})
	// }
	// if err != nil {	// 	return err
	// }

	ctx := context.Background()
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return "", err
	}
	return token.Token, nil
}

func (a *AzureRepository) GetSubscription(ctx context.Context, name string, id string) (*armsubscriptions.Subscription, error) {
	ctxSpan, span := a.Tracer.Tracer.Start(ctx, "AzureRepository.GetSubscription")
	defer span.End()

	sub, err := armsubscriptions.NewClient(a.Credential, nil)
	if err != nil {
		return nil, err
	}

	pager := sub.NewListPager(nil)

	if id == "" {
		id = a.Provider.Subscription
	}

	for pager.More() {
		page, err := pager.NextPage(ctxSpan)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		for _, subscription := range page.SubscriptionListResult.Value {
			if strings.EqualFold(*subscription.DisplayName, name) {
				return subscription, nil
			}
			if *subscription.SubscriptionID == id {
				return subscription, nil
			}
		}

	}

	return nil, nil
}

func (a *AzureRepository) FilterResources(ctx context.Context, name ...string) ([]*armresources.ResourceGroup, error) {
	ctxSpan, span := a.Tracer.Tracer.Start(ctx, "AzureRepository.FilterResources")
	defer span.End()

	resourceGroupsClient, _ := armresources.NewResourceGroupsClient(a.Provider.Subscription, a.Credential, nil)

	pager := resourceGroupsClient.NewListPager(&armresources.ResourceGroupsClientListOptions{})

	var rsg []*armresources.ResourceGroup
	var err error
	for pager.More() {
		page, err := pager.NextPage(ctxSpan)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		if len(name) > 0 {
			for _, rg := range page.Value {
				if strings.Contains(*rg.Name, name[0]) {
					rsg = append(rsg, rg)
				}
			}
			return rsg, err
		}
		rsg = page.Value
	}
	return rsg, err
}
