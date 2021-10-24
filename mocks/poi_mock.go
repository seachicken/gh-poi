// Code generated by MockGen. DO NOT EDIT.
// Source: poi.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockConnection is a mock of Connection interface.
type MockConnection struct {
	ctrl     *gomock.Controller
	recorder *MockConnectionMockRecorder
}

// MockConnectionMockRecorder is the mock recorder for MockConnection.
type MockConnectionMockRecorder struct {
	mock *MockConnection
}

// NewMockConnection creates a new mock instance.
func NewMockConnection(ctrl *gomock.Controller) *MockConnection {
	mock := &MockConnection{ctrl: ctrl}
	mock.recorder = &MockConnectionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConnection) EXPECT() *MockConnectionMockRecorder {
	return m.recorder
}

// DeleteBranches mocks base method.
func (m *MockConnection) DeleteBranches(branchNames []string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteBranches", branchNames)
	ret0, _ := ret[0].(string)
	return ret0
}

// DeleteBranches indicates an expected call of DeleteBranches.
func (mr *MockConnectionMockRecorder) DeleteBranches(branchNames interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBranches", reflect.TypeOf((*MockConnection)(nil).DeleteBranches), branchNames)
}

// FetchPrStates mocks base method.
func (m *MockConnection) FetchPrStates(hostname string, repoNames []string, queryHashes string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchPrStates", hostname, repoNames, queryHashes)
	ret0, _ := ret[0].(string)
	return ret0
}

// FetchPrStates indicates an expected call of FetchPrStates.
func (mr *MockConnectionMockRecorder) FetchPrStates(hostname, repoNames, queryHashes interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchPrStates", reflect.TypeOf((*MockConnection)(nil).FetchPrStates), hostname, repoNames, queryHashes)
}

// FetchRepoNames mocks base method.
func (m *MockConnection) FetchRepoNames() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchRepoNames")
	ret0, _ := ret[0].(string)
	return ret0
}

// FetchRepoNames indicates an expected call of FetchRepoNames.
func (mr *MockConnectionMockRecorder) FetchRepoNames() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchRepoNames", reflect.TypeOf((*MockConnection)(nil).FetchRepoNames))
}

// GetBrancheNames mocks base method.
func (m *MockConnection) GetBrancheNames() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBrancheNames")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetBrancheNames indicates an expected call of GetBrancheNames.
func (mr *MockConnectionMockRecorder) GetBrancheNames() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBrancheNames", reflect.TypeOf((*MockConnection)(nil).GetBrancheNames))
}

// GetRemoteName mocks base method.
func (m *MockConnection) GetRemoteName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRemoteName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetRemoteName indicates an expected call of GetRemoteName.
func (mr *MockConnectionMockRecorder) GetRemoteName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRemoteName", reflect.TypeOf((*MockConnection)(nil).GetRemoteName))
}