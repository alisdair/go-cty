package stdlib

import (
	"fmt"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestHasIndex(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Key        cty.Value
		Want       cty.Value
	}{
		{
			cty.ListValEmpty(cty.Number),
			cty.NumberIntVal(2),
			cty.False,
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.StringVal("hello"),
			cty.False,
		},
		{
			cty.MapValEmpty(cty.Bool),
			cty.StringVal("hello"),
			cty.False,
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.True}),
			cty.StringVal("hello"),
			cty.True,
		},
		{
			cty.EmptyTupleVal,
			cty.StringVal("hello"),
			cty.False,
		},
		{
			cty.EmptyTupleVal,
			cty.NumberIntVal(0),
			cty.False,
		},
		{
			cty.TupleVal([]cty.Value{cty.True}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.UnknownVal(cty.List(cty.Bool)),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.UnknownVal(cty.Bool),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("HasIndex(%#v,%#v)", test.Collection, test.Key), func(t *testing.T) {
			got, err := HasIndex(test.Collection, test.Key)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		Values []cty.Value
		Want   cty.Value
		Err    bool
	}{
		{
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("b"),
				}),
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("b"),
				"c": cty.StringVal("d"),
			}),
			false,
		},
		{ // handle unknowns
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.UnknownVal(cty.String),
				}),
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.UnknownVal(cty.String),
				"c": cty.StringVal("d"),
			}),
			false,
		},
		{ // handle null map
			[]cty.Value{
				cty.NullVal(cty.Map(cty.String)),
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"c": cty.StringVal("d"),
			}),
			false,
		},
		{ // handle null map
			[]cty.Value{
				cty.NullVal(cty.Map(cty.String)),
				cty.NullVal(cty.Object(map[string]cty.Type{
					"a": cty.List(cty.String),
				})),
			},
			cty.NullVal(cty.EmptyObject),
			false,
		},
		{ // handle null object
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
				cty.NullVal(cty.Object(map[string]cty.Type{
					"a": cty.List(cty.String),
				})),
			},
			cty.ObjectVal(map[string]cty.Value{
				"c": cty.StringVal("d"),
			}),
			false,
		},
		{ // handle unknowns
			[]cty.Value{
				cty.UnknownVal(cty.Map(cty.String)),
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
			},
			cty.UnknownVal(cty.Map(cty.String)),
			false,
		},
		{ // handle dynamic unknown
			[]cty.Value{
				cty.UnknownVal(cty.DynamicPseudoType),
				cty.MapVal(map[string]cty.Value{
					"c": cty.StringVal("d"),
				}),
			},
			cty.DynamicVal,
			false,
		},
		{ // merge with conflicts is ok, last in wins
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("b"),
					"c": cty.StringVal("d"),
				}),
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("x"),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.StringVal("x"),
				"c": cty.StringVal("d"),
			}),
			false,
		},
		{ // only accept maps
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("b"),
					"c": cty.StringVal("d"),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("a"),
					cty.StringVal("x"),
				}),
			},
			cty.NilVal,
			true,
		},

		{ // argument error, for a null type
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.StringVal("b"),
				}),
				cty.NullVal(cty.String),
			},
			cty.NilVal,
			true,
		},
		{ // merge maps of maps
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.MapVal(map[string]cty.Value{
						"b": cty.StringVal("c"),
					}),
				}),
				cty.MapVal(map[string]cty.Value{
					"d": cty.MapVal(map[string]cty.Value{
						"e": cty.StringVal("f"),
					}),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.MapVal(map[string]cty.Value{
					"b": cty.StringVal("c"),
				}),
				"d": cty.MapVal(map[string]cty.Value{
					"e": cty.StringVal("f"),
				}),
			}),
			false,
		},
		{ // map of lists
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
						cty.StringVal("c"),
					}),
				}),
				cty.MapVal(map[string]cty.Value{
					"d": cty.ListVal([]cty.Value{
						cty.StringVal("e"),
						cty.StringVal("f"),
					}),
				}),
			},
			cty.MapVal(map[string]cty.Value{
				"a": cty.ListVal([]cty.Value{
					cty.StringVal("b"),
					cty.StringVal("c"),
				}),
				"d": cty.ListVal([]cty.Value{
					cty.StringVal("e"),
					cty.StringVal("f"),
				}),
			}),
			false,
		},
		{ // merge map of various kinds
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
						cty.StringVal("c"),
					}),
				}),
				cty.MapVal(map[string]cty.Value{
					"d": cty.MapVal(map[string]cty.Value{
						"e": cty.StringVal("f"),
					}),
				}),
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.ListVal([]cty.Value{
					cty.StringVal("b"),
					cty.StringVal("c"),
				}),
				"d": cty.MapVal(map[string]cty.Value{
					"e": cty.StringVal("f"),
				}),
			}),
			false,
		},
		{ // merge objects of various shapes
			[]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
					}),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"d": cty.DynamicVal,
				}),
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.ListVal([]cty.Value{
					cty.StringVal("b"),
				}),
				"d": cty.DynamicVal,
			}),
			false,
		},
		{ // merge maps and objects
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
					}),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"d": cty.NumberIntVal(2),
				}),
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.ListVal([]cty.Value{
					cty.StringVal("b"),
				}),
				"d": cty.NumberIntVal(2),
			}),
			false,
		},
		{ // attr a type and value is overridden
			[]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
					}),
					"b": cty.StringVal("b"),
				}),
				cty.ObjectVal(map[string]cty.Value{
					"a": cty.ObjectVal(map[string]cty.Value{
						"e": cty.StringVal("f"),
					}),
				}),
			},
			cty.ObjectVal(map[string]cty.Value{
				"a": cty.ObjectVal(map[string]cty.Value{
					"e": cty.StringVal("f"),
				}),
				"b": cty.StringVal("b"),
			}),
			false,
		},
		{ // argument error: non map type
			[]cty.Value{
				cty.MapVal(map[string]cty.Value{
					"a": cty.ListVal([]cty.Value{
						cty.StringVal("b"),
						cty.StringVal("c"),
					}),
				}),
				cty.ListVal([]cty.Value{
					cty.StringVal("d"),
					cty.StringVal("e"),
				}),
			},
			cty.NilVal,
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("merge(%#v)", test.Values), func(t *testing.T) {
			got, err := Merge(test.Values...)

			if test.Err {
				if err == nil {
					t.Fatal("succeeded; want error")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestIndex(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Key        cty.Value
		Want       cty.Value
	}{
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.True}),
			cty.StringVal("hello"),
			cty.True,
		},
		{
			cty.TupleVal([]cty.Value{cty.True, cty.StringVal("hello")}),
			cty.NumberIntVal(0),
			cty.True,
		},
		{
			cty.TupleVal([]cty.Value{cty.True, cty.StringVal("hello")}),
			cty.NumberIntVal(1),
			cty.StringVal("hello"),
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.UnknownVal(cty.List(cty.Bool)),
			cty.UnknownVal(cty.Number),
			cty.UnknownVal(cty.Bool),
		},
		{
			cty.ListValEmpty(cty.Number),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.MapValEmpty(cty.Number),
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.StringVal("hello"),
			cty.DynamicVal,
		},
		{
			cty.DynamicVal,
			cty.DynamicVal,
			cty.DynamicVal,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Index(%#v,%#v)", test.Collection, test.Key), func(t *testing.T) {
			got, err := Index(test.Collection, test.Key)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Want       cty.Value
	}{
		{
			cty.ListValEmpty(cty.Number),
			cty.NumberIntVal(0),
		},
		{
			cty.ListVal([]cty.Value{cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.SetValEmpty(cty.Number),
			cty.NumberIntVal(0),
		},
		{
			cty.SetVal([]cty.Value{cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.MapValEmpty(cty.Bool),
			cty.NumberIntVal(0),
		},
		{
			cty.MapVal(map[string]cty.Value{"hello": cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.EmptyTupleVal,
			cty.NumberIntVal(0),
		},
		{
			cty.TupleVal([]cty.Value{cty.True}),
			cty.NumberIntVal(1),
		},
		{
			cty.UnknownVal(cty.List(cty.Bool)),
			cty.UnknownVal(cty.Number),
		},
		{
			cty.DynamicVal,
			cty.UnknownVal(cty.Number),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Length(%#v)", test.Collection), func(t *testing.T) {
			got, err := Length(test.Collection)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}

func TestLookup(t *testing.T) {
	tests := []struct {
		Collection cty.Value
		Key        cty.Value
		Default    cty.Value
		Want       cty.Value
	}{
		{
			cty.MapValEmpty(cty.String),
			cty.StringVal("baz"),
			cty.StringVal("foo"),
			cty.StringVal("foo"),
		},
		{
			cty.MapVal(map[string]cty.Value{
				"foo": cty.StringVal("bar"),
			}),
			cty.StringVal("foo"),
			cty.StringVal("nope"),
			cty.StringVal("bar"),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Lookup(%#v,%#v,%#v)", test.Collection, test.Key, test.Default), func(t *testing.T) {
			got, err := Lookup(test.Collection, test.Key, test.Default)

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if !got.RawEquals(test.Want) {
				t.Errorf("wrong result\ngot:  %#v\nwant: %#v", got, test.Want)
			}
		})
	}
}
