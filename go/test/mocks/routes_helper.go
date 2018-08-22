// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/s12chung/gostatic/go/content/routes (interfaces: Helper)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	atom "github.com/s12chung/gostatic/go/lib/atom"
	goodreads "github.com/s12chung/gostatic/go/lib/goodreads"
	router "github.com/s12chung/gostatic/go/lib/router"
	reflect "reflect"
)

// MockHelper is a mock of Helper interface
type MockHelper struct {
	ctrl     *gomock.Controller
	recorder *MockHelperMockRecorder
}

// MockHelperMockRecorder is the mock recorder for MockHelper
type MockHelperMockRecorder struct {
	mock *MockHelper
}

// NewMockHelper creates a new mock instance
func NewMockHelper(ctrl *gomock.Controller) *MockHelper {
	mock := &MockHelper{ctrl: ctrl}
	mock.recorder = &MockHelperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHelper) EXPECT() *MockHelperMockRecorder {
	return m.recorder
}

// GoodreadsSettings mocks base method
func (m *MockHelper) GoodreadsSettings() *goodreads.Settings {
	ret := m.ctrl.Call(m, "GoodreadsSettings")
	ret0, _ := ret[0].(*goodreads.Settings)
	return ret0
}

// GoodreadsSettings indicates an expected call of GoodreadsSettings
func (mr *MockHelperMockRecorder) GoodreadsSettings() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GoodreadsSettings", reflect.TypeOf((*MockHelper)(nil).GoodreadsSettings))
}

// ManifestUrl mocks base method
func (m *MockHelper) ManifestUrl(arg0 string) string {
	ret := m.ctrl.Call(m, "ManifestUrl", arg0)
	ret0, _ := ret[0].(string)
	return ret0
}

// ManifestUrl indicates an expected call of ManifestUrl
func (mr *MockHelperMockRecorder) ManifestUrl(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ManifestUrl", reflect.TypeOf((*MockHelper)(nil).ManifestUrl), arg0)
}

// RespondAtom mocks base method
func (m *MockHelper) RespondAtom(arg0 router.Context, arg1, arg2 string, arg3 []*atom.HtmlEntry) error {
	ret := m.ctrl.Call(m, "RespondAtom", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// RespondAtom indicates an expected call of RespondAtom
func (mr *MockHelperMockRecorder) RespondAtom(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RespondAtom", reflect.TypeOf((*MockHelper)(nil).RespondAtom), arg0, arg1, arg2, arg3)
}

// RespondHTML mocks base method
func (m *MockHelper) RespondHTML(arg0 router.Context, arg1 string, arg2 interface{}) error {
	ret := m.ctrl.Call(m, "RespondHTML", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// RespondHTML indicates an expected call of RespondHTML
func (mr *MockHelperMockRecorder) RespondHTML(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RespondHTML", reflect.TypeOf((*MockHelper)(nil).RespondHTML), arg0, arg1, arg2)
}

// RespondUrlHTML mocks base method
func (m *MockHelper) RespondUrlHTML(arg0 router.Context, arg1 interface{}) error {
	ret := m.ctrl.Call(m, "RespondUrlHTML", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RespondUrlHTML indicates an expected call of RespondUrlHTML
func (mr *MockHelperMockRecorder) RespondUrlHTML(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RespondUrlHTML", reflect.TypeOf((*MockHelper)(nil).RespondUrlHTML), arg0, arg1)
}
