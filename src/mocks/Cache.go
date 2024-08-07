// Code generated by mockery v2.44.1. DO NOT EDIT.

package mocks

import (
	agglog "github.com/kevin-hanselman/dud/src/agglog"
	artifact "github.com/kevin-hanselman/dud/src/artifact"

	mock "github.com/stretchr/testify/mock"

	progress "github.com/kevin-hanselman/dud/src/progress"

	strategy "github.com/kevin-hanselman/dud/src/strategy"
)

// Cache is an autogenerated mock type for the Cache type
type Cache struct {
	mock.Mock
}

// Checkout provides a mock function with given fields: workDir, art, s, p, l
func (_m *Cache) Checkout(workDir string, art artifact.Artifact, s strategy.CheckoutStrategy, p progress.Progress, l *agglog.AggLogger) error {
	ret := _m.Called(workDir, art, s, p, l)

	if len(ret) == 0 {
		panic("no return value specified for Checkout")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, artifact.Artifact, strategy.CheckoutStrategy, progress.Progress, *agglog.AggLogger) error); ok {
		r0 = rf(workDir, art, s, p, l)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Commit provides a mock function with given fields: workDir, art, s, l
func (_m *Cache) Commit(workDir string, art *artifact.Artifact, s strategy.CheckoutStrategy, l *agglog.AggLogger) error {
	ret := _m.Called(workDir, art, s, l)

	if len(ret) == 0 {
		panic("no return value specified for Commit")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *artifact.Artifact, strategy.CheckoutStrategy, *agglog.AggLogger) error); ok {
		r0 = rf(workDir, art, s, l)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Fetch provides a mock function with given fields: remoteSrc, arts, l
func (_m *Cache) Fetch(remoteSrc string, arts map[string]*artifact.Artifact, l *agglog.AggLogger) error {
	ret := _m.Called(remoteSrc, arts, l)

	if len(ret) == 0 {
		panic("no return value specified for Fetch")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, map[string]*artifact.Artifact, *agglog.AggLogger) error); ok {
		r0 = rf(remoteSrc, arts, l)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Push provides a mock function with given fields: remoteDst, arts, l
func (_m *Cache) Push(remoteDst string, arts map[string]*artifact.Artifact, l *agglog.AggLogger) error {
	ret := _m.Called(remoteDst, arts, l)

	if len(ret) == 0 {
		panic("no return value specified for Push")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, map[string]*artifact.Artifact, *agglog.AggLogger) error); ok {
		r0 = rf(remoteDst, arts, l)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Status provides a mock function with given fields: workDir, art, shortCircuit
func (_m *Cache) Status(workDir string, art artifact.Artifact, shortCircuit bool) (artifact.Status, error) {
	ret := _m.Called(workDir, art, shortCircuit)

	if len(ret) == 0 {
		panic("no return value specified for Status")
	}

	var r0 artifact.Status
	var r1 error
	if rf, ok := ret.Get(0).(func(string, artifact.Artifact, bool) (artifact.Status, error)); ok {
		return rf(workDir, art, shortCircuit)
	}
	if rf, ok := ret.Get(0).(func(string, artifact.Artifact, bool) artifact.Status); ok {
		r0 = rf(workDir, art, shortCircuit)
	} else {
		r0 = ret.Get(0).(artifact.Status)
	}

	if rf, ok := ret.Get(1).(func(string, artifact.Artifact, bool) error); ok {
		r1 = rf(workDir, art, shortCircuit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewCache creates a new instance of Cache. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCache(t interface {
	mock.TestingT
	Cleanup(func())
}) *Cache {
	mock := &Cache{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
