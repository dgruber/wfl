package wfl_test

import (
	"github.com/dgruber/wfl"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dgruber/drmaa2interface"
	"strconv"
)

var _ = Describe("Template", func() {

	var jt drmaa2interface.JobTemplate

	// itr increments the number in the first argument for the job by one
	itr := func(t drmaa2interface.JobTemplate) drmaa2interface.JobTemplate {
		if len(t.Args) == 0 {
			t.Args = []string{"0"}
		} else {
			i, err := strconv.Atoi(t.Args[0])
			if err != nil {
				t.Args = []string{"0"}
			} else {
				i++
				t.Args[0] = strconv.Itoa(i)
			}
		}
		return t
	}

	rmt := func(t drmaa2interface.JobTemplate) drmaa2interface.JobTemplate {
		t.RemoteCommand = "rmt"
		return t
	}

	Context("happy path", func() {

		BeforeEach(func() {
			jt = drmaa2interface.JobTemplate{
				RemoteCommand: "test",
				Args:          []string{"0"},
			}
		})

		It("should create JobTemplates when no Iterator or Mapping function is registered", func() {
			template := wfl.NewTemplate(jt)
			Ω(template).ShouldNot(BeNil())

			njt := template.Next()
			Ω(njt.RemoteCommand).Should(Equal("test"))

			Ω(template.Next().RemoteCommand).Should(Equal("test"))
			Ω(template.Next().RemoteCommand).Should(Equal("test"))
		})

		It("should apply a registered Iterator when Next is called", func() {
			template := wfl.NewTemplate(jt)
			Ω(template).ShouldNot(BeNil())

			Ω(template.Next().Args[0]).Should(Equal("0"))

			template.AddIterator("test", itr)

			Ω(template.Next().Args[0]).Should(Equal("1"))
			Ω(template.Next().Args[0]).Should(Equal("2"))
			Ω(template.Next().Args[0]).Should(Equal("3"))

			template.AddIterator("test3", rmt)
			Ω(template.Next().RemoteCommand).Should(Equal("rmt"))

		})

		It("should apply multiple Iterators when Next is called", func() {
			template := wfl.NewTemplate(jt)
			Ω(template).ShouldNot(BeNil())

			Ω(template.Next().Args[0]).Should(Equal("0"))

			template.AddIterator("test", itr)
			template.AddIterator("test2", itr)

			Ω(template.Next().Args[0]).Should(Equal("2"))
			Ω(template.Next().Args[0]).Should(Equal("4"))
			Ω(template.Next().Args[0]).Should(Equal("6"))
		})

		It("should temporarly convert a JobTemplate with MapTo a given output system", func() {
			template := wfl.NewTemplate(jt)
			Ω(template).ShouldNot(BeNil())

			Ω(template.MapTo("nonExisting").Args[0]).Should(Equal("0"))

			template.AddMap("existing", itr)

			Ω(template.MapTo("nonExisting").Args[0]).Should(Equal("0"))

			Ω(template.MapTo("existing").Args[0]).Should(Equal("1"))
			// non-permanent changes
			Ω(template.MapTo("existing").Args[0]).Should(Equal("1"))

			Ω(template.MapTo("nonExisting").Args[0]).Should(Equal("0"))
			Ω(template.MapTo("nonExisting").Args[0]).Should(Equal("0"))
		})

		It("should apply multiple Iterators and the mapping function when Next is called", func() {
			template := wfl.NewTemplate(jt)
			Ω(template).ShouldNot(BeNil())

			Ω(template.Next().Args[0]).Should(Equal("0"))

			template.AddIterator("test", itr)
			template.AddIterator("test2", itr)

			template.AddMap("existing", itr)

			Ω(template.NextMap("existing").Args[0]).Should(Equal("3"))     // 2 + 1
			Ω(template.NextMap("non-existing").Args[0]).Should(Equal("4")) // 4 + 0
			Ω(template.NextMap("existing").Args[0]).Should(Equal("7"))     // 6 + 1
		})
	})

})
