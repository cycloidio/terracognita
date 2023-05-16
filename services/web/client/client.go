package client

import (
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2021-02-01/web"
	"github.com/hashicorp/terraform-provider-azurerm/common"
)

type Client struct {
	AppServiceEnvironmentsClient *web.AppServiceEnvironmentsClient
	AppServicePlansClient        *web.AppServicePlansClient
	AppServicesClient            *web.AppsClient
	BaseClient                   *web.BaseClient
	CertificatesClient           *web.CertificatesClient
	CertificatesOrderClient      *web.AppServiceCertificateOrdersClient
	StaticSitesClient            *web.StaticSitesClient
}

func NewClient(o *common.ClientOptions) *Client {
	appServiceEnvironmentsClient := web.NewAppServiceEnvironmentsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&appServiceEnvironmentsClient.Client, o.ResourceManagerAuthorizer)

	appServicePlansClient := web.NewAppServicePlansClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&appServicePlansClient.Client, o.ResourceManagerAuthorizer)

	appServicesClient := web.NewAppsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&appServicesClient.Client, o.ResourceManagerAuthorizer)

	baseClient := web.NewWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&baseClient.Client, o.ResourceManagerAuthorizer)

	certificatesClient := web.NewCertificatesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&certificatesClient.Client, o.ResourceManagerAuthorizer)

	certificatesOrderClient := web.NewAppServiceCertificateOrdersClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&certificatesOrderClient.Client, o.ResourceManagerAuthorizer)

	staticSitesClient := web.NewStaticSitesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&staticSitesClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		AppServiceEnvironmentsClient: &appServiceEnvironmentsClient,
		AppServicePlansClient:        &appServicePlansClient,
		AppServicesClient:            &appServicesClient,
		BaseClient:                   &baseClient,
		CertificatesClient:           &certificatesClient,
		CertificatesOrderClient:      &certificatesOrderClient,
		StaticSitesClient:            &staticSitesClient,
	}
}
