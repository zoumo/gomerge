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
	"strconv"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/extensions/table"
	"github.com/onsi/gomega"
)

func init() {
	config.DefaultReporterConfig.NoColor = true
}

func TestPorterSuit(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Porter suit")
}

var (
	Describe       = ginkgo.Describe
	DescribeTable  = table.DescribeTable
	Entry          = table.Entry
	Context        = ginkgo.Context
	It             = ginkgo.It
	BeforeEach     = ginkgo.BeforeEach
	AfterEach      = ginkgo.AfterEach
	JustBeforeEach = ginkgo.JustBeforeEach
	JustAfterEach  = ginkgo.JustAfterEach
	BeforeSuite    = ginkgo.BeforeSuite
	AfterSuit      = ginkgo.AfterSuite
	Fail           = ginkgo.Fail
	Skip           = ginkgo.Skip
	Expect         = gomega.Expect
	Equal          = gomega.Equal
	BeNil          = gomega.BeNil
	BeTrue         = gomega.BeTrue
)

var (
	p    *porter
	opts *Options
)

var _ = BeforeEach(func() {
	p = newPorter()
	opts = newOptions()
})

var _ = Describe("add custom funcs", func() {
	AfterEach(func() {
		p = newPorter()
	})

	DescribeTable(
		"",
		func(ft interface{}, mergeLen, convertLen int, wantErr bool) {
			err := p.addCustomFuncs(ft)
			if wantErr {
				Expect(err).NotTo(BeNil())
			} else {
				Expect(err).To(BeNil())
				Expect(p.mergeFuncs).To(gomega.HaveLen(mergeLen))
				Expect(p.convertFuncs).To(gomega.HaveLen(convertLen))
			}
		},
		Entry(
			"add merge func",
			func(d, s float64, o *Options) (float64, error) {
				return s, nil
			},
			1,
			0,
			false,
		),
		Entry(
			"add convert func",
			func(d string, s int, o *Options) (string, error) {
				return strconv.Itoa(s), nil
			},
			0,
			1,
			false,
		),
		Entry(
			"err",
			func(d string, s int, o *Options) (int, error) {
				return 0, nil
			},
			0,
			0,
			true,
		),
	)

})

var _ = Describe("custom merge function", func() {
	BeforeEach(func() {
		p.addCustomFuncs(func(dst, src int, o *Options) (int, error) {
			return dst + src, nil
		})
	})
	AfterEach(func() {
		p = newPorter()
	})
	Context("with overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = true
		})
		It("add int", func() {
			dst := 1
			src := 2
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(3))
		})
		It("tempInt", func() {
			type tempInt int
			dst := tempInt(1)
			src := tempInt(2)
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(tempInt(2)))
		})
	})
	Context("without overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = false
		})
		It("add int", func() {
			dst := 1
			src := 2
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(1))
		})
	})
})

var _ = Describe("convert", func() {
	Context("with go convertion", func() {
		BeforeEach(func() {
			opts.Overwrite = true
			opts.GoConvertion = true
		})
		It("int32 to int", func() {
			dst := 1
			src := int32(2)
			p.convert(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(2))
		})
		It("float64 to int", func() {
			dst := 1
			src := 2.1
			p.convert(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(2))
		})
		It("int to tempInt32", func() {
			type tempInt int32
			dst := tempInt(1)
			src := 2
			p.convert(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(tempInt(2)))
		})
		It("*int32 to *int", func() {
			v1 := 1
			v2 := int32(2)
			dst := &v1
			src := &v2
			p.convert(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(*dst).To(Equal(2))
		})
		It("*int to int", func() {
			v2 := 2
			dst := 1
			src := &v2
			p.convert(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(2))
		})
		It("interface{}", func() {
			var dst, src interface{}
			dst = 1
			src = int32(2)
			p.convert(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(&src).Elem(), opts)
			Expect(dst).To(Equal(2))
		})
	})
	// Context("without overwrite", func() {
	// 	BeforeEach(func() {
	// 		opts.Overwrite = false
	// 	})
	// 	It("int32 to int", func() {
	// 		dst := 1
	// 		src := int32(2)
	// 		p.convert(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
	// 		Expect(dst).To(Equal(1))
	// 	})
	// })
})

var _ = Describe("deep merge basic type, without ptr interface{} map slice struct", func() {
	Context("with overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = true
		})
		It("string", func() {
			dst := "1"
			src := "2"
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("int", func() {
			dst := 1
			src := 2
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("bool", func() {
			dst := true
			src := false
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("float64", func() {
			dst := 1.1
			src := 1.2
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
	})
	Context("without overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = false
		})
		It("string", func() {
			dst := "1"
			src := "2"
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal("1"))
			dst = ""
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("int", func() {
			dst := 1
			src := 2
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(1))
			dst = 0
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("bool", func() {
			dst := true
			src := false
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(BeTrue())
			dst = false
			src = true
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("float64", func() {
			dst := 1.1
			src := 1.2
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(1.1))
			dst = float64(0)
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
	})
})

var _ = Describe("deep merge ptr", func() {
	Context("with overwrite", func() {
		var (
			dst, src *int
			i1, i2   int
		)
		BeforeEach(func() {
			opts.Overwrite = true
			i1, i2 = 0, 1
			dst, src = &i1, &i2
		})

		It("simple", func() {
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).NotTo(BeNil())
			Expect(dst == src).NotTo(BeTrue())
			Expect(*dst).To(Equal(1))
		})
		It("dst is nil", func() {
			dst = nil
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).NotTo(BeNil())
			Expect(dst == src).NotTo(BeTrue())
			Expect(*dst).To(Equal(1))
		})
		It("src is nil", func() {
			src = nil
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).NotTo(BeNil())
			Expect(dst == src).NotTo(BeTrue())
			Expect(*dst).To(Equal(0))
		})
		It("dst and src are nil", func() {
			dst = nil
			src = nil
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(BeNil())
		})
	})

	Context("without overwrite", func() {
		var (
			dst, src *int
			i1, i2   int
		)
		BeforeEach(func() {
			opts.Overwrite = false
			i1, i2 = 0, 1
			dst, src = &i1, &i2
		})

		It("simple", func() {
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).NotTo(BeNil())
			Expect(dst == src).NotTo(BeTrue())
			Expect(*dst).To(Equal(1))

			*src = 2
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(*dst).To(Equal(1), "dst is not empty, src should not overwrite dst")

		})
		It("dst is nil", func() {
			dst = nil
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).NotTo(BeNil())
			Expect(dst == src).NotTo(BeTrue())
			Expect(*dst).To(Equal(1))
		})
		It("src is nil", func() {
			src = nil
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).NotTo(BeNil())
			Expect(dst == src).NotTo(BeTrue())
			Expect(*dst).To(Equal(0))
		})
		It("dst and src are nil", func() {
			dst = nil
			src = nil
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(BeNil())
		})
	})
})

var _ = Describe("deep merge interface{}", func() {
	Context("with overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = true
		})
		DescribeTable(
			"",
			func(dst, src, expect interface{}) {
				p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(&src).Elem(), opts)
				Expect(dst).To(Equal(expect))
			},
			Entry("int", 1, 2, 2),
			Entry("map should be replaced", map[string]string{"1": "1"}, map[string]string{"2": "2"}, map[string]string{"2": "2"}),
			Entry("slice should be replaced", []string{"1"}, []string{"2"}, []string{"2"}),
			Entry("with converter for int", 1, int32(2), 2),
			Entry("dst is nil", nil, 2, 2),
			Entry("src is nil", 1, nil, 1),
		)
	})
})

var _ = Describe("deep merge struct", func() {
	type export struct {
		A string
	}
	type unexport struct {
		A string
		b int
	}
	var (
		exportDst, exportSrc     export
		unexportDst, unexportSrc unexport
	)
	BeforeEach(func() {
		exportDst = export{
			A: "1",
		}
		exportSrc = export{
			A: "2",
		}
		unexportDst = unexport{
			A: "1",
			b: 1,
		}
		unexportSrc = unexport{
			A: "2",
			b: 2,
		}
	})

	Context("with overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = true
		})
		It("export", func() {
			p.deepMerge(reflect.ValueOf(&exportDst).Elem(), reflect.ValueOf(exportSrc), opts)
			Expect(exportDst.A).To(Equal("2"))
		})

		It("unexport struct should be overwritten", func() {
			p.deepMerge(reflect.ValueOf(&unexportDst).Elem(), reflect.ValueOf(unexportSrc), opts)
			Expect(unexportDst.A).To(Equal("2"))
			Expect(unexportDst.b).To(Equal(2))
		})
	})
	Context("without overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = false
		})
		It("export", func() {
			p.deepMerge(reflect.ValueOf(&exportDst).Elem(), reflect.ValueOf(exportSrc), opts)
			Expect(exportDst.A).To(Equal("1"))
		})

		It("unexport struct should be overwritten", func() {
			p.deepMerge(reflect.ValueOf(&unexportDst).Elem(), reflect.ValueOf(unexportSrc), opts)
			Expect(unexportDst.A).To(Equal("1"))
			Expect(unexportDst.b).To(Equal(1))
		})
	})

})

var _ = Describe("deep merge slice", func() {
	Context("with overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = true
		})
		DescribeTable(
			"replace mode",
			func(dst, src, expect []int) {
				opts.SliceMode = ReplaceSlice
				p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst).To(Equal(expect))
			},
			Entry("simple", []int{1, 2}, []int{2, 3}, []int{2, 3}),
			Entry("dst is nil", nil, []int{2, 3}, []int{2, 3}),
			Entry("dst is empty", []int{}, []int{2, 3}, []int{2, 3}),
			Entry("src is nil", []int{1, 2}, nil, []int{1, 2}),
			Entry("src is empty", []int{1, 2}, []int{}, []int{}),
		)
		DescribeTable(
			"append mode",
			func(dst, src, expect []int) {
				opts.SliceMode = AppendSlice
				p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst).To(Equal(expect))
			},
			Entry("simple", []int{1, 2}, []int{2, 3}, []int{1, 2, 2, 3}),
			Entry("dst is nil", nil, []int{2, 3}, []int{2, 3}),
			Entry("dst is empty", []int{}, []int{2, 3}, []int{2, 3}),
			Entry("src is nil ", []int{1, 2}, nil, []int{1, 2}),
			Entry("src is empty ", []int{1, 2}, []int{}, []int{1, 2}),
		)
		DescribeTable(
			"union mode",
			func(dst, src, expect []int) {
				opts.SliceMode = UniteSlice
				p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst).To(Equal(expect))
			},
			Entry("simple", []int{1, 2}, []int{2, 3}, []int{1, 2, 3}),
			Entry("dst is nil", nil, []int{2, 3}, []int{2, 3}),
			Entry("dst is empty", []int{}, []int{2, 3}, []int{2, 3}),
			Entry("src is nil ", []int{1, 2}, nil, []int{1, 2}),
			Entry("src is empty ", []int{1, 2}, []int{}, []int{1, 2}),
		)
		DescribeTable(
			"union mode for bool slice",
			func(dst, src, expect []bool) {
				opts.SliceMode = UniteSlice
				p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst).To(Equal(expect))
			},
			Entry("simple", []bool{true, false}, []bool{true, false}, []bool{true, false, true, false}),
		)
	})
	Context("without overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = false
		})
		DescribeTable(
			"replace mode",
			func(dst, src, expect []int) {
				p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst).To(Equal(expect))
			},
			Entry("simple", []int{1, 2}, []int{2, 3}, []int{1, 2}),
			Entry("dst is nil", nil, []int{2, 3}, []int{2, 3}),
			Entry("dst is empty", []int{}, []int{2, 3}, []int{2, 3}),
			Entry("src is nil", []int{1, 2}, nil, []int{1, 2}),
			Entry("src is empty", []int{1, 2}, []int{}, []int{1, 2}),
		)
		DescribeTable(
			"append mode",
			func(dst, src, expect []int) {
				opts.SliceMode = AppendSlice
				p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst).To(Equal(expect))
			},
			Entry("simple", []int{1, 2}, []int{2, 3}, []int{1, 2}),
			Entry("dst is nil", nil, []int{2, 3}, []int{2, 3}),
			Entry("dst is empty", []int{}, []int{2, 3}, []int{2, 3}),
			Entry("src is nil ", []int{1, 2}, nil, []int{1, 2}),
			Entry("src is empty ", []int{1, 2}, []int{}, []int{1, 2}),
		)
		DescribeTable(
			"union mode",
			func(dst, src, expect []int) {
				opts.SliceMode = UniteSlice
				p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst).To(Equal(expect))
			},
			Entry("simple", []int{1, 2}, []int{2, 3}, []int{1, 2}),
			Entry("dst is nil", nil, []int{2, 3}, []int{2, 3}),
			Entry("dst is empty", []int{}, []int{2, 3}, []int{2, 3}),
			Entry("src is nil ", []int{1, 2}, nil, []int{1, 2}),
			Entry("src is empty ", []int{1, 2}, []int{}, []int{1, 2}),
		)
	})
})

var _ = Describe("deep merge map[string]string", func() {
	var (
		mDst, mSrc map[string]string
		empty      map[string]string
	)
	mDst = map[string]string{
		"1": "1",
		"2": "2",
	}
	mSrc = map[string]string{
		"1": "2",
		"3": "3",
	}
	empty = map[string]string{}

	Context("with overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = true
		})
		DescribeTable(
			"",
			func(dst, src, expect map[string]string) {
				dstCopy := dst
				p.deepMerge(reflect.ValueOf(&dstCopy).Elem(), reflect.ValueOf(src), opts)
				Expect(dstCopy).To(Equal(expect))
			},
			Entry("simple", mDst, mSrc, map[string]string{"1": "2", "2": "2", "3": "3"}),
			Entry("dst is nil", nil, mSrc, mSrc),
			Entry("dst is empty", empty, mSrc, mSrc),
			Entry("src is nil", mDst, nil, mDst),
			Entry("src is empty", mDst, empty, mDst),
		)
	})
})

var _ = Describe("deep merge map[string]interface{}", func() {

	Context("with overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = true
		})

		It("", func() {
			s1 := "1"
			s2 := "2"
			ptr1 := &s1
			ptr2 := &s2
			dst := map[string]interface{}{
				"str": "",
				"int": 1,
				"map": map[string]interface{}{
					"a": "",
					"b": "b",
				},
				"strSlice": []string{
					"1", "2",
				},
				"slice": []interface{}{
					"1", 1,
				},
				"ptr": ptr1,
			}
			src := map[string]interface{}{
				"str": "2",
				"int": 2,
				"map": map[string]interface{}{
					"a": "2",
					"b": "2",
					"c": "2",
				},
				"strSlice": []string{
					"2", "3",
				},
				"slice": []interface{}{
					"2", 2,
				},
				"ptr": ptr2,
			}
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
	})

	Context("without overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = false
		})

		It("", func() {
			s1 := "1"
			s2 := "2"
			ptr1 := &s1
			ptr2 := &s2
			dst := map[string]interface{}{
				"str": "",
				"int": 1,
				"map": map[string]interface{}{
					"a": "",
					"b": "b",
				},
				"strSlice": []string{
					"1", "2",
				},
				"slice": []interface{}{
					"1", 1,
				},
				"ptr": ptr1,
			}
			src := map[string]interface{}{
				"str": "2",
				"int": 2,
				"map": map[string]interface{}{
					"a": "2",
					"b": "2",
					"c": "2",
				},
				"strSlice": []string{
					"2", "3",
				},
				"slice": []interface{}{
					"2", 2,
				},
				"ptr": ptr2,
			}
			p.deepMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(map[string]interface{}{
				"str": "2",
				"int": 1,
				"map": map[string]interface{}{
					"a": "2",
					"b": "b",
					"c": "2",
				},
				"strSlice": []string{
					"1", "2",
				},
				"slice": []interface{}{
					"1", 1,
				},
				"ptr": ptr1,
			}))
		})
	})

})

var _ = Describe("direct merge", func() {
	Context("with overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = true
		})
		It("string", func() {
			dst := "1"
			src := "2"
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("int", func() {
			dst := 1
			src := 2
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("bool", func() {
			dst := true
			src := false
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("float64", func() {
			dst := 1.1
			src := 1.2
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})

		Context("int ptr", func() {
			var (
				dst, src *int
				i1, i2   int
			)
			BeforeEach(func() {
				i1, i2 = 0, 1
				dst, src = &i1, &i2
			})
			It("simple", func() {
				directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst == src).To(BeTrue())
				Expect(*dst).To(Equal(i2))
			})
			It("dst is nil", func() {
				dst = nil
				directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst).NotTo(BeNil())
				Expect(dst == src).To(BeTrue())
				Expect(*dst).To(Equal(i2))
			})
			It("src is nil, dst should be nil", func() {
				src = nil
				directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst).To(BeNil())
			})
		})
		DescribeTable(
			"interface{}",
			func(dst, src, expect interface{}, wantErr bool) {
				dstCopy := dst
				err := directMerge(reflect.ValueOf(&dstCopy).Elem(), reflect.ValueOf(src), opts)
				if wantErr {
					Expect(err).NotTo(BeNil())
				} else {
					Expect(dstCopy).To(Equal(expect))
					Expect(err).To(BeNil())
				}
			},
			Entry("int", 1, 2, 2, false),
			Entry("int is 0", 0, 2, 2, false),
			Entry("dst is nil", nil, 2, nil, true),
			Entry("src is nil", 1, nil, 1, false),
			Entry("string and int", "", 2, "", true),
		)
	})
	Context("without overwrite", func() {
		BeforeEach(func() {
			opts.Overwrite = false
		})
		It("string", func() {
			dst := "1"
			src := "2"
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal("1"))
			dst = ""
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("int", func() {
			dst := 1
			src := 2
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(1))
			dst = 0
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("bool", func() {
			dst := true
			src := false
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(BeTrue())
			dst = false
			src = true
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		It("float64", func() {
			dst := 1.1
			src := 1.2
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(1.1))
			dst = float64(0)
			directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
			Expect(dst).To(Equal(src))
		})
		Context("int ptr", func() {
			var (
				dst, src *int
				i1, i2   int
			)
			BeforeEach(func() {
				i1, i2 = 0, 1
				dst, src = &i1, &i2
			})
			It("simple", func() {
				directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst == src).NotTo(BeTrue())
				Expect(*dst).To(Equal(0))
			})
			It("dst is nil", func() {
				dst = nil
				directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst).NotTo(BeNil())
				Expect(dst == src).To(BeTrue())
				Expect(dst).To(Equal(src))
			})
			It("src is nil, dst should not be nil", func() {
				src = nil
				directMerge(reflect.ValueOf(&dst).Elem(), reflect.ValueOf(src), opts)
				Expect(dst).NotTo(BeNil())
			})
		})
		DescribeTable(
			"interface{}",
			func(dst, src, expect interface{}, wantErr bool) {
				dstCopy := dst
				err := directMerge(reflect.ValueOf(&dstCopy).Elem(), reflect.ValueOf(src), opts)
				if wantErr {
					Expect(err).NotTo(BeNil())
				} else {
					Expect(dstCopy).To(Equal(expect))
					Expect(err).To(BeNil())
				}
			},
			Entry("int", 1, 2, 1, false),
			Entry("int is 0", 0, 2, 2, false),
			Entry("dst is nil", nil, 2, nil, true),
			Entry("src is nil", 1, nil, 1, false),
			Entry("string and int", "", 2, "", true),
		)
	})
})

var _ = Describe("derefPtr", func() {
	s := ""

	DescribeTable(
		"",
		func(in interface{}, wantKind reflect.Kind) {
			got := derefPtr(reflect.ValueOf(in))
			Expect(got.Kind()).To(Equal(wantKind))
		},
		Entry("int", 1, reflect.Int),
		Entry("string", "1", reflect.String),
		Entry("nil", nil, reflect.Invalid),
		Entry("string ptr", &s, reflect.String),
	)
})

var _ = Describe("verifyCustomMergeFunctionSignature", func() {
	DescribeTable(
		"",
		func(in interface{}, wantConvertion, wantErr bool) {
			gotConvertion, err := verifyCustomMergeFunctionSignature(reflect.TypeOf(in))
			Expect(gotConvertion).To(Equal(wantConvertion))
			if wantErr {
				Expect(err).NotTo(gomega.BeNil())
			} else {
				Expect(err).To(gomega.BeNil())
			}
		},
		Entry(
			"merge int",
			func(d, s int, o *Options) (int, error) {
				return 0, nil
			},
			false,
			false,
		),
		Entry(
			"int to string",
			func(d string, s int, o *Options) (string, error) {
				return "", nil
			},
			true,
			false,
		),
		Entry(
			"error",
			func(d string, s int, o *Options) (int, error) {
				return 0, nil
			},
			true,
			true,
		),
	)
})

var _ = Describe("hasUnexportedField", func() {
	type export struct {
		A string
	}
	type unexport struct {
		A string
		b int
	}
	DescribeTable(
		"",
		func(in interface{}, want bool) {
			got := hasUnexportedField(reflect.TypeOf(in))
			Expect(got).To(Equal(want))
		},
		Entry("export", export{}, false),
		Entry("unexport", unexport{}, true),
	)
})

var _ = Describe("convertible", func() {
	type temp string
	type tempInt int

	DescribeTable(
		"",
		func(src, dst interface{}, want bool) {
			got := convertible(reflect.TypeOf(dst), reflect.TypeOf(src))
			Expect(got).To(Equal(want))
		},
		Entry("int to int32", 1, int32(1), true),
		Entry("int32 to int", int32(1), 1, true),
		Entry("int to uint", 1, uint(1), true),
		Entry("uint to int", uint(1), 1, true),
		Entry("int to float64", 1, 1.2, true),
		Entry("float64 to int", 1.2, 1, true),
		Entry("string to temp(string)", "str", temp(""), true),
		Entry("temp(string) to string", temp(""), "str", true),
		Entry("int to tempInt(int)", 1, tempInt(1), true),
		Entry("tempInt(int) to int", tempInt(1), 1, true),
		Entry("int to string", 1, "", false),
		Entry("string to int", "", 1, false),
	)
})
