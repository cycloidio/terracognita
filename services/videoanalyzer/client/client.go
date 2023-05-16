package client

import (
	"github.com/hashicorp/terraform-provider-azurerm/common"
	"github.com/hashicorp/terraform-provider-azurerm/services/videoanalyzer/sdk/2021-05-01-preview/videoanalyzer"
)

type Client struct {
	VideoAnalyzersClient *videoanalyzer.VideoAnalyzerClient
}

func NewClient(o *common.ClientOptions) *Client {
	VideoAnalyzersClient := videoanalyzer.NewVideoAnalyzerClientWithBaseURI(o.ResourceManagerEndpoint)

	o.ConfigureClient(&VideoAnalyzersClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		VideoAnalyzersClient: &VideoAnalyzersClient,
	}
}
