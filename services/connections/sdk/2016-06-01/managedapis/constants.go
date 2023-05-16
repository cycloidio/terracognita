package managedapis

import "strings"

type ApiType string

const (
	ApiTypeNotSpecified ApiType = "NotSpecified"
	ApiTypeRest         ApiType = "Rest"
	ApiTypeSoap         ApiType = "Soap"
)

func PossibleValuesForApiType() []string {
	return []string{
		string(ApiTypeNotSpecified),
		string(ApiTypeRest),
		string(ApiTypeSoap),
	}
}

func parseApiType(input string) (*ApiType, error) {
	vals := map[string]ApiType{
		"notspecified": ApiTypeNotSpecified,
		"rest":         ApiTypeRest,
		"soap":         ApiTypeSoap,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := ApiType(input)
	return &out, nil
}

type ConnectionParameterType string

const (
	ConnectionParameterTypeArray        ConnectionParameterType = "array"
	ConnectionParameterTypeBool         ConnectionParameterType = "bool"
	ConnectionParameterTypeConnection   ConnectionParameterType = "connection"
	ConnectionParameterTypeInt          ConnectionParameterType = "int"
	ConnectionParameterTypeOauthSetting ConnectionParameterType = "oauthSetting"
	ConnectionParameterTypeObject       ConnectionParameterType = "object"
	ConnectionParameterTypeSecureobject ConnectionParameterType = "secureobject"
	ConnectionParameterTypeSecurestring ConnectionParameterType = "securestring"
	ConnectionParameterTypeString       ConnectionParameterType = "string"
)

func PossibleValuesForConnectionParameterType() []string {
	return []string{
		string(ConnectionParameterTypeArray),
		string(ConnectionParameterTypeBool),
		string(ConnectionParameterTypeConnection),
		string(ConnectionParameterTypeInt),
		string(ConnectionParameterTypeOauthSetting),
		string(ConnectionParameterTypeObject),
		string(ConnectionParameterTypeSecureobject),
		string(ConnectionParameterTypeSecurestring),
		string(ConnectionParameterTypeString),
	}
}

func parseConnectionParameterType(input string) (*ConnectionParameterType, error) {
	vals := map[string]ConnectionParameterType{
		"array":        ConnectionParameterTypeArray,
		"bool":         ConnectionParameterTypeBool,
		"connection":   ConnectionParameterTypeConnection,
		"int":          ConnectionParameterTypeInt,
		"oauthsetting": ConnectionParameterTypeOauthSetting,
		"object":       ConnectionParameterTypeObject,
		"secureobject": ConnectionParameterTypeSecureobject,
		"securestring": ConnectionParameterTypeSecurestring,
		"string":       ConnectionParameterTypeString,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := ConnectionParameterType(input)
	return &out, nil
}

type WsdlImportMethod string

const (
	WsdlImportMethodNotSpecified    WsdlImportMethod = "NotSpecified"
	WsdlImportMethodSoapPassThrough WsdlImportMethod = "SoapPassThrough"
	WsdlImportMethodSoapToRest      WsdlImportMethod = "SoapToRest"
)

func PossibleValuesForWsdlImportMethod() []string {
	return []string{
		string(WsdlImportMethodNotSpecified),
		string(WsdlImportMethodSoapPassThrough),
		string(WsdlImportMethodSoapToRest),
	}
}

func parseWsdlImportMethod(input string) (*WsdlImportMethod, error) {
	vals := map[string]WsdlImportMethod{
		"notspecified":    WsdlImportMethodNotSpecified,
		"soappassthrough": WsdlImportMethodSoapPassThrough,
		"soaptorest":      WsdlImportMethodSoapToRest,
	}
	if v, ok := vals[strings.ToLower(input)]; ok {
		return &v, nil
	}

	// otherwise presume it's an undefined value and best-effort it
	out := WsdlImportMethod(input)
	return &out, nil
}
