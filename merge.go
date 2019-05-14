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
	"errors"
	"reflect"
)

// Options ..
type Options struct {
	Overwrite      bool
	GoConvertion   bool
	SliceMode      SliceMergeMode
	AppendSlice    bool
	IntersectSlice bool
	delegate       *porter
}

// SliceMergeMode specify which merge strategy will be applied
// when merging slice
type SliceMergeMode string

const (
	// ReplaceSlice replaces tow slices
	ReplaceSlice SliceMergeMode = "Replace"
	// UniteSlice unite tow slices if the element in slice is hashable.
	// If the kind of element is not hashable of bool, it will fallthrough to use ApplenSlice
	// see hashable() to find out what kind is hashable now
	//
	// why we don't unite tow bool slices? Imagine that
	// if we want to unite []bool{true, false, true} and []bool{true}
	// the resule will be []bool{true, false, true}, it makes no sense.
	UniteSlice SliceMergeMode = "Unite"
	// AppendSlice appends all elements of source slice to the target
	AppendSlice SliceMergeMode = "Append"
)

func newOptions() *Options {
	return &Options{
		Overwrite:    true,
		GoConvertion: true,
		SliceMode:    ReplaceSlice,
		delegate:     newPorter(),
	}
}

// WithoutOverwrite ...
func WithoutOverwrite(o *Options) {
	o.Overwrite = false
}

// WithoutGoConvertion disables the golang defaultMerge rules
func WithoutGoConvertion(o *Options) {
	o.GoConvertion = false
}

// WithSliceMode changes slice merge mode
func WithSliceMode(mode SliceMergeMode) func(*Options) {
	return func(o *Options) {
		o.SliceMode = mode
	}
}

// WithConverters add custom convert funcs, func sign is like:
//
// func(dst string, src int, o *Options) (string, error) {}
//
// It follows the rules:
// - the function must have three params
// - the function must have tow return values
// - the first param and first return must be the same type
// - the third param must be *Option
// - the last return must be error
// - the first and second params should be different type
func WithConverters(fns ...interface{}) func(*Options) {
	return WithMergeFuncs(fns...)
}

// WithMergeFuncs add custom merge funcs, func sigh is like
//
// func(dst, src int, o *Options) (int, error) {}
//
// It follows the rules:
// - the function must have three params
// - the function must have tow return values
// - the first, second param and first return must be the same type
// - the third param must be *Option
// - the last return must be error
func WithMergeFuncs(fns ...interface{}) func(*Options) {
	return func(o *Options) {
		err := o.delegate.addCustomFuncs(fns...)
		if err != nil {
			panic(err)
		}
	}
}

// Merge the given source onto the given target following the options given. The target value
// must be a pointer. Merge will accept any two entities, even if their types are diffrent
// as long as there is convert function (see WithConverters).
func Merge(dst, src interface{}, opts ...func(*Options)) error {
	o := newOptions()

	for _, f := range opts {
		f(o)
	}
	return merge(dst, src, o)
}

func merge(dst, src interface{}, o *Options) error {
	if src == nil {
		return nil
	}

	vDst, vSrc, err := resolveValues(dst, src)
	if err != nil {
		return err
	}

	// make a copy to let dst stay in tact when an error occurs
	vDstCopy := vDst
	err = o.delegate.defaultMerge(vDstCopy, vSrc, o)
	if err != nil {
		return err
	}
	vDst.Set(vDstCopy)

	return nil
}

func resolveValues(dst, src interface{}) (vDst, vSrc reflect.Value, err error) {
	if dst == nil {
		err = errors.New("the target can not be nil")
		return
	}

	vDst = reflect.ValueOf(dst)
	if vDst.Kind() != reflect.Ptr {
		err = errors.New("the target must be a pointer")
		return
	}
	vDst = derefPtr(vDst)
	if !vDst.IsValid() {
		err = errors.New("target can not be zero value")
		return
	}

	// we dereference the src if it is a pointer
	vSrc = derefPtr(reflect.ValueOf(src))
	if !vSrc.IsValid() {
		err = errors.New("source can not be zero value")
		return
	}

	return
}
