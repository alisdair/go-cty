package convert

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/apparentlymart/go-cty/cty"
)

func TestOut(t *testing.T) {
	tests := []struct {
		CtyValue   cty.Value
		TargetType reflect.Type
		Want       interface{}
	}{

		// Bool
		{
			CtyValue:   cty.True,
			TargetType: reflect.TypeOf(false),
			Want:       true,
		},
		{
			CtyValue:   cty.False,
			TargetType: reflect.TypeOf(false),
			Want:       false,
		},
		{
			CtyValue:   cty.True,
			TargetType: reflect.PtrTo(reflect.TypeOf(false)),
			Want:       testOutAssertPtrVal(true),
		},
		{
			CtyValue:   cty.NullVal(cty.Bool),
			TargetType: reflect.PtrTo(reflect.TypeOf(false)),
			Want:       (*bool)(nil),
		},

		// String
		{
			CtyValue:   cty.StringVal("hello"),
			TargetType: reflect.TypeOf(""),
			Want:       "hello",
		},
		{
			CtyValue:   cty.StringVal(""),
			TargetType: reflect.TypeOf(""),
			Want:       "",
		},
		{
			CtyValue:   cty.StringVal("hello"),
			TargetType: reflect.PtrTo(reflect.TypeOf("")),
			Want:       testOutAssertPtrVal("hello"),
		},
		{
			CtyValue:   cty.NullVal(cty.String),
			TargetType: reflect.PtrTo(reflect.TypeOf("")),
			Want:       (*string)(nil),
		},

		// Number
		{
			CtyValue:   cty.NumberIntVal(5),
			TargetType: reflect.TypeOf(int(0)),
			Want:       int(5),
		},
		{
			CtyValue:   cty.NumberIntVal(5),
			TargetType: reflect.TypeOf(int8(0)),
			Want:       int8(5),
		},
		{
			CtyValue:   cty.NumberIntVal(5),
			TargetType: reflect.TypeOf(int16(0)),
			Want:       int16(5),
		},
		{
			CtyValue:   cty.NumberIntVal(5),
			TargetType: reflect.TypeOf(int32(0)),
			Want:       int32(5),
		},
		{
			CtyValue:   cty.NumberIntVal(5),
			TargetType: reflect.TypeOf(int64(0)),
			Want:       int64(5),
		},
		{
			CtyValue:   cty.NumberIntVal(5),
			TargetType: reflect.TypeOf(uint(0)),
			Want:       uint(5),
		},
		{
			CtyValue:   cty.NumberIntVal(5),
			TargetType: reflect.TypeOf(uint8(0)),
			Want:       uint8(5),
		},
		{
			CtyValue:   cty.NumberIntVal(5),
			TargetType: reflect.TypeOf(uint16(0)),
			Want:       uint16(5),
		},
		{
			CtyValue:   cty.NumberIntVal(5),
			TargetType: reflect.TypeOf(uint32(0)),
			Want:       uint32(5),
		},
		{
			CtyValue:   cty.NumberIntVal(5),
			TargetType: reflect.TypeOf(uint64(0)),
			Want:       uint64(5),
		},
		{
			CtyValue:   cty.NumberFloatVal(1.5),
			TargetType: reflect.TypeOf(float32(0)),
			Want:       float32(1.5),
		},
		{
			CtyValue:   cty.NumberFloatVal(1.5),
			TargetType: reflect.TypeOf(float64(0)),
			Want:       float64(1.5),
		},
		{
			CtyValue:   cty.NumberFloatVal(1.5),
			TargetType: reflect.PtrTo(bigFloatType),
			Want:       big.NewFloat(1.5),
		},
		{
			CtyValue:   cty.NumberIntVal(5),
			TargetType: reflect.PtrTo(bigIntType),
			Want:       big.NewInt(5),
		},

		// Passthrough
		{
			CtyValue:   cty.NumberIntVal(2),
			TargetType: valueType,
			Want:       cty.NumberIntVal(2),
		},
		{
			CtyValue:   cty.UnknownVal(cty.Bool),
			TargetType: valueType,
			Want:       cty.UnknownVal(cty.Bool),
		},
		{
			CtyValue:   cty.NullVal(cty.Bool),
			TargetType: valueType,
			Want:       cty.NullVal(cty.Bool),
		},
		{
			CtyValue:   cty.DynamicVal,
			TargetType: valueType,
			Want:       cty.DynamicVal,
		},
		{
			CtyValue:   cty.NullVal(cty.DynamicPseudoType),
			TargetType: valueType,
			Want:       cty.NullVal(cty.DynamicPseudoType),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%#v into %s", test.CtyValue, test.TargetType), func(t *testing.T) {
			target := reflect.New(test.TargetType)
			err := FromCtyValue(test.CtyValue, target.Interface())
			if err != nil {
				t.Fatalf("FromCtyValue returned error: %s", err)
			}

			got := target.Elem().Interface()

			if assertFunc, ok := test.Want.(testOutAssertFunc); ok {
				assertFunc(test.CtyValue, test.TargetType, got, t)
			} else if wantV, ok := test.Want.(cty.Value); ok {
				if gotV, ok := got.(cty.Value); ok {
					if !gotV.RawEquals(wantV) {
						testOutWrongResult(test.CtyValue, test.TargetType, got, test.Want, t)
					}
				} else {
					testOutWrongResult(test.CtyValue, test.TargetType, got, test.Want, t)
				}
			} else {
				if !reflect.DeepEqual(got, test.Want) {
					testOutWrongResult(test.CtyValue, test.TargetType, got, test.Want, t)
				}
			}
		})
	}
}

type testOutAssertFunc func(cty.Value, reflect.Type, interface{}, *testing.T)

func testOutAssertPtrVal(want interface{}) testOutAssertFunc {
	return func(ctyValue cty.Value, targetType reflect.Type, gotPtr interface{}, t *testing.T) {
		wantVal := reflect.ValueOf(want)
		gotVal := reflect.ValueOf(gotPtr)

		if gotVal.Kind() != reflect.Ptr {
			t.Fatalf("wrong type %s; want pointer to %T", gotVal.Type(), want)
		}
		gotVal = gotVal.Elem()

		want := wantVal.Interface()
		got := gotVal.Interface()
		if got != want {
			testOutWrongResult(
				ctyValue,
				targetType,
				got,
				want,
				t,
			)
		}
	}
}

func testOutWrongResult(ctyValue cty.Value, targetType reflect.Type, got interface{}, want interface{}, t *testing.T) {
	t.Errorf("wrong result\ninput:       %#v\ntarget type: %s\ngot:         %#v\nwant:        %#v", ctyValue, targetType, got, want)
}
