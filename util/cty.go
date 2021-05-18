package util

import (
	"encoding/json"

	hcty "github.com/hashicorp/go-cty/cty"
	hmsgpack "github.com/hashicorp/go-cty/cty/msgpack"
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
		return zcty.EmptyObjectVal, err
	}
	ty, err := HashicorpToZclonfType(ht)
	if err != nil {
		return zcty.EmptyObjectVal, err
	}
	zvalue, err := zmsgpack.Unmarshal(sb, ty)
	if err != nil {
		return zcty.EmptyObjectVal, err
	}
	return zvalue, nil
}
