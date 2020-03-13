package azurerm

import (
	"context"
	"fmt"

	azureResourcesAPI "github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/go-azure-helpers/sender"
)

//go:generate go run ./cmd

// AzureReader is the middleware between TC and AzureRM
type AzureReader struct {
	config     authentication.Config
	authorizer autorest.Authorizer

	resourceGroup azureResourcesAPI.Group
}

// NewAzureReader returns a AzureReader
func NewAzureReader(ctx context.Context, clientID, clientSecret, environment, resourceGroupName, subscriptionID, tenantID string) (*AzureReader, error) {
	// Config
	cfgBuilder := &authentication.Builder{
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		Environment:    environment,
		SubscriptionID: subscriptionID,
		TenantID:       tenantID,

		SupportsClientSecretAuth: true,
	}

	cfg, err := cfgBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("could not build 'azure/authentication.Config' because: %s", err)
	}

	// Authorizer
	env, err := authentication.DetermineEnvironment(cfg.Environment)
	if err != nil {
		return nil, fmt.Errorf("could not initialize 'azure.Environment.' because: %s", err)
	}

	oauthConfig, err := cfg.BuildOAuthConfig(env.ActiveDirectoryEndpoint)
	if err != nil {
		return nil, fmt.Errorf("could not initialize 'azure/authentication.OAuthConfig.' because: %s", err)
	}
	// OAuthConfigForTenant returns a pointer, which can be nil.
	if oauthConfig == nil {
		return nil, fmt.Errorf("could not configure OAuthConfig for tenant %s", cfg.TenantID)
	}

	azureSender := sender.BuildSender("AzureRM")

	auth, err := cfg.GetAuthorizationToken(azureSender, oauthConfig, env.ResourceManagerEndpoint)
	if err != nil {
		return nil, fmt.Errorf("could not initialize 'azure/autorest.Authorizer.' because: %s", err)
	}

	// Resource Group
	client := azureResourcesAPI.NewGroupsClient(cfg.SubscriptionID)
	client.Authorizer = auth
	resourceGroup, err := client.Get(ctx, resourceGroupName)
	if err != nil {
		return nil, fmt.Errorf("could not 'azure/resources.GroupsClient.Get' the resource group because: %s", err)
	}

	return &AzureReader{
		config:        *cfg,
		authorizer:    auth,
		resourceGroup: resourceGroup,
	}, nil
}

// GetResourceGroupName returns the current Resource Group name
func (ar *AzureReader) GetResourceGroupName() string {
	return *ar.resourceGroup.Name
}

// GetLocation returns the current Resource Group location
func (ar *AzureReader) GetLocation() string {
	return *ar.resourceGroup.Location
}
