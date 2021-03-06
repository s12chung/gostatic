// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/s12chung/gostatic/go/app (interfaces: Setter)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	router "github.com/s12chung/gostatic/go/lib/router"
	reflect "reflect"
)

// MockSetter is a mock of Setter interface
type MockSetter struct {
	ctrl     *gomock.Controller
	recorder *MockSetterMockRecorder
}

// MockSetterMockRecorder is the mock recorder for MockSetter
type MockSetterMockRecorder struct {
	mock *MockSetter
}

// NewMockSetter creates a new mock instance
func NewMockSetter(ctrl *gomock.Controller) *MockSetter {
	mock := &MockSetter{ctrl: ctrl}
	mock.recorder = &MockSetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSetter) EXPECT() *MockSetterMockRecorder {
	return m.recorder
}

// AssetsURL mocks base method
func (m *MockSetter) AssetsURL() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AssetsURL")
	ret0, _ := ret[0].(string)
	return ret0
}

// AssetsURL indicates an expected call of AssetsURL
func (mr *MockSetterMockRecorder) AssetsURL() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AssetsURL", reflect.TypeOf((*MockSetter)(nil).AssetsURL))
}

// GeneratedAssetsPath mocks base method
func (m *MockSetter) GeneratedAssetsPath() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GeneratedAssetsPath")
	ret0, _ := ret[0].(string)
	return ret0
}

// GeneratedAssetsPath indicates an expected call of GeneratedAssetsPath
func (mr *MockSetterMockRecorder) GeneratedAssetsPath() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GeneratedAssetsPath", reflect.TypeOf((*MockSetter)(nil).GeneratedAssetsPath))
}

// SetRoutes mocks base method
func (m *MockSetter) SetRoutes(arg0 router.Router) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetRoutes", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetRoutes indicates an expected call of SetRoutes
func (mr *MockSetterMockRecorder) SetRoutes(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetRoutes", reflect.TypeOf((*MockSetter)(nil).SetRoutes), arg0)
}

// URLBatches mocks base method
func (m *MockSetter) URLBatches(arg0 router.Router) ([][]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "URLBatches", arg0)
	ret0, _ := ret[0].([][]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// URLBatches indicates an expected call of URLBatches
func (mr *MockSetterMockRecorder) URLBatches(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "URLBatches", reflect.TypeOf((*MockSetter)(nil).URLBatches), arg0)
}
