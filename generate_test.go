package apispec

import (
	"fmt"
	"testing"
)

func TestIntToIdentIter(t *testing.T) {
	for i := range 100 {
		fmt.Printf("%d : %s\n", i, intToIdent(uint(i)))
	}
	t.Error("asjkldasldj")
}
