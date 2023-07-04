package util

import (
	"encoding/json"

	hcty "github.com/hashicorp/go-cty/cty"
	hctyjson "github.com/hashicorp/go-cty/cty/json"
	hmsgpack "github.com/hashicorp/go-cty/cty/msgpack"
	"github.com/pkg/errors"
	zcty "github.com/zclconf/go-cty/cty"
	zmsgpack "github.com/zclconf/go-cty/cty/msgpack"
)

// HashicorpToZclonfType converts from Hashicoprt.Type to zclconf.Type
func HashicorpToZclonfType(ht hcty.Type) (zcty.Type, error) {
	tb, err := json.Marshal(ht)
	if err != nil {
		return zcty.EmptyObject, err
	}
	var ty zcty.Type
	err = json.Unmarshal(tb, &ty)
	if err != nil {
		return zcty.EmptyObject, err
	}
	return ty, nil
}

// HashicorpToZclonfValue converts from Hashicoprt.Value to zclconf.Value
func HashicorpToZclonfValue(hv hcty.Value, ht hcty.Type) (zcty.Value, error) {
	sb, err := hmsgpack.Marshal(hv, ht)
	if err != nil {
		return zcty.EmptyObjectVal, errors.Wrapf(err, "failed to Hashicorp marshal")
	}
	ty, err := HashicorpToZclonfType(ht)
	if err != nil {
		return zcty.EmptyObjectVal, errors.Wrapf(err, "failed to convert from Hashiciprt to Zclon")
	}
	zvalue, err := zmsgpack.Unmarshal(sb, ty)
	if err != nil {
		return zcty.EmptyObjectVal, errors.Wrapf(err, "failed to Zclon unmarshal")
	}
	return zvalue, nil
}

// CtyObjectToUnstructured converts a Terraform specific cty.Object type manifest
// into a dynamic client specific unstructured object
func CtyObjectToUnstructured(in *hcty.Value) (map[string]interface{}, error) {
	simple := &hctyjson.SimpleJSONValue{Value: *in}
	jsonVal, err := simple.MarshalJSON()
	if err != nil {
		return nil, err
	}
	udata := map[string]interface{}{}
	err = json.Unmarshal(jsonVal, &udata)
	if err != nil {
		return nil, err
	}
	return udata, nil
}

// UnstructuredToCty converts a dynamic client specific unstructured object
// into a Terraform specific cty.Object type manifest
func UnstructuredToCty(in map[string]interface{}) (hcty.Value, error) {
	jsonVal, err := json.Marshal(in)
	if err != nil {
		return hcty.NilVal, errors.Wrapf(err, "unable to marshal value")
	}
	simple := &hctyjson.SimpleJSONValue{}
	err = simple.UnmarshalJSON(jsonVal)
	if err != nil {
		return hcty.NilVal, errors.Wrapf(err, "unable to unmarshal to simple value")
	}
	return simple.Value, nil
}
