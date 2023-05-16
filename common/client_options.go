package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/sender"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/hashicorp/terraform-provider-azurerm/features"
	"github.com/hashicorp/terraform-provider-azurerm/version"
)

type EndpointTokenFunc func(endpoint string) (autorest.Authorizer, error)

type ClientOptions struct {
	SubscriptionId   string
	TenantID         string
	PartnerId        string
	TerraformVersion string

	KeyVaultAuthorizer        autorest.Authorizer
	ResourceManagerAuthorizer autorest.Authorizer
	ResourceManagerEndpoint   string
	StorageAuthorizer         autorest.Authorizer
	SynapseAuthorizer         autorest.Authorizer
	BatchManagementAuthorizer autorest.Authorizer

	SkipProviderReg             bool
	CustomCorrelationRequestID  string
	DisableCorrelationRequestID bool
	DisableTerraformPartnerID   bool
	Environment                 azure.Environment
	Features                    features.UserFeatures
	StorageUseAzureAD           bool

	// Some Dataplane APIs require a token scoped for a specific endpoint
	TokenFunc EndpointTokenFunc

	// TODO: remove graph configuration in v3.0
	GraphAuthorizer autorest.Authorizer
	GraphEndpoint   string
}

func (o ClientOptions) ConfigureClient(c *autorest.Client, authorizer autorest.Authorizer) {
	setUserAgent(c, o.TerraformVersion, o.PartnerId, o.DisableTerraformPartnerID)

	c.Authorizer = authorizer
	c.Sender = sender.BuildSender("AzureRM")
	c.SkipResourceProviderRegistration = o.SkipProviderReg
	if !o.DisableCorrelationRequestID {
		id := o.CustomCorrelationRequestID
		if id == "" {
			id = correlationRequestID()
		}
		c.RequestInspector = withCorrelationRequestID(id)
	}
}

func setUserAgent(client *autorest.Client, tfVersion, partnerID string, disableTerraformPartnerID bool) {
	tfUserAgent := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io) Terraform Plugin SDK/%s", tfVersion, meta.SDKVersionString())

	providerUserAgent := fmt.Sprintf("%s terraform-provider-azurerm/%s", tfUserAgent, version.ProviderVersion)
	if features.FourPointOhBeta() {
		providerUserAgent = fmt.Sprintf("%s terraform-provider-azurerm/%s+4.0-beta", tfUserAgent, version.ProviderVersion)
	}
	client.UserAgent = strings.TrimSpace(fmt.Sprintf("%s %s", client.UserAgent, providerUserAgent))

	// append the CloudShell version to the user agent if it exists
	if azureAgent := os.Getenv("AZURE_HTTP_USER_AGENT"); azureAgent != "" {
		client.UserAgent = fmt.Sprintf("%s %s", client.UserAgent, azureAgent)
	}

	// only one pid can be interpreted currently
	// hence, send partner ID if present, otherwise send Terraform GUID
	// unless users have opted out
	if partnerID == "" && !disableTerraformPartnerID {
		// Microsoft’s Terraform Partner ID is this specific GUID
		partnerID = "222c6c49-1b0a-5959-a213-6608f9eb8820"
	}

	if partnerID != "" {
		client.UserAgent = fmt.Sprintf("%s pid-%s", client.UserAgent, partnerID)
	}
}
