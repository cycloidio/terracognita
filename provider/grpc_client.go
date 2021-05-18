package provider

import (
	"context"
	"fmt"
	"path"
	"runtime"

	"github.com/hashicorp/go-cty/cty"
	ctyjson "github.com/hashicorp/go-cty/cty/json"
	"github.com/hashicorp/go-cty/cty/msgpack"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform/tfdiags"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCClient is an inmemory implementation of the TF GRPC
// This should implement the terraform/providers.Interface but
// TF is still on zclconf/go-cty and we need hashicorp/go-cty
// so it compiles
type GRPCClient struct {
	server   *schema.GRPCProviderServer
	provider *schema.Provider
}

// NewGRPCClient wraps the pv into a GRPCClient
func NewGRPCClient(pv *schema.Provider) *GRPCClient {
	sv := schema.NewGRPCProviderServer(pv)
	return &GRPCClient{
		server:   sv,
		provider: pv,
	}
}

// ReadResource reads the Resource from the Provider
func (c *GRPCClient) ReadResource(r ReadResourceRequest) (resp ReadResourceResponse) {
	resSchema := c.getResourceSchema(r.TypeName)

	mp, err := msgpack.Marshal(r.PriorState, resSchema.CoreConfigSchema().ImpliedType())
	if err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}

	protoReq := &tfprotov5.ReadResourceRequest{
		TypeName:     r.TypeName,
		CurrentState: &tfprotov5.DynamicValue{MsgPack: mp},
		Private:      r.Private,
	}

	protoResp, err := c.server.ReadResource(context.Background(), protoReq)
	if err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(grpcErr(err))
		return resp
	}
	for _, d := range protoResp.Diagnostics {
		resp.Diagnostics = resp.Diagnostics.Append(errors.New(d.Summary))
	}

	state, err := decodeDynamicValue(protoResp.NewState, resSchema.CoreConfigSchema().ImpliedType())
	if err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(err)
		return resp
	}
	resp.NewState = state
	resp.Private = protoResp.Private

	return resp
}

// ImportResourceState imports the state of the resource from the Provider
func (c *GRPCClient) ImportResourceState(r ImportResourceStateRequest) (resp ImportResourceStateResponse) {
	protoReq := &tfprotov5.ImportResourceStateRequest{
		TypeName: r.TypeName,
		ID:       r.ID,
	}

	protoResp, err := c.server.ImportResourceState(context.Background(), protoReq)
	if err != nil {
		resp.Diagnostics = resp.Diagnostics.Append(grpcErr(err))
		return resp
	}
	for _, d := range protoResp.Diagnostics {
		resp.Diagnostics = resp.Diagnostics.Append(errors.New(d.Summary))
	}

	for _, imported := range protoResp.ImportedResources {
		resource := ImportedResource{
			TypeName: imported.TypeName,
			Private:  imported.Private,
		}

		resSchema := c.getResourceSchema(resource.TypeName)
		state, err := decodeDynamicValue(imported.State, resSchema.CoreConfigSchema().ImpliedType())
		if err != nil {
			resp.Diagnostics = resp.Diagnostics.Append(err)
			return resp
		}
		resource.State = state
		resp.ImportedResources = append(resp.ImportedResources, resource)
	}

	return resp

}

// getResourceSchema is a helper to extract the schema for a resource, and
// panics if the schema is not available.
func (c *GRPCClient) getResourceSchema(name string) *schema.Resource {
	resSchema, ok := c.provider.ResourcesMap[name]
	if !ok {
		panic("unknown resource type " + name)
	}
	return resSchema
}

// Decode a DynamicValue from either the JSON or MsgPack encoding.
func decodeDynamicValue(v *tfprotov5.DynamicValue, ty cty.Type) (cty.Value, error) {
	// always return a valid value
	var err error
	res := cty.NullVal(ty)
	if v == nil {
		return res, nil
	}

	switch {
	case len(v.MsgPack) > 0:
		res, err = msgpack.Unmarshal(v.MsgPack, ty)
	case len(v.JSON) > 0:
		res, err = ctyjson.Unmarshal(v.JSON, ty)
	}
	return res, err
}

// grpcErr extracts some known error types and formats them into better
// representations for core. This must only be called from plugin methods.
// Since we don't use RPC status errors for the plugin protocol, these do not
// contain any useful details, and we can return some text that at least
// indicates the plugin call and possible error condition.
func grpcErr(err error) (diags tfdiags.Diagnostics) {
	if err == nil {
		return
	}

	// extract the method name from the caller.
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return diags.Append(err)
	}

	f := runtime.FuncForPC(pc)

	// Function names will contain the full import path. Take the last
	// segment, which will let users know which method was being called.
	_, requestName := path.Split(f.Name())

	// TODO: while this expands the error codes into somewhat better messages,
	// this still does not easily link the error to an actual user-recognizable
	// plugin. The grpc plugin does not know its configured name, and the
	// errors are in a list of diagnostics, making it hard for the caller to
	// annotate the returned errors.
	switch status.Code(err) {
	case codes.Unavailable:
		// This case is when the plugin has stopped running for some reason,
		// and is usually the result of a crash.
		diags = diags.Append(tfdiags.Sourceless(
			tfdiags.Error,
			"Plugin did not respond",
			fmt.Sprintf("The plugin encountered an error, and failed to respond to the %s call. "+
				"The plugin logs may contain more details.", requestName),
		))
	case codes.Canceled:
		diags = diags.Append(tfdiags.Sourceless(
			tfdiags.Error,
			"Request cancelled",
			fmt.Sprintf("The %s request was cancelled.", requestName),
		))
	case codes.Unimplemented:
		diags = diags.Append(tfdiags.Sourceless(
			tfdiags.Error,
			"Unsupported plugin method",
			fmt.Sprintf("The %s method is not supported by this plugin.", requestName),
		))
	default:
		diags = diags.Append(tfdiags.Sourceless(
			tfdiags.Error,
			"Plugin error",
			fmt.Sprintf("The plugin returned an unexpected error from %s: %v", requestName, err),
		))
	}
	return
}

// ReadResourceRequest is the request sent to Read the Resource
// copied from terraform/providers.ReadResourceRequest
type ReadResourceRequest struct {
	// TypeName is the name of the resource type being read.
	TypeName string

	// PriorState contains the previously saved state value for this resource.
	PriorState cty.Value

	// Private is an opaque blob that will be stored in state along with the
	// resource. It is intended only for interpretation by the provider itself.
	Private []byte

	// ProviderMeta is the configuration for the provider_meta block for the
	// module and provider this resource belongs to. Its use is defined by
	// each provider, and it should not be used without coordination with
	// HashiCorp. It is considered experimental and subject to change.
	ProviderMeta cty.Value
}

// ReadResourceResponse is the response from Reading the Resource
// copied from terraform/providers.ReadResourceResponse
type ReadResourceResponse struct {
	// NewState contains the current state of the resource.
	NewState cty.Value

	// Diagnostics contains any warnings or errors from the method call.
	Diagnostics tfdiags.Diagnostics

	// Private is an opaque blob that will be stored in state along with the
	// resource. It is intended only for interpretation by the provider itself.
	Private []byte
}

// ImportResourceStateRequest is the request sent to Import the Resource State
// copied from terraform/providers.ImportResourceStateRequest
type ImportResourceStateRequest struct {
	// TypeName is the name of the resource type to be imported.
	TypeName string

	// ID is a string with which the provider can identify the resource to be
	// imported.
	ID string
}

// ImportResourceStateResponse is the response from Importing the Resource State
// copied from terraform/providers.ImportResourceStateResponse
type ImportResourceStateResponse struct {
	// ImportedResources contains one or more state values related to the
	// imported resource. It is not required that these be complete, only that
	// there is enough identifying information for the provider to successfully
	// update the states in ReadResource.
	ImportedResources []ImportedResource

	// Diagnostics contains any warnings or errors from the method call.
	Diagnostics tfdiags.Diagnostics
}

// ImportedResource is the resource information to Import
// copied from terraform/providers.ImportedResource
type ImportedResource struct {
	// TypeName is the name of the resource type associated with the
	// returned state. It's possible for providers to import multiple related
	// types with a single import request.
	TypeName string

	// State is the state of the remote object being imported. This may not be
	// complete, but must contain enough information to uniquely identify the
	// resource.
	State cty.Value

	// Private is an opaque blob that will be stored in state along with the
	// resource. It is intended only for interpretation by the provider itself.
	Private []byte
}
