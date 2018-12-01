package wfl

import (
	g "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dgruber/drmaa2interface"
)

var _ = g.Describe("JobHelp", func() {

	g.Context("JobTemplate merge tests", func() {

		category := "shub://GodloveD/lolcow"
		stdin := "/dev/stdin"
		stdout := "/dev/stdout"
		stderr := "/dev/stderr"

		g.It("should return an unset job template when both are unset", func() {
			var req drmaa2interface.JobTemplate
			var def drmaa2interface.JobTemplate
			jt := mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.RemoteCommand).Should(Equal(""))
			Ω(jt.Args).Should(BeNil())
			Ω(jt.AccountingID).Should(Equal(""))
			Ω(jt.SubmitAsHold).Should(BeFalse())
		})

		g.It("should override JobCategory settings from the default template", func() {
			var req drmaa2interface.JobTemplate
			var def drmaa2interface.JobTemplate

			def.JobCategory = category

			jt := mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.JobCategory).Should(Equal(category))
		})

		g.It("should override input/output/error path settings from the default template", func() {
			var req drmaa2interface.JobTemplate
			var def drmaa2interface.JobTemplate

			def.InputPath = stdin
			def.OutputPath = stdout
			def.ErrorPath = stderr

			jt := mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.InputPath).Should(Equal(stdin))
			Ω(jt.OutputPath).Should(Equal(stdout))
			Ω(jt.ErrorPath).Should(Equal(stderr))
		})

		g.It("should merge environment settings", func() {
			var req drmaa2interface.JobTemplate
			var def drmaa2interface.JobTemplate

			def.JobEnvironment = map[string]string{
				"ENV": "CONTENT",
			}

			jt := mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.JobEnvironment).ShouldNot(BeNil())
			Ω(jt.JobEnvironment["ENV"]).Should(Equal("CONTENT"))

			req.JobEnvironment = map[string]string{
				"ENV": "ORIGINAL",
			}

			jt = mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.JobEnvironment).ShouldNot(BeNil())
			Ω(jt.JobEnvironment["ENV"]).Should(Equal("ORIGINAL"))

			def.JobEnvironment["ENV2"] = "CONTENT2"

			jt = mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.JobEnvironment).ShouldNot(BeNil())
			Ω(jt.JobEnvironment["ENV"]).Should(Equal("ORIGINAL"))
			Ω(jt.JobEnvironment["ENV2"]).Should(Equal("CONTENT2"))

		})

		g.It("should merge stage-in files", func() {
			var req drmaa2interface.JobTemplate
			var def drmaa2interface.JobTemplate

			jt := mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.StageInFiles).Should(BeNil())

			def.StageInFiles = map[string]string{
				"/dir": "/containerDir",
			}

			jt = mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.StageInFiles).ShouldNot(BeNil())
			Ω(jt.StageInFiles["/dir"]).Should(Equal("/containerDir"))

			req.JobEnvironment = map[string]string{
				"/dir": "/containerDir",
			}

			jt = mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.StageInFiles).ShouldNot(BeNil())
			Ω(jt.StageInFiles["/dir"]).Should(Equal("/containerDir"))

			def.StageInFiles["/dir2"] = "/containerDir2"

			jt = mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.StageInFiles).ShouldNot(BeNil())
			Ω(jt.StageInFiles["/dir"]).Should(Equal("/containerDir"))
			Ω(jt.StageInFiles["/dir2"]).Should(Equal("/containerDir2"))
		})

		g.It("should set the extensions if specified in default template", func() {
			var req drmaa2interface.JobTemplate
			var def drmaa2interface.JobTemplate

			def.ExtensionList = map[string]string{
				"the Dø": "a mess like this",
			}
			jt := mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.ExtensionList).ShouldNot(BeNil())
			Ω(jt.ExtensionList["the Dø"]).Should(Equal("a mess like this"))

			req.ExtensionList = map[string]string{
				"Mees Dierdorp": "Wild Windows",
			}

			jt = mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.ExtensionList).ShouldNot(BeNil())
			Ω(jt.ExtensionList["Mees Dierdorp"]).Should(Equal("Wild Windows"))

			_, exists := jt.ExtensionList["the Dø"]
			Ω(exists).Should(BeFalse())

		})

		g.It("should set the candidate machines if specified in default template", func() {
			var req drmaa2interface.JobTemplate
			var def drmaa2interface.JobTemplate

			def.CandidateMachines = []string{"the Dø", "a mess like this"}
			jt := mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.CandidateMachines).ShouldNot(BeNil())
			Ω(jt.CandidateMachines).Should(ContainElement("a mess like this"))

			req.CandidateMachines = []string{"Shadows - Edwin Oosterwald Dub"}

			jt = mergeJobTemplateWithDefaultTemplate(req, def)

			Ω(jt.CandidateMachines).ShouldNot(BeNil())
			Ω(jt.CandidateMachines).Should(ContainElement("Shadows - Edwin Oosterwald Dub"))

			// candidate machines should be completely overriden
			Ω(jt.CandidateMachines).ShouldNot(ContainElement("the Dø"))
		})
	})

})
