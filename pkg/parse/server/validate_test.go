package server

import "testing"

func TestIsAllLowerA_Z(t *testing.T) {
	type test struct {
		s       string
		isValid bool
	}
	tests := []test{
		{"", false},
		{"a_z", true},
		{"a-z", true},
		{"A", false},
		{"-", true},
		{"_", true},
	}
	for _, tt := range tests {
		isValid := isAllLowerA_Z(tt.s)
		if tt.isValid != isValid {
			t.Errorf("failed validating lowerA_Z for string: %s, got: %t, wanted: %t", tt.s, isValid, tt.isValid)
		}
	}
}
