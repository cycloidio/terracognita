package storage_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/services/storage"
)

type StorageAccountSasDataSource struct{}

func TestAccDataSourceStorageAccountSas_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_storage_account_sas", "test")
	ipAddresses := "10.0.0.1-10.0.0.4"
	utcNow := time.Now().UTC()
	startDate := utcNow.Format(time.RFC3339)
	endDate := utcNow.Add(time.Hour * 24).Format(time.RFC3339)

	data.DataSourceTest(t, []acceptance.TestStep{
		{
			Config: StorageAccountSasDataSource{}.basic(data, startDate, endDate, ipAddresses),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("https_only").HasValue("true"),
				check.That(data.ResourceName).Key("ip_addresses").HasValue(ipAddresses),
				check.That(data.ResourceName).Key("signed_version").HasValue("2019-10-10"),
				check.That(data.ResourceName).Key("start").HasValue(startDate),
				check.That(data.ResourceName).Key("expiry").HasValue(endDate),
				check.That(data.ResourceName).Key("sas").Exists(),
			),
		},
	})
}

func (d StorageAccountSasDataSource) basic(data acceptance.TestData, startDate string, endDate string, ipAddresses string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-storage-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                = "acctestsads%s"
  resource_group_name = azurerm_resource_group.test.name

  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "production"
  }
}

data "azurerm_storage_account_sas" "test" {
  connection_string = azurerm_storage_account.test.primary_connection_string
  https_only        = true
  ip_addresses      = "%s"
  signed_version    = "2019-10-10"

  resource_types {
    service   = true
    container = false
    object    = false
  }

  services {
    blob  = true
    queue = false
    table = false
    file  = false
  }

  start  = "%s"
  expiry = "%s"

  permissions {
    read    = true
    write   = true
    delete  = false
    list    = false
    add     = true
    create  = true
    update  = false
    process = false
    tag     = false
    filter  = false
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, ipAddresses, startDate, endDate)
}

func TestAccDataSourceStorageAccountSas_resourceTypesString(t *testing.T) {
	testCases := []struct {
		input    map[string]interface{}
		expected string
	}{
		{map[string]interface{}{"service": true}, "s"},
		{map[string]interface{}{"container": true}, "c"},
		{map[string]interface{}{"object": true}, "o"},
		{map[string]interface{}{"service": true, "container": true, "object": true}, "sco"},
	}

	for _, test := range testCases {
		result := storage.BuildResourceTypesString(test.input)
		if test.expected != result {
			t.Fatalf("Failed to build resource type string: expected: %s, result: %s", test.expected, result)
		}
	}
}

func TestAccDataSourceStorageAccountSas_servicesString(t *testing.T) {
	testCases := []struct {
		input    map[string]interface{}
		expected string
	}{
		{map[string]interface{}{"blob": true}, "b"},
		{map[string]interface{}{"queue": true}, "q"},
		{map[string]interface{}{"table": true}, "t"},
		{map[string]interface{}{"file": true}, "f"},
		{map[string]interface{}{"blob": true, "queue": true, "table": true, "file": true}, "bqtf"},
	}

	for _, test := range testCases {
		result := storage.BuildServicesString(test.input)
		if test.expected != result {
			t.Fatalf("Failed to build resource type string: expected: %s, result: %s", test.expected, result)
		}
	}
}

func TestAccDataSourceStorageAccountSas_permissionsString(t *testing.T) {
	testCases := []struct {
		input    map[string]interface{}
		expected string
	}{
		{map[string]interface{}{"read": true}, "r"},
		{map[string]interface{}{"write": true}, "w"},
		{map[string]interface{}{"delete": true}, "d"},
		{map[string]interface{}{"list": true}, "l"},
		{map[string]interface{}{"add": true}, "a"},
		{map[string]interface{}{"create": true}, "c"},
		{map[string]interface{}{"update": true}, "u"},
		{map[string]interface{}{"process": true}, "p"},
		{map[string]interface{}{"tag": true}, "t"},
		{map[string]interface{}{"filter": true}, "f"},
		{map[string]interface{}{"read": true, "write": true, "add": true, "create": true}, "rwac"},
	}

	for _, test := range testCases {
		result := storage.BuildPermissionsString(test.input)
		if test.expected != result {
			t.Fatalf("Failed to build resource type string: expected: %s, result: %s", test.expected, result)
		}
	}
}
