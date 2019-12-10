// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/readr-media/readr-restful/pkg/subscription (interfaces: Subscriber)

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	subscription "github.com/readr-media/readr-restful/pkg/subscription"
	reflect "reflect"
)

// MockSubscriber is a mock of Subscriber interface
type MockSubscriber struct {
	ctrl     *gomock.Controller
	recorder *MockSubscriberMockRecorder
}

// MockSubscriberMockRecorder is the mock recorder for MockSubscriber
type MockSubscriberMockRecorder struct {
	mock *MockSubscriber
}

// NewMockSubscriber creates a new mock instance
func NewMockSubscriber(ctrl *gomock.Controller) *MockSubscriber {
	mock := &MockSubscriber{ctrl: ctrl}
	mock.recorder = &MockSubscriberMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSubscriber) EXPECT() *MockSubscriberMockRecorder {
	return m.recorder
}

// CreateSubscription mocks base method
func (m *MockSubscriber) CreateSubscription(arg0 subscription.Subscription) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSubscription", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateSubscription indicates an expected call of CreateSubscription
func (mr *MockSubscriberMockRecorder) CreateSubscription(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSubscription", reflect.TypeOf((*MockSubscriber)(nil).CreateSubscription), arg0)
}

// GetSubscriptions mocks base method
func (m *MockSubscriber) GetSubscriptions() ([]subscription.Subscription, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSubscriptions")
	ret0, _ := ret[0].([]subscription.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSubscriptions indicates an expected call of GetSubscriptions
func (mr *MockSubscriberMockRecorder) GetSubscriptions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSubscriptions", reflect.TypeOf((*MockSubscriber)(nil).GetSubscriptions))
}
