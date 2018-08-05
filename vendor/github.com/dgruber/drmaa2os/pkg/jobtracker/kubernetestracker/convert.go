package kubernetestracker

import (
	"errors"
	"fmt"
	"github.com/dgruber/drmaa2interface"
	batchv1 "k8s.io/api/batch/v1"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

func newVolumes(jt drmaa2interface.JobTemplate) ([]k8sv1.Volume, error) {
	//v := k8sv1.Volume{}
	return nil, nil
}

func newContainers(jt drmaa2interface.JobTemplate) ([]k8sv1.Container, error) {
	if jt.JobCategory == "" {
		return nil, errors.New("JobCategory (image name) not set in JobTemplate")
	}
	if jt.RemoteCommand == "" {
		return nil, errors.New("RemoteCommand not set in JobTemplate")
	}
	c := k8sv1.Container{
		Name:       jt.JobName,
		Image:      jt.JobCategory,
		Command:    []string{jt.RemoteCommand},
		Args:       jt.Args,
		WorkingDir: jt.WorkingDirectory,
	}

	// spec.template.spec.containers[0].name: Required value"
	if jt.JobName == "" {
		c.Name = "drmaa2osstandardcontainer"
	}

	// if len(jt.CandidateMachines) == 1 {
	//	c = jt.CandidateMachines[0]
	// }
	return []k8sv1.Container{c}, nil
}

func newNodeSelector(jt drmaa2interface.JobTemplate) (map[string]string, error) {
	return nil, nil
}

/* 	deadlineTime returns the deadline of the job as int pointer converting from
    AbsoluteTime to a relative time.
	"
	Specifies a deadline after which the implementation or the DRM system SHOULD change the job state to
		any of the “Terminated” states (see Section 8.1).
    	The support for this attribute is optional, as expressed by the
       	- DrmaaCapability::JT_DEADLINE
		DeadlineTime is defined as AbsoluteTime.
	"
*/
func deadlineTime(jt drmaa2interface.JobTemplate) (*int64, error) {
	var deadline int64
	if !jt.DeadlineTime.IsZero() {
		if jt.DeadlineTime.After(time.Now()) {
			deadline = jt.DeadlineTime.Unix() - time.Now().Unix()
		} else {
			return nil, fmt.Errorf("deadlineTime (%s) in job template is in the past", jt.DeadlineTime.String())
		}
	}
	return &deadline, nil
}

// https://godoc.org/k8s.io/api/core/v1#PodSpec
// https://github.com/kubernetes/kubernetes/blob/886e04f1fffbb04faf8a9f9ee141143b2684ae68/pkg/api/types.go
func newPodSpec(v []k8sv1.Volume, c []k8sv1.Container, ns map[string]string, activeDeadline *int64) k8sv1.PodSpec {
	spec := k8sv1.PodSpec{
		Volumes:       v,
		Containers:    c,
		NodeSelector:  ns,
		RestartPolicy: "Never",
	}
	if *activeDeadline > 0 {
		spec.ActiveDeadlineSeconds = activeDeadline
	}
	return spec
}

func addExtensions(job *batchv1.Job, jt drmaa2interface.JobTemplate) *batchv1.Job {
	if jt.ExtensionList == nil {
		return job
	}
	if jt.ExtensionList["namespace"] != "" {
		//Namespace: v1.NamespaceDefault
		job.Namespace = jt.ExtensionList["namespace"]
	}
	if jt.ExtensionList["labels"] != "" {
		// "key=value,key=value,..."
		for _, labels := range strings.Split(jt.ExtensionList["labels"], ",") {
			l := strings.Split(labels, "=")
			if len(l) == 2 {
				if l[0] == "drmaa2jobsession" {
					continue // don't allow to override job session
				}
				job.Labels[l[0]] = l[1]
			}
		}
	}
	return job
}

func convertJob(jobsession string, jt drmaa2interface.JobTemplate) (*batchv1.Job, error) {
	volumes, err := newVolumes(jt)
	if err != nil {
		return nil, fmt.Errorf("converting job (newVolumes): %s", err)
	}
	containers, err := newContainers(jt)
	if err != nil {
		return nil, fmt.Errorf("converting job (newContainer): %s", err)
	}
	nodeSelector, err := newNodeSelector(jt)
	if err != nil {
		return nil, fmt.Errorf("converting job (newNodeSelector): %s", err)
	}

	// settings for command etc.
	dl, err := deadlineTime(jt)
	if err != nil {
		return nil, err
	}
	podSpec := newPodSpec(volumes, containers, nodeSelector, dl)

	var one int32 = 1
	job := batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		// Standard object's metadata.
		// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
		// +optional
		ObjectMeta: metav1.ObjectMeta{
			Name:         jt.JobName,
			Labels:       map[string]string{"drmaa2jobsession": jobsession},
			GenerateName: "drmaa2os",
		},
		// Specification of the desired behavior of a job.
		// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
		// +optional
		Spec: batchv1.JobSpec{
			/*ManualSelector: ,
			Selector: &unversioned.LabelSelector{
				MatchLabels: options.labels,
			}, */
			Parallelism:  &one,
			Completions:  &one,
			BackoffLimit: &one,

			// Describes the pod that will be created when executing a job.
			// More info: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/
			Template: k8sv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:         "drmaa2osjob",
					GenerateName: "drmaa2os",
					//Labels: options.labels,
				},
				Spec: podSpec,
			},
		},
	}
	return addExtensions(&job, jt), nil
}
