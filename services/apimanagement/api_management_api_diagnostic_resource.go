package apimanagement

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2021-08-01/apimanagement"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/clients"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/parse"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/schemaz"
	"github.com/hashicorp/terraform-provider-azurerm/services/apimanagement/validate"
	"github.com/hashicorp/terraform-provider-azurerm/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/tf/set"
	"github.com/hashicorp/terraform-provider-azurerm/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceApiManagementApiDiagnostic() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceApiManagementApiDiagnosticCreateUpdate,
		Read:   resourceApiManagementApiDiagnosticRead,
		Update: resourceApiManagementApiDiagnosticCreateUpdate,
		Delete: resourceApiManagementApiDiagnosticDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ApiDiagnosticID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"identifier": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"applicationinsights",
					"azuremonitor",
				}, false),
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"api_management_name": schemaz.SchemaApiManagementName(),

			"api_name": schemaz.SchemaApiManagementApiName(),

			"api_management_logger_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validate.LoggerID,
			},

			"sampling_percentage": {
				Type:         pluginsdk.TypeFloat,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.FloatBetween(0.0, 100.0),
			},

			"always_log_errors": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Computed: true,
			},

			"verbosity": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(apimanagement.VerbosityVerbose),
					string(apimanagement.VerbosityInformation),
					string(apimanagement.VerbosityError),
				}, false),
			},

			"log_client_ip": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Computed: true,
			},

			"http_correlation_protocol": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(apimanagement.HTTPCorrelationProtocolNone),
					string(apimanagement.HTTPCorrelationProtocolLegacy),
					string(apimanagement.HTTPCorrelationProtocolW3C),
				}, false),
			},

			"frontend_request": resourceApiManagementApiDiagnosticAdditionalContentSchema(),

			"frontend_response": resourceApiManagementApiDiagnosticAdditionalContentSchema(),

			"backend_request": resourceApiManagementApiDiagnosticAdditionalContentSchema(),

			"backend_response": resourceApiManagementApiDiagnosticAdditionalContentSchema(),

			"operation_name_format": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Default:  string(apimanagement.OperationNameFormatName),
				ValidateFunc: validation.StringInSlice([]string{
					string(apimanagement.OperationNameFormatName),
					string(apimanagement.OperationNameFormatURL),
				}, false),
			},
		},
	}
}

func resourceApiManagementApiDiagnosticAdditionalContentSchema() *pluginsdk.Schema {
	//lintignore:XS003
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		MaxItems: 1,
		Optional: true,
		Computed: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"body_bytes": {
					Type:         pluginsdk.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntBetween(0, 8192),
				},
				"headers_to_log": {
					Type:     pluginsdk.TypeSet,
					Optional: true,
					Elem: &pluginsdk.Schema{
						Type: pluginsdk.TypeString,
					},
					Set: pluginsdk.HashString,
				},
				"data_masking": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"query_params": schemaApiManagementDataMaskingEntityList(),
							"headers":      schemaApiManagementDataMaskingEntityList(),
						},
					},
				},
			},
		},
	}
}

func resourceApiManagementApiDiagnosticCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.ApiDiagnosticClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := parse.NewApiDiagnosticID(subscriptionId, d.Get("resource_group_name").(string), d.Get("api_management_name").(string), d.Get("api_name").(string), d.Get("identifier").(string))

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.ServiceName, id.ApiName, id.DiagnosticName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Diagnostic %s: %s", id, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_api_management_api_diagnostic", id.ID())
		}
	}

	parameters := apimanagement.DiagnosticContract{
		DiagnosticContractProperties: &apimanagement.DiagnosticContractProperties{
			LoggerID:            utils.String(d.Get("api_management_logger_id").(string)),
			OperationNameFormat: apimanagement.OperationNameFormat(d.Get("operation_name_format").(string)),
		},
	}

	if samplingPercentage, ok := d.GetOk("sampling_percentage"); ok {
		parameters.Sampling = &apimanagement.SamplingSettings{
			SamplingType: apimanagement.SamplingTypeFixed,
			Percentage:   utils.Float(samplingPercentage.(float64)),
		}
	} else {
		parameters.Sampling = nil
	}

	if alwaysLogErrors, ok := d.GetOk("always_log_errors"); ok && alwaysLogErrors.(bool) {
		parameters.AlwaysLog = apimanagement.AlwaysLogAllErrors
	}

	if verbosity, ok := d.GetOk("verbosity"); ok {
		parameters.Verbosity = apimanagement.Verbosity(verbosity.(string))
	}

	//lint:ignore SA1019 SDKv2 migration  - staticcheck's own linter directives are currently being ignored under golanci-lint
	if logClientIP, exists := d.GetOkExists("log_client_ip"); exists { //nolint:staticcheck
		parameters.LogClientIP = utils.Bool(logClientIP.(bool))
	}

	if httpCorrelationProtocol, ok := d.GetOk("http_correlation_protocol"); ok {
		parameters.HTTPCorrelationProtocol = apimanagement.HTTPCorrelationProtocol(httpCorrelationProtocol.(string))
	}

	frontendRequest, frontendRequestSet := d.GetOk("frontend_request")
	frontendResponse, frontendResponseSet := d.GetOk("frontend_response")
	if frontendRequestSet || frontendResponseSet {
		parameters.Frontend = &apimanagement.PipelineDiagnosticSettings{}
		if frontendRequestSet {
			parameters.Frontend.Request = expandApiManagementApiDiagnosticHTTPMessageDiagnostic(frontendRequest.([]interface{}))
		}
		if frontendResponseSet {
			parameters.Frontend.Response = expandApiManagementApiDiagnosticHTTPMessageDiagnostic(frontendResponse.([]interface{}))
		}
	}

	backendRequest, backendRequestSet := d.GetOk("backend_request")
	backendResponse, backendResponseSet := d.GetOk("backend_response")
	if backendRequestSet || backendResponseSet {
		parameters.Backend = &apimanagement.PipelineDiagnosticSettings{}
		if backendRequestSet {
			parameters.Backend.Request = expandApiManagementApiDiagnosticHTTPMessageDiagnostic(backendRequest.([]interface{}))
		}
		if backendResponseSet {
			parameters.Backend.Response = expandApiManagementApiDiagnosticHTTPMessageDiagnostic(backendResponse.([]interface{}))
		}
	}

	if _, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.ServiceName, id.ApiName, id.DiagnosticName, parameters, ""); err != nil {
		return fmt.Errorf("creating or updating Diagnostic %s: %+v", id, err)
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ServiceName, id.ApiName, id.DiagnosticName)
	if err != nil {
		return fmt.Errorf("retrieving Diagnostic %s: %+v", id, err)
	}
	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("reading ID for Diagnostic %s: ID is empty", id)
	}
	d.SetId(id.ID())

	return resourceApiManagementApiDiagnosticRead(d, meta)
}

func resourceApiManagementApiDiagnosticRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.ApiDiagnosticClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	diagnosticId, err := parse.ApiDiagnosticID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, diagnosticId.ResourceGroup, diagnosticId.ServiceName, diagnosticId.ApiName, diagnosticId.DiagnosticName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Diagnostic %q (Resource Group %q / API Management Service %q / API %q) was not found - removing from state!", diagnosticId.DiagnosticName, diagnosticId.ResourceGroup, diagnosticId.ServiceName, diagnosticId.ApiName)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("making Read request for Diagnostic %q (Resource Group %q / API Management Service %q / API %q): %+v", diagnosticId.DiagnosticName, diagnosticId.ResourceGroup, diagnosticId.ServiceName, diagnosticId.ApiName, err)
	}

	d.Set("api_name", diagnosticId.ApiName)
	d.Set("identifier", resp.Name)
	d.Set("resource_group_name", diagnosticId.ResourceGroup)
	d.Set("api_management_name", diagnosticId.ServiceName)
	if props := resp.DiagnosticContractProperties; props != nil {
		d.Set("api_management_logger_id", props.LoggerID)
		if props.Sampling != nil && props.Sampling.Percentage != nil {
			d.Set("sampling_percentage", props.Sampling.Percentage)
		}
		d.Set("always_log_errors", props.AlwaysLog == apimanagement.AlwaysLogAllErrors)
		d.Set("verbosity", props.Verbosity)
		d.Set("log_client_ip", props.LogClientIP)
		d.Set("http_correlation_protocol", props.HTTPCorrelationProtocol)
		if frontend := props.Frontend; frontend != nil {
			d.Set("frontend_request", flattenApiManagementApiDiagnosticHTTPMessageDiagnostic(frontend.Request))
			d.Set("frontend_response", flattenApiManagementApiDiagnosticHTTPMessageDiagnostic(frontend.Response))
		} else {
			d.Set("frontend_request", nil)
			d.Set("frontend_response", nil)
		}
		if backend := props.Backend; backend != nil {
			d.Set("backend_request", flattenApiManagementApiDiagnosticHTTPMessageDiagnostic(backend.Request))
			d.Set("backend_response", flattenApiManagementApiDiagnosticHTTPMessageDiagnostic(backend.Response))
		} else {
			d.Set("backend_request", nil)
			d.Set("backend_response", nil)
		}

		format := string(apimanagement.OperationNameFormatName)
		if props.OperationNameFormat != "" {
			format = string(props.OperationNameFormat)
		}
		d.Set("operation_name_format", format)
	}

	return nil
}

func resourceApiManagementApiDiagnosticDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.ApiDiagnosticClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	diagnosticId, err := parse.ApiDiagnosticID(d.Id())
	if err != nil {
		return err
	}

	if resp, err := client.Delete(ctx, diagnosticId.ResourceGroup, diagnosticId.ServiceName, diagnosticId.ApiName, diagnosticId.DiagnosticName, ""); err != nil {
		if !utils.ResponseWasNotFound(resp) {
			return fmt.Errorf("deleting Diagnostic %q (Resource Group %q / API Management Service %q / API %q): %+v", diagnosticId.DiagnosticName, diagnosticId.ResourceGroup, diagnosticId.ServiceName, diagnosticId.ApiName, err)
		}
	}

	return nil
}

func expandApiManagementApiDiagnosticHTTPMessageDiagnostic(input []interface{}) *apimanagement.HTTPMessageDiagnostic {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	v := input[0].(map[string]interface{})

	result := &apimanagement.HTTPMessageDiagnostic{
		Body: &apimanagement.BodyDiagnosticSettings{},
	}

	if bodyBytes, ok := v["body_bytes"]; ok {
		result.Body.Bytes = utils.Int32(int32(bodyBytes.(int)))
	}
	if headersSetRaw, ok := v["headers_to_log"]; ok {
		headersSet := headersSetRaw.(*pluginsdk.Set).List()
		headers := []string{}
		for _, header := range headersSet {
			headers = append(headers, header.(string))
		}
		result.Headers = &headers
	}

	result.DataMasking = expandApiManagementDataMasking(v["data_masking"].([]interface{}))

	return result
}

func flattenApiManagementApiDiagnosticHTTPMessageDiagnostic(input *apimanagement.HTTPMessageDiagnostic) []interface{} {
	result := make([]interface{}, 0)

	if input == nil {
		return result
	}

	diagnostic := map[string]interface{}{}

	if input.Body != nil && input.Body.Bytes != nil {
		diagnostic["body_bytes"] = input.Body.Bytes
	}

	if input.Headers != nil {
		diagnostic["headers_to_log"] = set.FromStringSlice(*input.Headers)
	}

	diagnostic["data_masking"] = flattenApiManagementDataMasking(input.DataMasking)

	result = append(result, diagnostic)

	return result
}

func schemaApiManagementDataMaskingEntityList() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"mode": {
					Type:     pluginsdk.TypeString,
					Required: true,
					ValidateFunc: validation.StringInSlice([]string{
						string(apimanagement.DataMaskingModeHide),
						string(apimanagement.DataMaskingModeMask),
					}, false),
				},

				"value": {
					Type:         pluginsdk.TypeString,
					Required:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
		},
	}
}

func expandApiManagementDataMasking(input []interface{}) *apimanagement.DataMasking {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	inputRaw := input[0].(map[string]interface{})
	return &apimanagement.DataMasking{
		QueryParams: expandApiManagementDataMaskingEntityList(inputRaw["query_params"].([]interface{})),
		Headers:     expandApiManagementDataMaskingEntityList(inputRaw["headers"].([]interface{})),
	}
}

func expandApiManagementDataMaskingEntityList(input []interface{}) *[]apimanagement.DataMaskingEntity {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	result := make([]apimanagement.DataMaskingEntity, 0)
	for _, v := range input {
		entity := v.(map[string]interface{})
		result = append(result, apimanagement.DataMaskingEntity{
			Mode:  apimanagement.DataMaskingMode(entity["mode"].(string)),
			Value: utils.String(entity["value"].(string)),
		})
	}
	return &result
}

func flattenApiManagementDataMasking(dataMasking *apimanagement.DataMasking) []interface{} {
	if dataMasking == nil {
		return []interface{}{}
	}

	var queryParams, headers []interface{}
	if dataMasking.QueryParams != nil {
		queryParams = flattenApiManagementDataMaskingEntityList(dataMasking.QueryParams)
	}
	if dataMasking.Headers != nil {
		headers = flattenApiManagementDataMaskingEntityList(dataMasking.Headers)
	}

	return []interface{}{
		map[string]interface{}{
			"query_params": queryParams,
			"headers":      headers,
		},
	}
}

func flattenApiManagementDataMaskingEntityList(dataMaskingList *[]apimanagement.DataMaskingEntity) []interface{} {
	if dataMaskingList == nil || len(*dataMaskingList) == 0 {
		return []interface{}{}
	}

	result := []interface{}{}

	for _, entity := range *dataMaskingList {
		var value string
		if entity.Value != nil {
			value = *entity.Value
		}
		result = append(result, map[string]interface{}{
			"mode":  string(entity.Mode),
			"value": value,
		})
	}

	return result
}
