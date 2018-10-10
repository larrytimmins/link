// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/coreos/etcd/clientv3 (interfaces: Txn)

// Package etcdmock is a generated GoMock package.
package etcdmock

import (
	reflect "reflect"

	clientv3 "github.com/coreos/etcd/clientv3"
	gomock "github.com/golang/mock/gomock"
)

// MockTxn is a mock of Txn interface
type MockTxn struct {
	ctrl     *gomock.Controller
	recorder *MockTxnMockRecorder
}

// MockTxnMockRecorder is the mock recorder for MockTxn
type MockTxnMockRecorder struct {
	mock *MockTxn
}

// NewMockTxn creates a new mock instance
func NewMockTxn(ctrl *gomock.Controller) *MockTxn {
	mock := &MockTxn{ctrl: ctrl}
	mock.recorder = &MockTxnMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTxn) EXPECT() *MockTxnMockRecorder {
	return m.recorder
}

// Commit mocks base method
func (m *MockTxn) Commit() (*clientv3.TxnResponse, error) {
	ret := m.ctrl.Call(m, "Commit")
	ret0, _ := ret[0].(*clientv3.TxnResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Commit indicates an expected call of Commit
func (mr *MockTxnMockRecorder) Commit() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockTxn)(nil).Commit))
}

// Else mocks base method
func (m *MockTxn) Else(arg0 ...clientv3.Op) clientv3.Txn {
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Else", varargs...)
	ret0, _ := ret[0].(clientv3.Txn)
	return ret0
}

// Else indicates an expected call of Else
func (mr *MockTxnMockRecorder) Else(arg0 ...interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Else", reflect.TypeOf((*MockTxn)(nil).Else), arg0...)
}

// If mocks base method
func (m *MockTxn) If(arg0 ...clientv3.Cmp) clientv3.Txn {
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "If", varargs...)
	ret0, _ := ret[0].(clientv3.Txn)
	return ret0
}

// If indicates an expected call of If
func (mr *MockTxnMockRecorder) If(arg0 ...interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "If", reflect.TypeOf((*MockTxn)(nil).If), arg0...)
}

// Then mocks base method
func (m *MockTxn) Then(arg0 ...clientv3.Op) clientv3.Txn {
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Then", varargs...)
	ret0, _ := ret[0].(clientv3.Txn)
	return ret0
}

// Then indicates an expected call of Then
func (mr *MockTxnMockRecorder) Then(arg0 ...interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Then", reflect.TypeOf((*MockTxn)(nil).Then), arg0...)
}