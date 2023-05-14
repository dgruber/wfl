package wfl_test

import (
	"fmt"
	"os"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Llm", func() {

	Context("OpenAI provider related tests", func() {

		It("should return an error when no OpenAI token is provided", func() {
			flow := wfl.NewWorkflow(wfl.NewProcessContext()).WithLLMOpenAI(
				wfl.OpenAIConfig{
					Token: "",
				})
			Expect(flow.HasError()).To(BeTrue())
		})

		It("should return an error when wrong OpenAI token is provided", func() {
			flow := wfl.NewWorkflow(wfl.NewProcessContext()).WithLLMOpenAI(
				wfl.OpenAIConfig{
					Token: "wrong",
				})
			Expect(flow.HasError()).To(BeTrue())
		})

		It("should not return an error when correct OpenAI token is provided", func() {
			if os.Getenv("OPENAI_KEY") == "" {
				Skip("OPENAI_KEY not set")
			}
			flow := wfl.NewWorkflow(wfl.NewProcessContext()).WithLLMOpenAI(
				wfl.OpenAIConfig{
					Token: os.Getenv("OPENAI_KEY"),
				})
			Expect(flow.HasError()).To(BeFalse())
		})

	})

	Context("ErrorP related tests", func() {

		It("should convert an error message", func() {
			if os.Getenv("OPENAI_KEY") == "" {
				Skip("OPENAI_KEY not set")
			}

			flow := wfl.NewWorkflow(wfl.NewProcessContext()).WithLLMOpenAI(
				wfl.OpenAIConfig{
					Token: os.Getenv("OPENAI_KEY"),
				})
			job := flow.RunT(drmaa2interface.JobTemplate{
				RemoteCommand: "/bin/bsh", // NOTE the job error here!
			})
			Expect(job.Errored()).To(BeTrue())
			Expect(job.ErrorP("Explain the error and provide a solution")).NotTo(BeEmpty())
		})

	})

	Context("OutputP related tests", func() {

		It("should convert a job output", func() {

			if os.Getenv("OPENAI_KEY") == "" {
				Skip("OPENAI_KEY not set")
			}

			flow := wfl.NewWorkflow(wfl.NewProcessContext()).WithLLMOpenAI(
				wfl.OpenAIConfig{
					Token: os.Getenv("OPENAI_KEY"),
				})
			job := flow.RunT(drmaa2interface.JobTemplate{
				RemoteCommand: "/bin/bash",
				Args: []string{
					"-c",
					`uname -a`,
				},
				OutputPath: "/tmp/wfl-test-output.txt",
			}).Wait()
			Expect(job.Errored()).To(BeFalse())
			Expect(job.Success()).To(BeTrue())
			Expect(job.OutputP("Describe the system")).NotTo(BeEmpty())

		})

	})

	Context("TemplateP related tests", func() {

		It("should create a job template", func() {

			if os.Getenv("OPENAI_KEY") == "" {
				Skip("OPENAI_KEY not set")
			}

			flow := wfl.NewWorkflow(wfl.NewProcessContext()).WithLLMOpenAI(
				wfl.OpenAIConfig{
					Token: os.Getenv("OPENAI_KEY"),
				})

			// generates a bash job template for Linux
			template, err := flow.TemplateP("How much memory is available?")
			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())
			Expect(template.RemoteCommand).To(Equal("/bin/bash"))
			Expect(template.Args[1]).NotTo(Equal(""))

			fmt.Printf("Script: %s\n", template.Args[1])

		})

	})

})
