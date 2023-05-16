package compute_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type SharedImageGalleryResource struct{}

func TestAccSharedImageGallery_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_shared_image_gallery", "test")
	r := SharedImageGalleryResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("description").HasValue(""),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSharedImageGallery_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_shared_image_gallery", "test")
	r := SharedImageGalleryResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("description").HasValue(""),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_shared_image_gallery"),
		},
	})
}

func TestAccSharedImageGallery_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_shared_image_gallery", "test")
	r := SharedImageGalleryResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("description").HasValue("Shared images and things."),
				check.That(data.ResourceName).Key("tags.%").HasValue("2"),
				check.That(data.ResourceName).Key("tags.Hello").HasValue("There"),
				check.That(data.ResourceName).Key("tags.World").HasValue("Example"),
			),
		},
		data.ImportStep(),
	})
}

func (t SharedImageGalleryResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.SharedImageGalleryID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.GalleriesClient.Get(ctx, id.ResourceGroup, id.GalleryName, "", "")
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Shared Image Gallery %q", id.String())
	}

	return utils.Bool(resp.ID != nil), nil
}

func (SharedImageGalleryResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_shared_image_gallery" "test" {
  name                = "acctestsig%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r SharedImageGalleryResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_shared_image_gallery" "import" {
  name                = azurerm_shared_image_gallery.test.name
  resource_group_name = azurerm_shared_image_gallery.test.resource_group_name
  location            = azurerm_shared_image_gallery.test.location
}
`, r.basic(data))
}

func (SharedImageGalleryResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_shared_image_gallery" "test" {
  name                = "acctestsig%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  description         = "Shared images and things."

  tags = {
    Hello = "There"
    World = "Example"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
