package matrix_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl/pkg/matrix"
)

var _ = Describe("Helper", func() {

	Context("GetNextValue", func() {

		It("should increment the current number", func() {

			n, err := matrix.GetNextValue([]int{1, 1, 2}, []int{0, 0, 0})
			Expect(err).To(BeNil())
			Expect(len(n)).To(BeNumerically("==", 3))
			Expect(n).To(Equal([]int{0, 0, 1}))

			n, err = matrix.GetNextValue([]int{1, 1, 2}, n)
			Expect(err).To(BeNil())
			Expect(n).To(Equal([]int{0, 0, 2}))

			n, err = matrix.GetNextValue([]int{1, 1, 2}, n)
			Expect(err).To(BeNil())
			Expect(n).To(Equal([]int{0, 1, 0}))

			n, err = matrix.GetNextValue([]int{1, 1, 2}, n)
			Expect(err).To(BeNil())
			Expect(n).To(Equal([]int{0, 1, 1}))

			n, err = matrix.GetNextValue([]int{1, 1, 2}, n)
			Expect(err).To(BeNil())
			Expect(n).To(Equal([]int{0, 1, 2}))

			n, err = matrix.GetNextValue([]int{1, 1, 2}, n)
			Expect(err).To(BeNil())
			Expect(n).To(Equal([]int{1, 0, 0}))

			n, err = matrix.GetNextValue([]int{1, 1, 2}, n)
			Expect(err).To(BeNil())
			Expect(n).To(Equal([]int{1, 0, 1}))

			n, err = matrix.GetNextValue([]int{1, 1, 2}, n)
			Expect(err).To(BeNil())
			Expect(n).To(Equal([]int{1, 0, 2}))

			n, err = matrix.GetNextValue([]int{1, 1, 2}, n)
			Expect(err).To(BeNil())
			Expect(n).To(Equal([]int{1, 1, 0}))

			n, err = matrix.GetNextValue([]int{1, 1, 2}, n)
			Expect(err).To(BeNil())
			Expect(n).To(Equal([]int{1, 1, 1}))

			n, err = matrix.GetNextValue([]int{1, 1, 2}, n)
			Expect(err).To(BeNil())
			Expect(n).To(Equal([]int{1, 1, 2}))

			n, err = matrix.GetNextValue([]int{1, 1, 2}, n)
			Expect(err).NotTo(BeNil())
			Expect(n).To(BeNil())

			// most simple test
			one, err := matrix.GetNextValue([]int{1}, []int{0})
			Expect(err).To(BeNil())
			Expect(one).To(Equal([]int{1}))

			one, err = matrix.GetNextValue([]int{1}, one)
			Expect(err).NotTo(BeNil())
			Expect(one).To(BeNil())

		})

	})

	Context("CopyTemplate", func() {

		It("should copy the template", func() {

			t := drmaa2interface.JobTemplate{
				RemoteCommand: "test",
				Args:          []string{"arg1", "arg2"},
			}
			t.ExtensionList = map[string]string{
				"key1": "value1",
				"key2": "value2",
			}

			jt, err := matrix.CopyJobTemplate(t)
			Expect(err).To(BeNil())
			Expect(jt.RemoteCommand).To(Equal("test"))
			Expect(len(jt.Args)).To(BeNumerically("==", 2))
			Expect(jt.Args[0]).To(Equal("arg1"))
			Expect(jt.Args[1]).To(Equal("arg2"))
			Expect(len(jt.ExtensionList)).To(BeNumerically("==", 2))
			Expect(jt.ExtensionList["key1"]).To(Equal("value1"))
			Expect(jt.ExtensionList["key2"]).To(Equal("value2"))

		})

	})

	Context("Replacements", func() {

		It("should replace all values in the JobTemplate", func() {

			t := drmaa2interface.JobTemplate{
				RemoteCommand: "test",
				Args:          []string{"arg1", "arg2"},
				JobEnvironment: map[string]string{
					"key1": "value1",
					"key2": "key",
				},
			}
			t.ExtensionList = map[string]string{
				"key1": "value1",
				"key2": "value2",
			}

			jt, err := matrix.ReplaceInField(t, "RemoteCommand", "test", "test2")
			Expect(err).To(BeNil())
			Expect(jt.RemoteCommand).To(Equal("test2"))
			Expect(len(jt.Args)).To(BeNumerically("==", 2))
			Expect(jt.Args[0]).To(Equal("arg1"))
			Expect(jt.Args[1]).To(Equal("arg2"))
			Expect(len(jt.ExtensionList)).To(BeNumerically("==", 2))
			Expect(jt.ExtensionList["key1"]).To(Equal("value1"))
			Expect(jt.ExtensionList["key2"]).To(Equal("value2"))

			jt, err = matrix.ReplaceInField(t, "Args", "arg", "X")
			Expect(err).To(BeNil())
			Expect(jt.RemoteCommand).To(Equal("test"))
			Expect(len(jt.Args)).To(BeNumerically("==", 2))
			Expect(jt.Args[0]).To(Equal("X1"))
			Expect(jt.Args[1]).To(Equal("X2"))

			jt, err = matrix.ReplaceInField(t, "SubmitAsHold", "", "true")
			Expect(err).To(BeNil())
			Expect(jt.SubmitAsHold).To(BeTrue())

			jt, err = matrix.ReplaceInField(t, "JobEnvironment", "key", "new")
			Expect(err).To(BeNil())
			Expect(len(jt.JobEnvironment)).To(BeNumerically("==", 2))
			Expect(jt.JobEnvironment["new1"]).To(Equal("value1"))
			Expect(jt.JobEnvironment["new2"]).To(Equal("new"))
		})

	})

})
