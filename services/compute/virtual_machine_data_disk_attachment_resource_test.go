package compute_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-11-01/compute"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/compute/parse"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type VirtualMachineDataDiskAttachmentResource struct{}

func TestAccVirtualMachineDataDiskAttachment_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_machine_data_disk_attachment", "test")
	r := VirtualMachineDataDiskAttachmentResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("virtual_machine_id").Exists(),
				check.That(data.ResourceName).Key("managed_disk_id").Exists(),
				check.That(data.ResourceName).Key("lun").HasValue("0"),
				check.That(data.ResourceName).Key("caching").HasValue("None"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVirtualMachineDataDiskAttachment_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_machine_data_disk_attachment", "test")
	r := VirtualMachineDataDiskAttachmentResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_virtual_machine_data_disk_attachment"),
		},
	})
}

func TestAccVirtualMachineDataDiskAttachment_destroy(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_machine_data_disk_attachment", "test")
	r := VirtualMachineDataDiskAttachmentResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		data.DisappearsStep(acceptance.DisappearsStepData{
			Config:       r.basic,
			TestResource: r,
		}),
	})
}

func TestAccVirtualMachineDataDiskAttachment_multipleDisks(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_machine_data_disk_attachment", "first")
	r := VirtualMachineDataDiskAttachmentResource{}

	secondResourceName := "azurerm_virtual_machine_data_disk_attachment.second"

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.multipleDisks(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("virtual_machine_id").Exists(),
				check.That(data.ResourceName).Key("managed_disk_id").Exists(),
				check.That(data.ResourceName).Key("lun").HasValue("10"),
				check.That(data.ResourceName).Key("caching").HasValue("None"),
				acceptance.TestCheckResourceAttrSet(secondResourceName, "virtual_machine_id"),
				acceptance.TestCheckResourceAttrSet(secondResourceName, "managed_disk_id"),
				acceptance.TestCheckResourceAttr(secondResourceName, "lun", "20"),
				acceptance.TestCheckResourceAttr(secondResourceName, "caching", "ReadOnly"),
			),
		},
		data.ImportStep(),
		{
			ResourceName:      secondResourceName,
			ImportState:       true,
			ImportStateVerify: true,
		},
	})
}

func TestAccVirtualMachineDataDiskAttachment_updatingCaching(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_machine_data_disk_attachment", "test")
	r := VirtualMachineDataDiskAttachmentResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("caching").HasValue("None"),
			),
		},
		{
			Config: r.readOnly(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("caching").HasValue("ReadOnly"),
			),
		},
		{
			Config: r.readWrite(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("caching").HasValue("ReadWrite"),
			),
		},
	})
}

func TestAccVirtualMachineDataDiskAttachment_updatingWriteAccelerator(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_machine_data_disk_attachment", "test")
	r := VirtualMachineDataDiskAttachmentResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.writeAccelerator(data, false),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("write_accelerator_enabled").HasValue("false"),
			),
		},
		{
			Config: r.writeAccelerator(data, true),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("write_accelerator_enabled").HasValue("true"),
			),
		},
		{
			Config: r.writeAccelerator(data, false),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("write_accelerator_enabled").HasValue("false"),
			),
		},
	})
}

func TestAccVirtualMachineDataDiskAttachment_managedServiceIdentity(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_machine_data_disk_attachment", "test")
	r := VirtualMachineDataDiskAttachmentResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.managedServiceIdentity(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("virtual_machine_id").Exists(),
				check.That(data.ResourceName).Key("managed_disk_id").Exists(),
				check.That(data.ResourceName).Key("lun").HasValue("0"),
				check.That(data.ResourceName).Key("caching").HasValue("None"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccVirtualMachineDataDiskAttachment_virtualMachineExtension(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_virtual_machine_data_disk_attachment", "test")
	r := VirtualMachineDataDiskAttachmentResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.virtualMachineExtensionPrep(data),
		},
		{
			Config: r.virtualMachineExtensionComplete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("virtual_machine_id").Exists(),
				check.That(data.ResourceName).Key("managed_disk_id").Exists(),
			),
		},
	})
}

func (t VirtualMachineDataDiskAttachmentResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.DataDiskID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Compute.VMClient.Get(ctx, id.ResourceGroup, id.VirtualMachineName, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Virtual Machine Data Disk Attachment %q", id)
	}

	var disk *compute.DataDisk
	if profile := resp.StorageProfile; profile != nil {
		if dataDisks := profile.DataDisks; dataDisks != nil {
			for _, dataDisk := range *dataDisks {
				// since this field isn't (and shouldn't be) case-sensitive; we're deliberately not using `strings.EqualFold`
				if *dataDisk.Name == id.Name {
					disk = &dataDisk
					break
				}
			}
		}
	}

	return utils.Bool(disk != nil), nil
}

func (VirtualMachineDataDiskAttachmentResource) Destroy(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.DataDiskID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Compute.VMClient.Get(ctx, id.ResourceGroup, id.VirtualMachineName, "")
	if err != nil {
		return nil, fmt.Errorf("retrieving Compute Virtual Machine Data Disk Attachment %q", id)
	}

	outputDisks := make([]compute.DataDisk, 0)
	for _, disk := range *resp.StorageProfile.DataDisks {
		// deliberately not using strings.Equals as this is case sensitive
		if *disk.Name == id.Name {
			continue
		}

		outputDisks = append(outputDisks, disk)
	}
	resp.StorageProfile.DataDisks = &outputDisks

	// fixes #2485
	resp.Identity = nil
	// fixes #1600
	resp.Resources = nil

	future, err := client.Compute.VMClient.CreateOrUpdate(ctx, id.ResourceGroup, id.VirtualMachineName, resp)
	if err != nil {
		return nil, fmt.Errorf("updating Virtual Machine %q: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Compute.VMClient.Client); err != nil {
		return nil, fmt.Errorf("waiting for Virtual Machine %q: %+v", id, err)
	}

	return utils.Bool(true), nil
}

func (r VirtualMachineDataDiskAttachmentResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_machine_data_disk_attachment" "test" {
  managed_disk_id    = azurerm_managed_disk.test.id
  virtual_machine_id = azurerm_virtual_machine.test.id
  lun                = "0"
  caching            = "None"
}
`, r.template(data))
}

func (r VirtualMachineDataDiskAttachmentResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_machine_data_disk_attachment" "import" {
  managed_disk_id    = azurerm_virtual_machine_data_disk_attachment.test.managed_disk_id
  virtual_machine_id = azurerm_virtual_machine_data_disk_attachment.test.virtual_machine_id
  lun                = azurerm_virtual_machine_data_disk_attachment.test.lun
  caching            = azurerm_virtual_machine_data_disk_attachment.test.caching
}
`, r.basic(data))
}

func (VirtualMachineDataDiskAttachmentResource) managedServiceIdentity(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_network_interface" "test" {
  name                = "acctni-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = azurerm_resource_group.test.location
  resource_group_name   = azurerm_resource_group.test.name
  network_interface_ids = [azurerm_network_interface.test.id]
  vm_size               = "Standard_F2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_managed_disk" "test" {
  name                 = "%d-disk1"
  location             = azurerm_resource_group.test.location
  resource_group_name  = azurerm_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = 10
}

resource "azurerm_virtual_machine_data_disk_attachment" "test" {
  managed_disk_id    = azurerm_managed_disk.test.id
  virtual_machine_id = azurerm_virtual_machine.test.id
  lun                = "0"
  caching            = "None"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r VirtualMachineDataDiskAttachmentResource) multipleDisks(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_machine_data_disk_attachment" "first" {
  managed_disk_id    = azurerm_managed_disk.test.id
  virtual_machine_id = azurerm_virtual_machine.test.id
  lun                = "10"
  caching            = "None"
}

resource "azurerm_managed_disk" "second" {
  name                 = "%d-disk2"
  location             = azurerm_resource_group.test.location
  resource_group_name  = azurerm_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = 10
}

resource "azurerm_virtual_machine_data_disk_attachment" "second" {
  managed_disk_id    = azurerm_managed_disk.second.id
  virtual_machine_id = azurerm_virtual_machine.test.id
  lun                = "20"
  caching            = "ReadOnly"
}
`, r.template(data), data.RandomInteger)
}

func (r VirtualMachineDataDiskAttachmentResource) readOnly(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_machine_data_disk_attachment" "test" {
  managed_disk_id    = azurerm_managed_disk.test.id
  virtual_machine_id = azurerm_virtual_machine.test.id
  lun                = "0"
  caching            = "ReadOnly"
}
`, r.template(data))
}

func (r VirtualMachineDataDiskAttachmentResource) readWrite(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_machine_data_disk_attachment" "test" {
  managed_disk_id    = azurerm_managed_disk.test.id
  virtual_machine_id = azurerm_virtual_machine.test.id
  lun                = "0"
  caching            = "ReadWrite"
}
`, r.template(data))
}

func (VirtualMachineDataDiskAttachmentResource) writeAccelerator(data acceptance.TestData, enabled bool) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_network_interface" "test" {
  name                = "acctni-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = azurerm_resource_group.test.location
  resource_group_name   = azurerm_resource_group.test.name
  network_interface_ids = [azurerm_network_interface.test.id]
  vm_size               = "Standard_M64s"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Premium_LRS"
  }

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "azurerm_managed_disk" "test" {
  name                 = "%d-disk1"
  location             = azurerm_resource_group.test.location
  resource_group_name  = azurerm_resource_group.test.name
  storage_account_type = "Premium_LRS"
  create_option        = "Empty"
  disk_size_gb         = 10
}

resource "azurerm_virtual_machine_data_disk_attachment" "test" {
  managed_disk_id           = azurerm_managed_disk.test.id
  virtual_machine_id        = azurerm_virtual_machine.test.id
  lun                       = "0"
  caching                   = "None"
  write_accelerator_enabled = %t
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, enabled)
}

func (VirtualMachineDataDiskAttachmentResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "acctsub-%d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_network_interface" "test" {
  name                = "acctni-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_virtual_machine" "test" {
  name                  = "acctvm-%d"
  location              = azurerm_resource_group.test.location
  resource_group_name   = azurerm_resource_group.test.name
  network_interface_ids = [azurerm_network_interface.test.id]
  vm_size               = "Standard_F2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "hn%d"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "azurerm_managed_disk" "test" {
  name                 = "%d-disk1"
  location             = azurerm_resource_group.test.location
  resource_group_name  = azurerm_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = 10
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (VirtualMachineDataDiskAttachmentResource) virtualMachineExtensionPrep(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctestvn-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "acctsub"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_public_ip" "test" {
  name                = "acctestpip%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurerm_network_interface" "test" {
  name                = "acctestni%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.test.id
  }
}

resource "azurerm_virtual_machine" "test" {
  name                  = "acctestvm%d"
  location              = azurerm_resource_group.test.location
  resource_group_name   = azurerm_resource_group.test.name
  network_interface_ids = [azurerm_network_interface.test.id]
  vm_size               = "Standard_F4"

  delete_os_disk_on_termination    = true
  delete_data_disks_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  os_profile {
    computer_name  = "testvm"
    admin_username = "tfuser123"
    admin_password = "Password1234!"
  }

  storage_os_disk {
    name              = "myosdisk1"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  tags = {
    environment = "staging"
  }
}

resource "azurerm_virtual_machine_extension" "test" {
  name                 = "random-script"
  virtual_machine_id   = azurerm_virtual_machine.test.id
  publisher            = "Microsoft.Azure.Extensions"
  type                 = "CustomScript"
  type_handler_version = "2.0"

  settings = <<SETTINGS
	{
		"commandToExecute": "hostname"
	}
SETTINGS

  tags = {
    environment = "Production"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (r VirtualMachineDataDiskAttachmentResource) virtualMachineExtensionComplete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_managed_disk" "test" {
  name                 = "acctest%d"
  location             = azurerm_resource_group.test.location
  resource_group_name  = azurerm_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = 10
}

resource "azurerm_virtual_machine_data_disk_attachment" "test" {
  managed_disk_id    = azurerm_managed_disk.test.id
  virtual_machine_id = azurerm_virtual_machine.test.id
  lun                = "11"
  caching            = "ReadWrite"
}
`, r.virtualMachineExtensionPrep(data), data.RandomInteger)
}
