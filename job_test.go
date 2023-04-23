package wfl_test

import (
	"sync"

	"github.com/dgruber/wfl"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgruber/drmaa2interface"
)

var _ = Describe("Job", func() {

	makeWfl := func() *wfl.Workflow {
		os.Remove("tmp.db")
		ctx := wfl.NewProcessContext()
		err := ctx.Error()
		Ω(err).Should(BeNil())
		wf := wfl.NewWorkflow(ctx)
		Ω(wf.HasError()).Should(BeFalse())
		return wf
	}

	jobT := drmaa2interface.JobTemplate{RemoteCommand: "sleep", Args: []string{"0"}}

	Context("Simple job operations", func() {

		var (
			wf *wfl.Workflow
		)

		BeforeEach(func() {
			wf = makeWfl()
		})

		It("should be possible to create a first job", func() {
			job := wf.Run("sleep", "0")
			Ω(job).ShouldNot(BeNil())

			job = job.Run("sleep", "0")
			Ω(job).ShouldNot(BeNil())
		})

		It("should be possible to tag a job", func() {
			job := wf.Run("sleep", "0").TagWith("Tag")
			Ω(job).ShouldNot(BeNil())
			Ω(job.Tag()).Should(Equal("Tag"))
			job.TagWith("Tag2")
			Ω(job.Tag()).Should(Equal("Tag2"))
		})

		It("should be possible to run a job every d times", func() {
			forTheNext50ms := time.Now().Add(time.Millisecond * 50)
			err := wfl.NewJob(wf).RunEvery(time.Millisecond*10, forTheNext50ms, "sleep", "0")
			Ω(err).Should(BeNil())
		})

		It("should show the correct exit status", func() {
			job := wf.Run("sleep", "0")
			Ω(job).ShouldNot(BeNil())

			job = job.Run("sleep", "0")
			Ω(job).ShouldNot(BeNil())

			exitStatus := job.ExitStatus()
			Ω(exitStatus).Should(BeNumerically("==", 0))

			job = job.Run("date", "invalidFormat")
			Ω(job).ShouldNot(BeNil())

			exitStatus = job.ExitStatus()
			Ω(exitStatus).Should(BeNumerically("==", 1))
		})

		It("should be possible to suspend and resume a job", func() {
			job := wf.Run("sleep", "1")
			Ω(job).ShouldNot(BeNil())
			job.Suspend()
			Ω(job.State().String()).Should(Equal(drmaa2interface.Suspended.String()))
			job.Resume()
			Ω(job.State().String()).Should(Equal(drmaa2interface.Running.String()))
			job.Wait()
			Ω(job.Success()).Should(BeTrue())
		})

		It("should be possible to kill a job", func() {
			job := wf.Run("/bin/sh", "-c", `for i in {1..10000}; do echo $i && sleep 0.1; done`).Kill()
			Ω(job).ShouldNot(BeNil())
			Ω(job.State().String()).Should(Equal(drmaa2interface.Failed.String()))
		})

		It("should retry a failed job", func() {
			job := wf.Run("sleep", "0")
			Ω(job).ShouldNot(BeNil())

			job = job.Run("sleep", "0").Retry(5)
			Ω(job).ShouldNot(BeNil())
			Ω(job.Success()).Should(BeTrue())

			job = job.Run("date", "invalidFormat").Retry(5)
			Ω(job).ShouldNot(BeNil())
			Ω(job.Success()).Should(BeFalse())
		})

		It("should synchronize", func() {
			start := time.Now()
			job := wf.Run("sleep", "1").Run("sleep", "0").Run("sleep", "0").Synchronize()
			Ω(time.Now()).Should(BeTemporally(">=", start.Add(time.Second)))
			Ω(job).ShouldNot(BeNil())
			Ω(job.Success()).Should(BeTrue())
		})

		It("should report that one job failed", func() {
			start := time.Now()
			job := wf.Run("sleep", "1").Run("sleep", "0").Run("sleep", "0").Synchronize()
			Ω(time.Now()).Should(BeTemporally(">=", start.Add(time.Second)))
			Ω(job).ShouldNot(BeNil())
			Ω(job.Success()).Should(BeTrue())
		})

		It("should execute a function after the job is finished", func() {
			itRun := false
			jobid := ""
			job := wf.Run("sleep", "0").Then(func(job drmaa2interface.Job) {
				itRun = true
				jobid = job.GetID()
			})
			Ω(job).ShouldNot(BeNil())
			Ω(itRun).Should(BeTrue())
			Ω(jobid).ShouldNot(Equal(""))
		})

		It("should execute a job after the job is finished", func() {
			job := wf.Run("sleep", "0").ThenRun("sleep", "0").Wait()
			Ω(job).ShouldNot(BeNil())

			job = wf.Run("sleep", "0").ThenRunT(jobT).Wait()
			Ω(job).ShouldNot(BeNil())
		})

		onSuccessRunFunction := func(command, arg string) (*wfl.Job, string) {
			jid := ""
			job := wf.Run(command, arg).OnSuccess(func(job drmaa2interface.Job) {
				jid = job.GetID()
			})
			return job, jid
		}

		It("should execute a function after a successful run of another job", func() {
			job, jid := onSuccessRunFunction("sleep", "0")
			Ω(job).ShouldNot(BeNil())
			Ω(jid).ShouldNot(Equal(""))
		})

		It("should not execute a function after a failed run of another job", func() {
			job, jid := onSuccessRunFunction("date", "invalidformat")
			Ω(job).ShouldNot(BeNil())
			Ω(jid).Should(Equal(""))
		})

		onSuccessRun := func(command, arg string) *wfl.Job {
			job := wf.Run(command, arg).OnSuccessRun("sleep", "0")
			return job
		}

		It("should execute a job after a successful run of another job", func() {
			job := onSuccessRun("sleep", "0")
			Ω(job).ShouldNot(BeNil())
			Ω(job.Template().RemoteCommand).ShouldNot(Equal("date"))
		})

		It("should execute a job after a successful run of another job", func() {
			job := onSuccessRun("date", "invalidformat")
			Ω(job).ShouldNot(BeNil())
			Ω(job.Template().RemoteCommand).Should(Equal("date"))
		})

		It("should detect when one job failed", func() {
			failed := wf.Run("sleep", "0").Run("date", "invalidformat").Run("sleep", "0").Synchronize().AnyFailed()
			Ω(failed).Should(BeTrue())
			failed = wf.Run("sleep", "0").Run("sleep", "0").Run("sleep", "0").Synchronize().AnyFailed()
			Ω(failed).Should(BeFalse())
		})

		It("should block with After", func() {
			start := time.Now()
			wf.Run("sleep", "0").After(time.Millisecond * 500)
			Ω(time.Now()).Should(BeTemporally(">=", start.Add(time.Millisecond*500)))
		})

		onFailureRun := func(command, arg string) *wfl.Job {
			job := wf.Run(command, arg).OnFailureRun("sleep", "0")
			return job
		}

		It("should execute a job after a failed run of another job", func() {
			job := onFailureRun("date", "invalidformat")
			Ω(job).ShouldNot(BeNil())
			Ω(job.ExitStatus()).Should(BeNumerically("==", 0))
			Ω(job.Template().RemoteCommand).Should(Equal("sleep"))
		})

		It("should not execute a job after a successful run of another job", func() {
			job := onFailureRun("sleep", "1")
			Ω(job).ShouldNot(BeNil())
			Ω(job.Template().Args[0]).Should(Equal("1"))
		})

		It("should execute OnSuccess() and OnFailure() properly with Do() before", func() {
			jobID := ""
			x := ""

			job := wf.Run("./test_scripts/exit.sh", "2").Do(func(j drmaa2interface.Job) {
				jobID = j.GetID()
			}).OnSuccess(func(j drmaa2interface.Job) {
				x = "success"
			}).OnFailure(func(j drmaa2interface.Job) {
				x = "failure"
			}).OnError(func(err error) {
				x = "error"
			})

			Ω(x).Should(Equal("failure"))
			Ω(job.ExitStatus()).Should(BeNumerically("==", 2))
			Ω(jobID).ShouldNot(Equal(""))
		})

		It("should execute OnSuccess() and OnFailure() properly", func() {
			x := ""
			job := wf.Run("./test_scripts/exit.sh", "2").OnSuccess(func(f drmaa2interface.Job) {
				x = "success"
			}).OnFailure(func(f drmaa2interface.Job) {
				x = "failure"
			}).OnError(func(e error) {
				x = "error"
			})

			Ω(x).Should(Equal("failure"))
			Ω(job.ExitStatus()).Should(BeNumerically("==", 2))

			x = ""
			wf.Run("./test_scripts/exit.sh", "2").OnFailure(func(f drmaa2interface.Job) {
				x = "failure"
			}).OnError(func(e error) {
				x = "error"
			}).OnSuccess(func(f drmaa2interface.Job) {
				x = "success"
			})

			Ω(x).Should(Equal("failure"))
			Ω(job.ExitStatus()).Should(BeNumerically("==", 2))

			x = ""
			wf.Run("./test_scripts/exit.sh", "2").OnError(func(e error) {
				x = "error"
			}).OnFailure(func(f drmaa2interface.Job) {
				x = "failure"
			}).OnSuccess(func(f drmaa2interface.Job) {
				x = "success"
			})

			Ω(x).Should(Equal("failure"))
			Ω(job.ExitStatus()).Should(BeNumerically("==", 2))

		})

		It("should display that the job failed when the exit code is 2", func() {
			job := onFailureRun("./test_scripts/exit.sh", "2")
			Ω(job).ShouldNot(BeNil())
			Ω(job.Template().Args[0]).Should(Equal("0"))

			job = onSuccessRun("./test_scripts/exit.sh", "2")
			Ω(job).ShouldNot(BeNil())
			Ω(job.Template().Args[0]).Should(Equal("2"))
		})

		onFailureRunFunction := func(command, arg string) (*wfl.Job, string) {
			jobid := ""
			job := wf.Run(command, arg).OnFailure(func(job drmaa2interface.Job) {
				jobid = job.GetID()
			})
			return job, jobid
		}

		It("should execute a job after a failed run of another job", func() {
			job, jobid := onFailureRunFunction("date", "invalidformat")
			Ω(job).ShouldNot(BeNil())
			Ω(job.Template().RemoteCommand).Should(Equal("date"))
			Ω(jobid).ShouldNot(Equal(""))
		})

		It("should not execute a job after a successful run of another job", func() {
			job, jobid := onFailureRunFunction("sleep", "1")
			Ω(job).ShouldNot(BeNil())
			Ω(job.Template().RemoteCommand).Should(Equal("sleep"))
			Ω(jobid).Should(Equal(""))
		})

		It("should Do() a function after a successful job submission", func() {
			jobid := ""
			job := wf.Run("sleep", "0").Do(func(j drmaa2interface.Job) {
				jobid = j.GetID()
			})
			Ω(job).ShouldNot(BeNil())
			Ω(jobid).ShouldNot(Equal(""))
		})

		It("should show that the job is failed", func() {
			success := wf.Run("date", "unknownformat").Wait().Success()
			Ω(success).Should(BeFalse())
		})

		It("should show that the job is not failed", func() {
			success := wf.Run("sleep", "0").Wait().Success()
			Ω(success).Should(BeTrue())
		})

		It("should return the job ID of the previously submitted job", func() {
			job := wf.Run("sleep", "0")
			Ω(job).ShouldNot(BeNil())
			id := job.JobID()
			Ω(id).ShouldNot(Equal(""))
			Ω(job.LastError()).Should(BeNil())
		})

		It("should list failed jobs", func() {
			job := wf.Run("sleep", "0").
				Run("date", "unknownformat").
				Run("sleep", "0").
				Run("date", "unknownformat")
			failed := job.ListAllFailed()
			Ω(len(failed)).Should(BeNumerically("==", 2))

			job = wf.Run("sleep", "0")
			failed = job.ListAllFailed()
			Ω(len(failed)).Should(BeNumerically("==", 0))

		})

		It("should signal if there is any failed jobs", func() {
			job := wf.Run("sleep", "0").
				Run("date", "unknownformat").
				Run("sleep", "0").
				Run("date", "unknownformat")
			Ω(job.HasAnyFailed()).Should(BeTrue())

			job = wf.Run("sleep", "0").
				Run("sleep", "0")
			Ω(job.HasAnyFailed()).Should(BeFalse())
		})

		It("should retry any failed jobs", func() {
			job := wf.Run("./test_scripts/randfail.sh").
				Run("./test_scripts/randfail.sh").
				Run("./test_scripts/randfail.sh").
				Run("./test_scripts/randfail.sh").
				Run("./test_scripts/randfail.sh").
				Run("./test_scripts/randfail.sh").
				Run("./test_scripts/randfail.sh").
				Run("./test_scripts/randfail.sh")

			job.RetryAnyFailed(1)
			interation := 0
			for len(job.ListAllFailed()) > 0 {
				fmt.Printf("retry failed jobs (%d)\n", interation)
				interation++
				job.RetryAnyFailed(1)
			}
			job.ReapAll()
		})

		It("should list all tasks as DRMAA2 jobs", func() {
			job := wf.Run("sleep", "0.1").
				Run("sleep", "0.1").
				Run("sleep", "0.1").
				Run("sleep", "0.1").
				Run("sleep", "0.1")
			jobs := job.ListAll()
			Expect(len(jobs)).Should(BeNumerically("==", 5))
			job.ReapAll()
		})

		Context("JobInfo related functions", func() {
			It("should return a JobInfo on success", func() {
				ji := wf.Run("sleep", "0").Wait().JobInfo()
				Ω(ji).ShouldNot(BeNil())
			})

			It("should return a JobInfo when failed", func() {
				ji := wf.Run("date", "unknownformat").Wait().JobInfo()
				Ω(ji).ShouldNot(BeNil())
			})

			It("should return a JobInfo when running", func() {
				job := wf.Run("sleep", "1")
				ji := job.JobInfo()
				Ω(ji).ShouldNot(BeNil())
				job.Kill()
			})

			It("should return JobInfos with one job", func() {
				ji := wf.Run("sleep", "0").Wait().JobInfos()
				Ω(ji).ShouldNot(BeNil())
				Ω(len(ji)).Should(BeNumerically("==", 1))
			})
		})

		Context("Job output related functions", func() {

			It("should return the output of a job", func() {
				os.Remove("tmp.db")
				ctx := wfl.NewProcessContext().WithDefaultJobTemplate(
					drmaa2interface.JobTemplate{
						OutputPath: wfl.RandomFileNameInTempDir() + "-{{.ID}}",
					},
				)
				flow := wfl.NewWorkflow(ctx)
				Ω(wf.HasError()).Should(BeFalse())

				job := flow.Run("echo", "hello")
				Ω(job.Output()).Should(Equal("hello"))

				// {{ .ID }} should be replaced by an internal number
				Ω(job.Template().OutputPath).ShouldNot(HaveSuffix("-{{.ID}}"))
			})

		})

	})

	Context("Job Array", func() {

		var (
			flow *wfl.Workflow
		)

		BeforeEach(func() {
			flow = makeWfl()
		})

		It("should run a bunch of jobs", func() {
			job := flow.RunArrayJob(1, 10, 1, 5, "sleep", "1").Wait()
			Ω(job.Success()).Should(BeTrue())
		})

		It("should run a bunch of failing jobs", func() {
			job := flow.RunArrayJob(1, 10, 1, 5, "/bin/bash", "-c", "exit 77").Wait()
			Ω(job.State().String()).Should(Equal(drmaa2interface.Failed.String()))
			Ω(job.Success()).Should(BeFalse())
		})

		It("should run a bunch of jobs with a job template", func() {
			job := flow.NewJob().RunArrayT(1, 10, 1, 5, drmaa2interface.JobTemplate{
				RemoteCommand: "sleep",
				Args:          []string{"0"},
			}).Wait()
			Ω(job.JobID()).ShouldNot(Equal(""))
			Ω(job.Success()).Should(BeTrue())
			Ω(job.Template().RemoteCommand).Should(Equal("sleep"))
			Ω(job.ReapAll().Errored()).Should(BeFalse())
			job.ThenRunArray(1, 10, 1, 5, "sleep", "0").Wait()
			Ω(job.Success()).Should(BeTrue())
		})

	})

	Context("Job Matrix", func() {

		var (
			flow              *wfl.Workflow
			getRemoteCommands func(drmaa2interface.Job, interface{}) error
		)

		BeforeEach(func() {
			flow = makeWfl()

			// Function which copies the remote command from the
			// job's job template to the interface which is expected
			// to be a pointer to a string slice (cmdList).
			getRemoteCommands = func(j drmaa2interface.Job, i interface{}) error {
				output := i.(*[]string)
				jt, err := j.GetJobTemplate()
				if err != nil {
					return err
				}
				*output = append(*output, jt.RemoteCommand)
				return nil
			}
		})

		It("should run a job matrix", func() {
			job := flow.NewJob().RunMatrixT(
				drmaa2interface.JobTemplate{
					RemoteCommand: "{{cmd}}",
					Args:          []string{"{{arg}}"},
				},
				wfl.Replacement{
					Fields:       []wfl.JobTemplateField{wfl.RemoteCommand},
					Pattern:      "{{cmd}}",
					Replacements: []string{"sleep", "echo"},
				},
				wfl.Replacement{
					Fields:       []wfl.JobTemplateField{wfl.Args},
					Pattern:      "{{arg}}",
					Replacements: []string{"0.1", "0.2"},
				},
			)
			// there should be no submission errors
			Expect(job.Errored()).Should(BeFalse())
			// wait for all jobs finished
			job.Synchronize()
			Expect(job.HasAnyFailed()).Should(BeFalse())
			jis := job.JobInfos()
			Expect(jis).NotTo(BeNil())
			Expect(len(jis)).Should(BeNumerically("==", 4))

			// get all job template commands
			cmdList := make([]string, 0, 4)
			Expect(job.ForEach(getRemoteCommands, &cmdList)).To(Succeed())
			Expect(cmdList).To(ConsistOf("sleep", "echo", "sleep", "echo"))

		})

		It("should run a job matrix with only one dimension", func() {
			job := flow.NewJob().RunMatrixT(
				drmaa2interface.JobTemplate{
					RemoteCommand: "sleep",
					Args:          []string{"{{arg}}"},
				},
				wfl.Replacement{
					Fields:       []wfl.JobTemplateField{wfl.Args},
					Pattern:      "{{arg}}",
					Replacements: []string{"0.1", "0.2", "0.3"},
				},
				wfl.Replacement{}, // leave empty
			)

			// there should be no submission errors
			Expect(job.Errored()).Should(BeFalse())
			// wait for all jobs finished
			job.Synchronize()
			Expect(job.HasAnyFailed()).Should(BeFalse())
			jis := job.JobInfos()
			Expect(jis).NotTo(BeNil())
			Expect(len(jis)).Should(BeNumerically("==", 3))
		})

		It("should run nothing when no replacements are specified", func() {
			job := flow.NewJob().RunMatrixT(
				drmaa2interface.JobTemplate{
					RemoteCommand: "sleep",
					Args:          []string{"1"},
				},
				wfl.Replacement{}, // leave empty
				wfl.Replacement{}, // leave empty
			)
			// there should be no submission errors
			Expect(job.Errored()).Should(BeFalse())
			// wait for all jobs finished
			job.Synchronize()
			Expect(job.HasAnyFailed()).Should(BeFalse())
			jis := job.JobInfos()
			Expect(jis).NotTo(BeNil())
			Expect(len(jis)).Should(BeNumerically("==", 0))
		})

	})

	Context("Job ouput", func() {

		It("should return the output for all job IDs", func() {
			flow := wfl.NewWorkflow(wfl.NewProcessContextByCfg(
				wfl.ProcessConfig{
					DefaultTemplate: drmaa2interface.JobTemplate{
						// OutputPath is required to set to a unique
						// file for each job. That is why we use
						// RandomFileNameInTempDir() here.
						OutputPath: wfl.RandomFileNameInTempDir(),
					},
				},
			))
			job := flow.Run("echo", "foo").Resubmit(4)
			job.Synchronize()

			outputMap := job.OutputsForJobIDs(nil)
			Ω(outputMap).Should(HaveLen(5))
			Ω(outputMap[job.JobID()]).Should(Equal("foo"))
			for _, j := range job.JobInfos() {
				Ω(outputMap[j.ID]).Should(Equal("foo"))
			}

			// test with a filter
			outputMap = job.OutputsForJobIDs([]string{job.JobID()})
			Ω(outputMap).Should(HaveLen(1))
			Ω(outputMap[job.JobID()]).Should(Equal("foo"))

			// test with a filter which does not match
			outputMap = job.OutputsForJobIDs([]string{"not-existing"})
			Ω(outputMap).Should(HaveLen(0))
		})
	})

	Context("ForEach and ForAll", func() {

		It("should run ForEach and ForAll on all jobs", func() {

			flow := wfl.NewWorkflow(wfl.NewProcessContext())

			job := flow.Run("echo", "foo").Resubmit(4).Synchronize()

			outputs := sync.Map{}

			collectJobIDs := func(j drmaa2interface.Job, i interface{}) error {
				ids := i.(*sync.Map)
				ids.Store(j.GetID(), true)
				return nil
			}

			// runs in parallel
			job.ForAll(collectJobIDs, &outputs)

			for _, ji := range job.JobInfos() {
				_, ok := outputs.Load(ji.ID)
				Ω(ok).Should(BeTrue())
			}

			// runs sequentially
			outputs2 := make(map[string]bool)

			collectJobIDs2 := func(j drmaa2interface.Job, i interface{}) error {
				ids := i.(*map[string]bool)
				(*ids)[j.GetID()] = true
				return nil
			}

			err := job.ForEach(collectJobIDs2, &outputs2)
			Ω(err).Should(BeNil())
			// len should be 5
			Ω(len(outputs2)).Should(BeNumerically("==", 5))

			// compare both maps
			outputs.Range(func(key, value interface{}) bool {
				val, ok := outputs2[key.(string)]
				Ω(ok).Should(BeTrue())
				Ω(val).Should(BeTrue())
				return true
			})

		})

		It("should run ForAll on all jobs in parallel", func() {
			flow := wfl.NewWorkflow(wfl.NewProcessContext())

			job := flow.Run("echo", "foo").Resubmit(4).Synchronize()

			timeConsumingFunction := func(j drmaa2interface.Job, i interface{}) error {
				<-time.Tick(50 * time.Millisecond)
				return nil
			}

			// runs in parallel
			start := time.Now()
			job.ForAll(timeConsumingFunction, nil)
			Ω(time.Since(start)).Should(BeNumerically("<=", 150*time.Millisecond))

			start = time.Now()
			job.ForEach(timeConsumingFunction, nil)
			Ω(time.Since(start)).Should(BeNumerically(">=", 250*time.Millisecond))
		})

	})

	Context("Basic error cases", func() {

		It("should error when no workflow is defined in a job", func() {
			job := wfl.EmptyJob().Run("sleep", "0")
			err := job.LastError()

			Ω(err).ShouldNot(BeNil())
			Ω(err.Error()).Should(ContainSubstring("no workflow defined"))
			Ω(job.Errored()).Should(BeTrue())

			job.Wait()
		})

		It("should error on error during OnSuccess()", func() {
			job := wfl.NewWorkflow(nil).Run("sleep", "0").OnSuccess(func(j drmaa2interface.Job) {})
			err := job.LastError()
			Ω(job).ShouldNot(BeNil())
			Ω(err).ShouldNot(BeNil())
			Ω(job.Errored()).Should(BeTrue())
		})

		It("should error when there is no JobTemplate", func() {
			emptyJob := wfl.EmptyJob()
			tmpl := emptyJob.Template()
			Ω(tmpl).Should(BeNil())
			Ω(emptyJob.LastError()).ShouldNot(BeNil())
			Ω(emptyJob.Errored()).Should(BeTrue())
		})

		It("should error when no context is defined in a job", func() {
			ewfl := wfl.NewWorkflow(nil)
			Ω(ewfl.HasError()).Should(BeTrue())

			job := ewfl.Run("sleep", "0")
			err := job.LastError()

			Ω(err).ShouldNot(BeNil())
			Ω(err.Error()).Should(ContainSubstring("no context defined"))
			Ω(job.Errored()).Should(BeTrue())
		})

		It("should error when suspending, resuming or killing an empty job", func() {
			job := wfl.EmptyJob().Suspend()
			Ω(job.LastError()).ShouldNot(BeNil())
			Ω(job.LastError().Error()).Should(ContainSubstring("job task not available"))

			job.Resume()
			Ω(job.Errored()).Should(BeTrue())
			Ω(job.LastError()).ShouldNot(BeNil())
			Ω(job.LastError().Error()).Should(ContainSubstring("job task not available"))

			job.Kill()
			Ω(job.Errored()).Should(BeTrue())
			Ω(job.LastError()).ShouldNot(BeNil())
			Ω(job.LastError().Error()).Should(ContainSubstring("job task not available"))

		})

		It("should error when getting the state for an empty job", func() {
			job := wfl.EmptyJob()
			state := job.State()
			Ω(job.Errored()).Should(BeTrue())
			Ω(job.LastError()).ShouldNot(BeNil())
			Ω(job.LastError().Error()).Should(ContainSubstring("job task not available"))
			Ω(state).Should(BeNumerically("==", drmaa2interface.Undetermined))
		})

		It("should error when getting a template for an empty job", func() {
			job := wfl.EmptyJob()
			template := job.Template()
			Ω(job.Errored()).Should(BeTrue())
			Ω(job.LastError()).ShouldNot(BeNil())
			Ω(job.LastError().Error()).Should(ContainSubstring("job task not available"))
			Ω(template).Should(BeNil())
		})

		It("should error when resubmit is done for an empty job", func() {
			job := wfl.EmptyJob().Resubmit(10)
			Ω(job.Errored()).Should(BeTrue())
			Ω(job.LastError()).ShouldNot(BeNil())
			Ω(job.LastError().Error()).Should(Equal("job not available"))
		})

		It("should return an empty job ID string in case of an error", func() {
			job := wfl.EmptyJob()
			Ω(job.JobID()).Should(Equal(""))
		})

		It("should error when running a job without a command", func() {
			wf := makeWfl()
			job := wfl.NewJob(wf).RunT(drmaa2interface.JobTemplate{RemoteCommand: ""})
			Ω(job.LastError()).ShouldNot(BeNil())
		})

		It("should execute a function OnError()", func() {
			var err error
			wfl.EmptyJob().Suspend().OnError(func(e error) { err = e })
			Ω(err).ShouldNot(BeNil())
		})

		It("should execute a function OnError() on job submission error", func() {
			var err error
			wfl.EmptyJob().Run("").OnError(func(e error) { err = e })
			Ω(err).ShouldNot(BeNil())
		})

		It("should not return the JobInfo in case of an error", func() {
			job := wfl.EmptyJob()
			job.JobInfo()
			err := job.LastError()
			Ω(err).ShouldNot(BeNil())
		})

		It("should return ExitStatus() -1 when no job is defined", func() {
			exit := wfl.EmptyJob().ExitStatus()
			Ω(exit).Should(BeNumerically("==", -1))
		})

		It("should error at Then() in case when prev. job is not found", func() {
			job := wfl.EmptyJob().Then(func(j drmaa2interface.Job) {})
			Ω(job.LastError().Error()).Should(ContainSubstring("task not available"))
		})

		It("should use default JobTemplate settings from Process Context", func() {
			template := drmaa2interface.JobTemplate{
				JobName:    "jobname",
				OutputPath: "/dev/stdout",
				ErrorPath:  "/dev/stderr",
			}
			flow := wfl.NewWorkflow(wfl.NewProcessContextByCfg(wfl.ProcessConfig{
				DefaultTemplate: template,
			}))
			job := flow.Run("sleep", "0").Wait()
			rt := job.Template()
			Ω(rt.JobName).Should(Equal("jobname"))
			Ω(rt.OutputPath).Should(Equal("/dev/stdout"))
			Ω(rt.ErrorPath).Should(Equal("/dev/stderr"))

			job.Resubmit(1).Synchronize()
			rt = job.Template()
			Ω(rt.JobName).Should(Equal("jobname"))
			Ω(rt.OutputPath).Should(Equal("/dev/stdout"))
			Ω(rt.ErrorPath).Should(Equal("/dev/stderr"))
		})

	})

	Context("Use Cases", func() {

		It("should return the right exit code", func() {
			var err error
			var exitStatus int

			job := makeWfl().Run("./test_scripts/exit.sh", "13")
			job.OnError(func(e error) { err = e })
			Ω(err).Should(BeNil())

			job.OnSuccess(func(j drmaa2interface.Job) { err = errors.New("should have failed") })
			Ω(err).Should(BeNil())
			Ω(job.ExitStatus()).Should(BeNumerically("==", 13))

			job.OnFailure(func(j drmaa2interface.Job) {
				ji, errJi := j.GetJobInfo()
				err = errJi
				exitStatus = ji.ExitStatus
			})
			Ω(err).Should(BeNil())
			Ω(exitStatus).Should(BeNumerically("==", 13))
		})

	})

})
