package batch

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/batch/mgmt/2021-06-01/batch"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/batch/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/batch/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceBatchCertificate() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceBatchCertificateCreate,
		Read:   resourceBatchCertificateRead,
		Update: resourceBatchCertificateUpdate,
		Delete: resourceBatchCertificateDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.CertificateID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"account_name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.AccountName,
			},

			// TODO: make this case sensitive once this API bug has been fixed:
			// https://github.com/Azure/azure-rest-api-specs/issues/5574
			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),

			"certificate": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 10000),
			},

			"format": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(batch.CertificateFormatCer),
					string(batch.CertificateFormatPfx),
				}, false),
			},

			"password": {
				Type:      pluginsdk.TypeString,
				Optional:  true, // Cannot be used when `format` is "Cer"
				Sensitive: true,
			},

			"thumbprint": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"thumbprint_algorithm": {
				Type:             pluginsdk.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     validation.StringInSlice([]string{"SHA1"}, false),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"public_data": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBatchCertificateCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Batch.CertificateClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure Batch certificate creation.")

	certificate := d.Get("certificate").(string)
	format := d.Get("format").(string)
	password := d.Get("password").(string)
	thumbprint := d.Get("thumbprint").(string)
	thumbprintAlgorithm := d.Get("thumbprint_algorithm").(string)
	name := thumbprintAlgorithm + "-" + thumbprint
	id := parse.NewCertificateID(subscriptionId, d.Get("resource_group_name").(string), d.Get("account_name").(string), name)

	if err := validateBatchCertificateFormatAndPassword(format, password); err != nil {
		return err
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.BatchAccountName, id.Name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_batch_certificate", id.ID())
		}
	}
	certificateProperties := batch.CertificateCreateOrUpdateProperties{
		Data:                &certificate,
		Format:              batch.CertificateFormat(format),
		Thumbprint:          &thumbprint,
		ThumbprintAlgorithm: &thumbprintAlgorithm,
	}
	if password != "" {
		certificateProperties.Password = &password
	}
	parameters := batch.CertificateCreateOrUpdateParameters{
		Name:                                &name,
		CertificateCreateOrUpdateProperties: &certificateProperties,
	}

	_, err := client.Create(ctx, id.ResourceGroup, id.BatchAccountName, id.Name, parameters, "", "")
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceBatchCertificateRead(d, meta)
}

func resourceBatchCertificateRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Batch.CertificateClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CertificateID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.BatchAccountName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			log.Printf("[DEBUG] Batch certificate %q was not found in Account %q / Resource Group %q - removing from state!", id.Name, id.BatchAccountName, id.ResourceGroup)
			return nil
		}
		return fmt.Errorf("retrieving Batch Certificate %q (Account %q / Resource Group %q): %+v", id.Name, id.BatchAccountName, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("account_name", id.BatchAccountName)
	d.Set("resource_group_name", id.ResourceGroup)

	if props := resp.CertificateProperties; props != nil {
		d.Set("format", props.Format)
		d.Set("public_data", props.PublicData)
		d.Set("thumbprint", props.Thumbprint)
		d.Set("thumbprint_algorithm", props.ThumbprintAlgorithm)
	}

	return nil
}

func resourceBatchCertificateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Batch.CertificateClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure Batch certificate update.")

	id, err := parse.CertificateID(d.Id())
	if err != nil {
		return err
	}

	certificate := d.Get("certificate").(string)
	format := d.Get("format").(string)
	password := d.Get("password").(string)
	thumbprint := d.Get("thumbprint").(string)
	thumbprintAlgorithm := d.Get("thumbprint_algorithm").(string)

	if err := validateBatchCertificateFormatAndPassword(format, password); err != nil {
		return err
	}

	parameters := batch.CertificateCreateOrUpdateParameters{
		Name: &id.Name,
		CertificateCreateOrUpdateProperties: &batch.CertificateCreateOrUpdateProperties{
			Data:                &certificate,
			Format:              batch.CertificateFormat(format),
			Password:            &password,
			Thumbprint:          &thumbprint,
			ThumbprintAlgorithm: &thumbprintAlgorithm,
		},
	}

	if _, err = client.Update(ctx, id.ResourceGroup, id.BatchAccountName, id.Name, parameters, ""); err != nil {
		return fmt.Errorf("updating Batch certificate %q (Account %q / Resource Group %q): %+v", id.Name, id.BatchAccountName, id.ResourceGroup, err)
	}

	read, err := client.Get(ctx, id.ResourceGroup, id.BatchAccountName, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving Batch Certificate %q (Account %q / Resource Group %q): %+v", id.Name, id.BatchAccountName, id.ResourceGroup, err)
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read ID for Batch certificate %q (Account: %q, Resource Group %q) ID", id.Name, id.BatchAccountName, id.ResourceGroup)
	}

	return resourceBatchCertificateRead(d, meta)
}

func resourceBatchCertificateDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Batch.CertificateClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CertificateID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.BatchAccountName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Batch Certificate %q (Account %q / Resource Group %q): %+v", id.Name, id.BatchAccountName, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if !response.WasNotFound(future.Response()) {
			return fmt.Errorf("waiting for deletion of Batch Certificate %q (Account %q / Resource Group %q): %+v", id.Name, id.BatchAccountName, id.ResourceGroup, err)
		}
	}

	return nil
}

func validateBatchCertificateFormatAndPassword(format string, password string) error {
	if format == "Cer" && password != "" {
		return fmt.Errorf(" Batch Certificate Password must not be specified when Format is `Cer`")
	}
	return nil
}
