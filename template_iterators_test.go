package wfl_test

import (
	"github.com/dgruber/wfl"

	"github.com/dgruber/drmaa2interface"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"
)

var _ = Describe("TemplateIterators", func() {

	Context("NewEnvSequenceIterator", func() {

		jt := drmaa2interface.JobTemplate{}

		It("should set the environment variable if not set", func() {
			iter := wfl.NewEnvSequenceIterator("test", 100, 2)
			newJt := iter(jt)

			Ω(jt.JobEnvironment).Should(BeNil())
			Ω(newJt.JobEnvironment).ShouldNot(BeNil())
			Ω(newJt.JobEnvironment["test"]).Should(Equal("100"))

			tmpl := wfl.NewTemplate(jt).AddIterator("test", iter)
			newJt = tmpl.Next()

			Ω(newJt.JobEnvironment).ShouldNot(BeNil())
			Ω(newJt.JobEnvironment["test"]).Should(Equal("102"))
			Ω(jt.JobEnvironment).Should(BeNil())
		})

		It("should override the environment variable if set", func() {
			iter := wfl.NewEnvSequenceIterator("test", 100, 2)

			jt.JobEnvironment = make(map[string]string, 1)
			jt.JobEnvironment["test"] = "1"

			newJt := iter(jt)

			Ω(jt.JobEnvironment).ShouldNot(BeNil())
			Ω(newJt.JobEnvironment).ShouldNot(BeNil())
			Ω(newJt.JobEnvironment["test"]).Should(Equal("100"))

			tmpl := wfl.NewTemplate(jt).AddIterator("test", iter)
			newJt = tmpl.Next()

			Ω(newJt.JobEnvironment).ShouldNot(BeNil())
			Ω(newJt.JobEnvironment["test"]).Should(Equal("102"))
			Ω(jt.JobEnvironment).ShouldNot(BeNil())
		})

	})

	Context("NewTimeIterator", func() {

		jt := drmaa2interface.JobTemplate{}

		It("should return a new job template after the given duration", func() {
			iter := wfl.NewTimeIterator(time.Millisecond * 30)
			Ω(iter).ShouldNot(BeNil())
			tmpl := wfl.NewTemplate(jt).AddIterator("timer", iter)
			now := time.Now()
			tmpl.Next()
			tmpl.Next()
			Ω(time.Now()).Should(BeTemporally(">=", now.Add(time.Millisecond*50)))
		})
	})

})
