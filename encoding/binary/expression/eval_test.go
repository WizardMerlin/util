// Copyright 2013 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package expression

import (
	"reflect"
	"testing"
)

func TestEval(t *testing.T) {
	type s struct {
		Something int
	}
	var str = reflect.ValueOf(struct {
		Length int
		Sub    s
	}{3, s{10}})
	var tests = []struct {
		in  string
		out int
	}{
		{"1", 1},
		{"1+2", 1 + 2},
		{"(3*4)+2", 3*4 + 2},
		{"Length", 3},
		{"Length+1", 4},
		{"Length-1", 2},
		{"(Length+3)&^3", 4},
		{"((Length-1)+3)&^3", 4},
		{"Length == 3", 1},
		{"Length == 4", 0},
		{"Length < 4", 1},
		{"Length > 3", 0},
		{"Length >= 3", 1},
		{"Sub.Something", 10},
		{"Sub.Something + Length", 13},
	}

	for i, test := range tests {
		var p EXPRESSION
		if !p.Parse(test.in) {
			t.Error(p.Error(), p.RootNode())
		} else {
			t.Log(p.RootNode())
		}
		if r, err := Eval(&str, p.RootNode()); err != nil {
			t.Error(err)
		} else if r != test.out {
			t.Errorf("%d: Expected %d, but got %d", i, test.out, r)
		}
	}
}
