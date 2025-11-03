package http

import "iter"

func (ps PathStrings) NoRootPaths() iter.Seq[*PathString] {
	return func(yield func(x *PathString) bool) {
		for _, p := range ps {
			if p.Type == PathROOT {
				continue
			}
			if !yield(p) {
				return
			}
		}
	}
}

func (endpoints Endpoints) Imports() iter.Seq[string] {
	return func(yield func(x string) bool) {
		for _, endpoint := range endpoints {
			for _, imp := range []string{
				endpoint.Body.Import,
				// endpoint.Response.Import, NOTE not used
				endpoint.Handler.Import,
			} {
				if imp == "" {
					continue
				}
				if !yield(imp) {
					return
				}
			}
		}
	}
}

func (endpoints Endpoints) Recievers() iter.Seq2[Reciever, string] {
	return func(yield func(x Reciever, imp string) bool) {
		for _, endpoint := range endpoints {
			method := endpoint.Handler.Reciever
			if method == nil {
				continue
			}
			if !yield(*method, endpoint.Handler.Import) {
				return
			}
		}
	}
}
