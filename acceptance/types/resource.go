package types

import (
	"context"

	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
)

type TestResource interface {
	Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error)
}

type TestResourceVerifyingRemoved interface {
	TestResource
	Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error)
}
