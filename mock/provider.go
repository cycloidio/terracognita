// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cycloidio/terracognita/provider (interfaces: Provider)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	filter "github.com/cycloidio/terracognita/filter"
	provider "github.com/cycloidio/terracognita/provider"
	gomock "github.com/golang/mock/gomock"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider is a mock of Provider interface
type Provider struct {
	ctrl     *gomock.Controller
	recorder *ProviderMockRecorder
}

// ProviderMockRecorder is the mock recorder for Provider
type ProviderMockRecorder struct {
	mock *Provider
}

// NewProvider creates a new mock instance
func NewProvider(ctrl *gomock.Controller) *Provider {
	mock := &Provider{ctrl: ctrl}
	mock.recorder = &ProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *Provider) EXPECT() *ProviderMockRecorder {
	return m.recorder
}

// HasResourceType mocks base method
func (m *Provider) HasResourceType(arg0 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasResourceType", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasResourceType indicates an expected call of HasResourceType
func (mr *ProviderMockRecorder) HasResourceType(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasResourceType", reflect.TypeOf((*Provider)(nil).HasResourceType), arg0)
}

// Region mocks base method
func (m *Provider) Region() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Region")
	ret0, _ := ret[0].(string)
	return ret0
}

// Region indicates an expected call of Region
func (mr *ProviderMockRecorder) Region() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Region", reflect.TypeOf((*Provider)(nil).Region))
}

// ResourceTypes mocks base method
func (m *Provider) ResourceTypes() []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResourceTypes")
	ret0, _ := ret[0].([]string)
	return ret0
}

// ResourceTypes indicates an expected call of ResourceTypes
func (mr *ProviderMockRecorder) ResourceTypes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResourceTypes", reflect.TypeOf((*Provider)(nil).ResourceTypes))
}

// Resources mocks base method
func (m *Provider) Resources(arg0 context.Context, arg1 string, arg2 *filter.Filter) ([]provider.Resource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Resources", arg0, arg1, arg2)
	ret0, _ := ret[0].([]provider.Resource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Resources indicates an expected call of Resources
func (mr *ProviderMockRecorder) Resources(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Resources", reflect.TypeOf((*Provider)(nil).Resources), arg0, arg1, arg2)
}

// String mocks base method
func (m *Provider) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String
func (mr *ProviderMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*Provider)(nil).String))
}

// TFClient mocks base method
func (m *Provider) TFClient() interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TFClient")
	ret0, _ := ret[0].(interface{})
	return ret0
}

// TFClient indicates an expected call of TFClient
func (mr *ProviderMockRecorder) TFClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TFClient", reflect.TypeOf((*Provider)(nil).TFClient))
}

// TFProvider mocks base method
func (m *Provider) TFProvider() *schema.Provider {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TFProvider")
	ret0, _ := ret[0].(*schema.Provider)
	return ret0
}

// TFProvider indicates an expected call of TFProvider
func (mr *ProviderMockRecorder) TFProvider() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TFProvider", reflect.TypeOf((*Provider)(nil).TFProvider))
}

// TagKey mocks base method
func (m *Provider) TagKey() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TagKey")
	ret0, _ := ret[0].(string)
	return ret0
}

// TagKey indicates an expected call of TagKey
func (mr *ProviderMockRecorder) TagKey() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TagKey", reflect.TypeOf((*Provider)(nil).TagKey))
}
