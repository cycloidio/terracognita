package managedapis

type ApiResourceProperties struct {
	ApiDefinitionUrl     *string                         `json:"apiDefinitionUrl,omitempty"`
	ApiDefinitions       *ApiResourceDefinitions         `json:"apiDefinitions,omitempty"`
	BackendService       *ApiResourceBackendService      `json:"backendService,omitempty"`
	Capabilities         *[]string                       `json:"capabilities,omitempty"`
	ConnectionParameters *map[string]ConnectionParameter `json:"connectionParameters,omitempty"`
	GeneralInformation   *ApiResourceGeneralInformation  `json:"generalInformation,omitempty"`
	Metadata             *ApiResourceMetadata            `json:"metadata,omitempty"`
	Name                 *string                         `json:"name,omitempty"`
	Policies             *ApiResourcePolicies            `json:"policies,omitempty"`
	RuntimeUrls          *[]string                       `json:"runtimeUrls,omitempty"`
	Swagger              *interface{}                    `json:"swagger,omitempty"`
}
