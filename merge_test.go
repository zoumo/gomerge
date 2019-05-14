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
	"reflect"
)

var _ = Describe("options", func() {
	AfterEach(func() {
		opts = newOptions()
	})
	DescribeTable(
		"",
		func(f func(*Options), expect func(*Options) bool) {
			f(opts)
			got := expect(opts)
			Expect(got).To(BeTrue())
		},
		Entry(
			"without overwrite",
			WithoutOverwrite,
			func(o *Options) bool {
				return o.Overwrite == false
			}),
		Entry(
			"without go convertion",
			WithoutGoConvertion,
			func(o *Options) bool {
				return o.GoConvertion == false
			}),
		Entry(
			"with slice mode",
			WithSliceMode(UniteSlice),
			func(o *Options) bool {
				return o.SliceMode == UniteSlice
			}),
		Entry(
			"with converter",
			WithConverters(func(int, int, *Options) (int, error) { return 0, nil }),
			func(o *Options) bool {
				return len(o.delegate.mergeFuncs) == 1
			},
		),
	)
})

var _ = Describe("resolve values", func() {
	v1 := 1
	v2 := "2"
	DescribeTable(
		"",
		func(dst, src interface{}, dstKind, srcKind reflect.Kind, wantErr bool) {
			dstV, srcV, err := resolveValues(dst, src)
			if wantErr {
				Expect(err).NotTo(BeNil())
			} else {
				Expect(dstV.Kind()).To(Equal(dstKind))
				Expect(srcV.Kind()).To(Equal(srcKind))
			}
		},
		Entry("dst is nil", nil, nil, reflect.Invalid, reflect.Invalid, true),
		Entry("dst is not ptr", 1, nil, reflect.Invalid, reflect.Invalid, true),
		Entry("dst is invalid", (*int)(nil), nil, reflect.Int, reflect.Invalid, true),
		Entry("src is nil", &v1, nil, reflect.Int, reflect.Invalid, true),
		Entry("simple", &v1, &v2, reflect.Int, reflect.String, false),
	)
})

var _ = Describe("Merge", func() {
	type test struct {
		Int    int
		String string
		Bool   bool
		Ptr    *int
		Slice  []string
		Map    map[string]string
	}

	var dst, src test
	var vi int

	BeforeEach(func() {
		vi = 2
		dst = test{
			Ptr:   &vi,
			Slice: []string{"1"},
			Map:   map[string]string{"1": "1"},
		}
		src = test{
			Int:    2,
			String: "2",
			Bool:   true,
			Ptr:    nil,
			Slice:  []string{"2"},
		}
	})

	It("with overwrite", func() {
		Merge(&dst, src)
		Expect(dst).To(Equal(test{
			Int:    2,
			String: "2",
			Bool:   true,
			Ptr:    &vi,
			Slice:  []string{"2"},
			Map:    map[string]string{"1": "1"},
		}))
	})
})
