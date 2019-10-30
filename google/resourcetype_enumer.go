// Code generated by "enumer -type ResourceType -addprefix google_ -transform snake -linecomment"; DO NOT EDIT.

package google

import (
	"fmt"
)

const _ResourceTypeName = "google_compute_instancegoogle_compute_firewallgoogle_compute_network"

var _ResourceTypeIndex = [...]uint8{0, 23, 46, 68}

const _ResourceTypeLowerName = "google_compute_instancegoogle_compute_firewallgoogle_compute_network"

func (i ResourceType) String() string {
	if i < 0 || i >= ResourceType(len(_ResourceTypeIndex)-1) {
		return fmt.Sprintf("ResourceType(%d)", i)
	}
	return _ResourceTypeName[_ResourceTypeIndex[i]:_ResourceTypeIndex[i+1]]
}

var _ResourceTypeValues = []ResourceType{0, 1, 2}

var _ResourceTypeNameToValueMap = map[string]ResourceType{
	_ResourceTypeName[0:23]:       0,
	_ResourceTypeLowerName[0:23]:  0,
	_ResourceTypeName[23:46]:      1,
	_ResourceTypeLowerName[23:46]: 1,
	_ResourceTypeName[46:68]:      2,
	_ResourceTypeLowerName[46:68]: 2,
}

var _ResourceTypeNames = []string{
	_ResourceTypeName[0:23],
	_ResourceTypeName[23:46],
	_ResourceTypeName[46:68],
}

// ResourceTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func ResourceTypeString(s string) (ResourceType, error) {
	if val, ok := _ResourceTypeNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to ResourceType values", s)
}

// ResourceTypeValues returns all values of the enum
func ResourceTypeValues() []ResourceType {
	return _ResourceTypeValues
}

// ResourceTypeStrings returns a slice of all String values of the enum
func ResourceTypeStrings() []string {
	strs := make([]string, len(_ResourceTypeNames))
	copy(strs, _ResourceTypeNames)
	return strs
}

// IsAResourceType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i ResourceType) IsAResourceType() bool {
	for _, v := range _ResourceTypeValues {
		if i == v {
			return true
		}
	}
	return false
}
