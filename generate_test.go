package apispec

import (
	"testing"
)

func TestIntToIdentIter(t *testing.T) {
	tests := []struct {
		input    uint
		expected string
	}{
		{0, "a"},
		{1, "b"},
		{25, "z"},
		{26, "az"},
		{27, "bz"},
		{51, "zz"},
		{52, "azz"},
	}

	for _, tt := range tests {
		result := intToIdent(tt.input)
		if result != tt.expected {
			t.Errorf("intToIdent(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}
