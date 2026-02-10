package generate

import (
	"slices"
	"strings"

	repr "github.com/simplicity-load/apispec/pkg/repr/http"
)

type importSet[T setValue] struct {
	m map[string]string
	_ T
}

func (iSet importSet[unsorted]) add(imp string) {
	iSet.m[imp] = ""
}

type importer struct {
	Import string
	Ident  string
}

func newImportSet() importSet[unsorted] {
	return importSet[unsorted]{m: make(map[string]string)}
}

func (iSet importSet[unsorted]) sort() (importSet[sorted], []importer) {
	imports := make([]string, 0, len(iSet.m))
	for imp := range iSet.m {
		imports = append(imports, imp)
	}

	slices.Sort(imports)

	importers := make([]importer, len(imports))
	for idx, imp := range imports {
		ident := importIdent(uint(idx))

		iSet.m[imp] = ident
		importers[idx] = importer{
			Import: imp,
			Ident:  ident,
		}
	}
	return importSet[sorted]{m: iSet.m}, importers
}

func (iSet importSet[sorted]) get(imp string) string {
	return iSet.m[imp]
}

func importIdent(count uint) string {
	return intToIdent(count) + "i"
}

func intToIdent(x uint) (ident string) {
	r := x % 26
	m := x / 26
	for range m {
		ident += "z"
	}
	return string(rune('a'+r)) + ident
}

type recieverKey struct {
	importPath string
	name       string
	pointer    bool
}

type sorted struct{}
type unsorted struct{}

type setValue interface {
	sorted | unsorted
}

type recieverSet[T setValue] struct {
	m map[recieverKey]string
	_ T
}

func newRecieverSet() recieverSet[unsorted] {
	return recieverSet[unsorted]{m: make(map[recieverKey]string)}
}
func recieverIdent(count uint) string {
	return intToIdent(count) + "r"
}

func (rSet recieverSet[unsorted]) add(recv *repr.Reciever, importPath string) {
	rSet.m[recieverKey{
		importPath: importPath,
		name:       recv.Name,
		pointer:    recv.Pointer}] = ""
}

type reciever struct {
	repr.Reciever
	Import string
	Ident  string
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (rSet recieverSet[unsorted]) sort() (
	recieverSet[sorted],
	[]reciever,
) {
	keys := make([]recieverKey, 0, len(rSet.m))
	for key := range rSet.m {
		keys = append(keys, key)
	}

	slices.SortFunc(keys, func(a, b recieverKey) int {
		if cmp := strings.Compare(a.importPath, b.importPath); cmp != 0 {
			return cmp
		}
		if cmp := strings.Compare(a.name, b.name); cmp != 0 {
			return cmp
		}
		return boolInt(a.pointer) - boolInt(b.pointer)
	})

	recievers := make([]reciever, len(keys))
	for i, key := range keys {
		ident := recieverIdent(uint(i))

		rSet.m[key] = ident
		recievers[i] = reciever{
			Import: key.importPath,
			Reciever: repr.Reciever{
				Name:    key.name,
				Pointer: key.pointer,
			},
			Ident: ident,
		}
	}
	return recieverSet[sorted]{m: rSet.m}, recievers
}

func (rSet recieverSet[sorted]) get(recv *repr.Reciever, importPath string) string {
	return rSet.m[recieverKey{
		importPath: importPath,
		name:       recv.Name,
		pointer:    recv.Pointer}]
}
