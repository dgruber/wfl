package wfl

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/mitchellh/copystructure"
)

// mergeJobTemplateWithDefaultTemplate adds requests from _def_ into _req_
// if specified in req
//
// Note there is no "unset" convention yet for the job template. It will
// only take actually implemented (in drmaa2os job tracker) settings into
// account.
func mergeJobTemplateWithDefaultTemplate(req, def drmaa2interface.JobTemplate) drmaa2interface.JobTemplate {
	if req.JobCategory == "" {
		req.JobCategory = def.JobCategory
	}
	if req.InputPath == "" {
		req.InputPath = def.InputPath
	}
	if req.OutputPath == "" {
		req.OutputPath = def.OutputPath
	}
	if req.ErrorPath == "" {
		req.ErrorPath = def.ErrorPath
	}
	if req.AccountingID == "" {
		req.AccountingID = def.AccountingID
	}
	if req.JobName == "" {
		req.JobName = def.JobName
	}
	// replaces destination machines
	if req.CandidateMachines == nil && def.CandidateMachines != nil {
		if cm, err := copystructure.Copy(def.CandidateMachines); err == nil {
			req.CandidateMachines = cm.([]string)
		}
	}
	// replace extensions
	if req.ExtensionList == nil && def.ExtensionList != nil {
		if el, err := copystructure.Copy(def.ExtensionList); err == nil {
			req.ExtensionList = el.(map[string]string)
		}
	}
	// join files to stage
	req.StageInFiles = mergeStringMap(req.StageInFiles, def.StageInFiles)
	// join enviroment variables
	req.JobEnvironment = mergeStringMap(req.JobEnvironment, def.JobEnvironment)
	// TODO implement more when required
	return req
}

func mergeStringMap(dst, src map[string]string) map[string]string {
	if src != nil {
		if dst == nil {
			dst = make(map[string]string, len(src))
		}
		for k, v := range src {
			if dst[k] == "" {
				dst[k] = v
			}
		}
	}
	return dst
}
