package kubernetestracker

import (
	"errors"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientBatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
)

func getJobInterfaceAndJob(kt *kubernetes.Clientset, jobid string) (clientBatchv1.JobInterface, *batchv1.Job, error) {
	if kt == nil {
		return nil, nil, errors.New("no clientset")
	}
	jc, err := getJobsClient(kt)
	if err != nil {
		return nil, nil, fmt.Errorf("can't get k8s client: %s", err.Error())
	}
	job, err := getJobByID(jc, jobid)
	if err != nil {
		return nil, nil, fmt.Errorf("can't find job: %s", err.Error())
	}
	return jc, job, nil
}

func jobStateChange(jc clientBatchv1.JobInterface, job *batchv1.Job, action string) error {
	if jc == nil || job == nil {
		return errors.New("internal error: can't change job status: job is nil")
	}
	switch action {
	case "suspend":
		return errors.New("Unsupported Operation")
	case "resume":
		return errors.New("Unsupported Operation")
	case "hold":
		return errors.New("Unsupported Operation")
	case "release":
		return errors.New("Unsupported Operation")
	case "terminate":
		// activeDeadlineSeconds to zero
		return jc.Delete(job.GetName(), &k8sapi.DeleteOptions{})
	}
	return fmt.Errorf("Undefined job operation")
}

func deleteJob(jc clientBatchv1.JobInterface, job *batchv1.Job) error {
	if jc == nil || job == nil {
		return errors.New("internal error: can't delete job: job is nil")
	}
	policy := k8sapi.DeletePropagationBackground
	return jc.Delete(job.GetName(), &k8sapi.DeleteOptions{PropagationPolicy: &policy})
}

func getJobByID(jc clientBatchv1.JobInterface, jobid string) (*batchv1.Job, error) {
	return jc.Get(jobid, k8sapi.GetOptions{})
}
