package testlib

import "testing"

type Assert struct {
	testing.T
}

func (a *Assert)False(expected bool) {
	a.NEquals(expected, true)
}

func (a *Assert)True(expected bool) {
	a.Equals(expected, true)
}

func (a *Assert)Equals(expected, actual interface{}) {
	if expected != actual {
		a.Errorf("Wants %v, got %v", expected, actual)
	}
}

func (a *Assert)NEquals(expected, actual interface{}) {
	if expected == actual {
		a.Errorf("Wants %v, got %v", expected, actual)
	}
}
