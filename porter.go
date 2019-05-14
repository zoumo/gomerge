/*
Copyright 2019 zoumo(jim.zoumo@gmail.com). All rights reserved

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gomerge

import (
	"fmt"
	"reflect"
)

var (
	goConvertion = reflect.ValueOf(func(dst, src interface{}, o *Options) (interface{}, error) {
		dstV := reflect.ValueOf(dst)
		srcV := reflect.ValueOf(src)
		if o.Overwrite || isEmptyValue(dstV) {
			// try to use default converter in reflect
			converted := srcV.Convert(dstV.Type())
			return converted.Interface(), nil
		}
		return dstV.Interface(), nil
	})
)

type pair struct {
	Dst reflect.Type
	Src reflect.Type
}

type porter struct {
	mergeFuncs   map[reflect.Type]reflect.Value
	convertFuncs map[pair]reflect.Value
}

func newPorter() *porter {
	return &porter{
		mergeFuncs:   map[reflect.Type]reflect.Value{},
		convertFuncs: map[pair]reflect.Value{},
	}
}

func (m *porter) defaultMerge(dst, src reflect.Value, o *Options) error {
	dstType := dst.Type()
	srcType := src.Type()
	if dstType != srcType {
		return m.convert(dst, src, o)
	}
	return m.deepMerge(dst, src, o)
}

func (m *porter) addCustomFuncs(fns ...interface{}) error {
	for _, fn := range fns {
		fv := reflect.ValueOf(fn)
		ft := fv.Type()
		convertion, err := verifyCustomMergeFunctionSignature(ft)
		if err != nil {
			return err
		}
		if convertion {
			m.convertFuncs[pair{ft.In(0), ft.In(1)}] = fv
			continue
		}
		m.mergeFuncs[ft.In(0)] = fv
	}
	return nil
}

func (m *porter) callCustom(custom, dstV, srcV reflect.Value, o *Options) (reflect.Value, error) {
	args := []reflect.Value{dstV, srcV, reflect.ValueOf(o)}
	rets := custom.Call(args)
	ret0 := rets[0]
	err := rets[1].Interface()
	if err == nil {
		return ret0, nil
	}
	return ret0, err.(error)
}

func (m *porter) converter(dst, src reflect.Type, o *Options) (reflect.Value, bool) {
	// find converter in factory
	convert, ok := m.convertFuncs[pair{dst, src}]
	if ok {
		return convert, true
	}
	if o.GoConvertion && convertible(dst, src) {
		return goConvertion, true
	}
	return reflect.Value{}, false
}

func (m *porter) convert(dst, src reflect.Value, o *Options) error {
	// deref
	dstType := dst.Type()
	srcType := src.Type()
	dstKind := dstType.Kind()
	srcKind := srcType.Kind()

	// get true type behind interface{}
	if dstKind == reflect.Interface {
		dstE := derefInterface(dst)
		dstType = dstE.Type()
		dstKind = dstE.Kind()
	}

	if srcKind == reflect.Interface {
		srcE := derefInterface(src)
		srcType = srcE.Type()
		srcKind = srcE.Kind()
	}

	if convert, ok := m.converter(dstType, srcType, o); ok {
		converted, err := m.callCustom(convert, dst, src, o)
		if err != nil {
			return err
		}
		return directMerge(dst, converted, o)
	}

	if dstKind == reflect.Ptr || srcKind == reflect.Ptr {
		// dereference dst and src, find the element type behind ptr
		// may be converter know how to convert int to string, but it don't know
		// how to convert *int to string
		dstEValue := derefPtr(dst)
		dstEType := dstEValue.Type()
		srcEValue := derefPtr(src)
		srcEType := srcEValue.Type()
		if dstEType != srcEType {
			if convert, ok := m.converter(dstEType, srcEType, o); ok {
				converted, err := m.callCustom(convert, dstEValue, srcEValue, o)
				if err != nil {
					return err
				}
				return directMerge(dstEValue, converted, o)
			}
			return fmt.Errorf("after dereference for <src %v, dst %v>, we still can't convert element from %v to %v type", srcType, dstType, srcEType, dstEType)
		}
		return m.deepMerge(dstEValue, srcEValue, o)
	}
	return fmt.Errorf("can not convert %v to %v", srcType, dstType)
}

func (m *porter) deepMerge(dst, src reflect.Value, o *Options) error {
	dstType := dst.Type()
	srcType := src.Type()

	if dstType != srcType {
		return fmt.Errorf("deepMerge: src %v and dst %v must be of same type", srcType, dstType)
	}

	if merge, ok := m.mergeFuncs[dstType]; ok {
		merged, err := m.callCustom(merge, dst, src, o)
		if err != nil {
			return err
		}
		return directMerge(dst, merged, o)
	}

	switch dst.Kind() {
	case reflect.Struct:
		if hasUnexportedField(dstType) {
			// if the struct contains unexported field, treating it as a single entity
			return directMerge(dst, src, o)
		}
		for i := 0; i < dst.NumField(); i++ {
			if err := m.deepMerge(dst.Field(i), src.Field(i), o); err != nil {
				return err
			}
		}
	case reflect.Map:
		if src.IsNil() {
			return nil
		}

		if dst.IsNil() {
			// avoid panic when SetMapIndex
			dst.Set(reflect.MakeMap(dstType))
		}
		for _, key := range src.MapKeys() {
			srcE := derefInterface(src.MapIndex(key))
			dstE := derefInterface(dst.MapIndex(key))
			if !dstE.IsValid() {
				// the key is not present in dst map, set it anyway
				dst.SetMapIndex(key, srcE)
				continue
			}

			srcEType := srcE.Type()
			dstEType := dstE.Type()
			isEmpty := isEmptyValue(dstE)

			switch dstEType.Kind() {
			case reflect.Struct, reflect.Ptr, reflect.Map:
			default:
				// some type can not be changed directly
				// slice can not be set yet
				dstE = reflect.New(dstEType).Elem()
			}

			if dstEType != srcEType {
				if err := m.convert(dstE, srcE, o); err != nil {
					return err
				}
			} else {
				if err := m.deepMerge(dstE, srcE, o); err != nil {
					return err
				}
			}
			switch dstEType.Kind() {
			case reflect.Struct, reflect.Ptr, reflect.Map:
			default:
				// set it directly
				if o.Overwrite || isEmpty {
					dst.SetMapIndex(key, dstE)
				}
			}

		}
	case reflect.Slice:
		if src.IsNil() {
			// skip
			return nil
		}
		switch o.SliceMode {
		case AppendSlice:
			return directMerge(dst, reflect.AppendSlice(dst, src), o)
		case UniteSlice:
			dstType := dst.Type()
			dstEType := dstType.Elem()
			if !hashable(dstEType) || dstEType.Kind() == reflect.Bool {
				// exclude bool because it makes no sense to unite two bool slice
				// fallthrough to use ApplenSlice
				return directMerge(dst, reflect.AppendSlice(dst, src), o)
			}
			mapType := reflect.MapOf(dstEType, reflect.TypeOf(true))
			mapValue := reflect.ValueOf(true)
			existed := reflect.MakeMap(mapType)
			newElem := []reflect.Value{}

			// get existed
			for i := 0; i < dst.Len(); i++ {
				key := dst.Index(i)
				existed.SetMapIndex(key, mapValue)
			}
			// get new element
			for i := 0; i < src.Len(); i++ {
				key := src.Index(i)
				result := existed.MapIndex(key)
				if result.IsValid() {
					// already exists
					continue
				}
				newElem = append(newElem, key)
			}
			// append new elements
			if len(newElem) > 0 {
				return directMerge(dst, reflect.Append(dst, newElem...), o)
			}
			return nil
		}
		return directMerge(dst, src, o)
	case reflect.Ptr:
		if src.IsNil() {
			// skip
			return nil
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		// dereference, and merge theirs elements
		dstE := dst.Elem()
		srcE := src.Elem()
		if dstE.Type() != srcE.Type() {
			return m.convert(dstE, srcE, o)
		}
		return m.deepMerge(dstE, srcE, o)
	case reflect.Interface:
		if src.IsNil() {
			// skip
			return nil
		}
		if dst.IsNil() {
			// the directMerge can not merge interface{}(nil) and interface{}(other)
			if dst.CanSet() {
				dst.Set(src)
			}
			return nil
		}
		// var i interface{} = 1
		// use reflcet.ValueOf(&i).Elem() to get value of i, it is an interface and can
		// be set, the element behind interface is an int, but it can not be set
		// so we must set dst directly
		dstE := derefInterface(dst)
		srcE := derefInterface(src)
		if dstE.Type() != srcE.Type() {
			return m.convert(dst, src, o)
		}

		return directMerge(dst, src, o)
	default:
		return directMerge(dst, src, o)
	}
	return nil
}

// Verifies whether a conversion function has a correct signature.
func verifyCustomMergeFunctionSignature(ft reflect.Type) (convertion bool, err error) {
	if ft.Kind() != reflect.Func {
		err = fmt.Errorf("expected func, got: %v", ft)
		return
	}
	if ft.NumIn() != 3 {
		err = fmt.Errorf("expected three 'in' params, got: %v", ft)
		return
	}
	if ft.NumOut() != 2 {
		err = fmt.Errorf("expected two 'out' param, got: %v", ft)
		return
	}
	if ft.In(0) != ft.In(1) {
		convertion = true
	}
	if ft.In(0) != ft.Out(0) {
		err = fmt.Errorf("expected 'in' param 0 and 'out' param 0 must be the same type, got <%v, %v>", ft.In(0), ft.Out(0))
		return
	}
	opts := &Options{}
	if e, a := reflect.TypeOf(opts), ft.In(2); e != a {
		err = fmt.Errorf("expected '%v' arg for 'in' param 2, got '%v' (%v)", e, a, ft)
		return
	}
	var forErrorType error
	// This convolution is necessary, otherwise TypeOf picks up on the fact
	// that forErrorType is nil.
	errorType := reflect.TypeOf(&forErrorType).Elem()
	if ft.Out(1) != errorType {
		err = fmt.Errorf("expected 'out' param 1 is error, got: %v", ft)
		return
	}
	return
}

// directMerge treats dst and src as single entity and use dst.Set(src)
// to merge them directly
// the dst and src must be the same type
func directMerge(dst, src reflect.Value, o *Options) error {
	// get the element behind interface{}
	dstE := derefInterface(dst)
	srcE := derefInterface(src)

	if !dstE.IsValid() {
		return fmt.Errorf("directMerge: invalid dst")
	}

	if !srcE.IsValid() {
		return nil
	}

	if dstE.Type() != srcE.Type() {
		return fmt.Errorf("directMerge: src %v and dst %v must be of same type", srcE.Type(), dstE.Type())
	}
	if !dst.CanSet() {
		// can not set
		return nil
	}
	if o.Overwrite || isEmptyValue(dstE) {
		// if overwrite or element behind dst is empty (ingore interface{})
		dst.Set(srcE)
	}
	return nil
}

func convertible(dst, src reflect.Type) bool {
	switch src.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch dst.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return true
		case reflect.Float32, reflect.Float64:
			return true
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		switch dst.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return true
		case reflect.Float32, reflect.Float64:
			return true
		}
	case reflect.Float32, reflect.Float64:
		switch dst.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return true
		case reflect.Float32, reflect.Float64:
			return true
		}
	case reflect.Complex64, reflect.Complex128:
		switch dst.Kind() {
		case reflect.Complex64, reflect.Complex128:
			return true
		}
	case reflect.String:
		if dst.Kind() == reflect.String {
			return true
		}
	}
	return false
}

func hashable(in reflect.Type) bool {
	switch in.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Bool:
		return true
	case reflect.String:
		return true
	case reflect.Ptr:
		return true
	case reflect.UnsafePointer:
		return true
	}

	return false
}

func derefInterface(in reflect.Value) reflect.Value {
	for in.Kind() == reflect.Interface {
		in = in.Elem()
	}
	return in
}

func derefPtr(in reflect.Value) reflect.Value {
	// TODO:
	for in.Kind() == reflect.Ptr || in.Kind() == reflect.Interface {
		in = in.Elem()
	}
	return in
}

func hasUnexportedField(tpy reflect.Type) bool {
	for i := 0; i < tpy.NumField(); i++ {
		field := tpy.Field(i)
		// PkgPath  is empty for upper case (exported) field names.
		if len(field.PkgPath) > 0 {
			return true
		}
	}
	return false
}

// copy from encoding/json
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
