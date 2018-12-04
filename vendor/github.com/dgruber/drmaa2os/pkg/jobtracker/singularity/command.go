package singularity

import (
	"github.com/dgruber/drmaa2interface"
)

// createProcessJobTemplate converts the JobTemplate into a JobTemplate for the
// OS process implementation, i.e. it injects Singularity options from the JobTemplate
// extension map and uses Singularity as RemoteCommand.
func createProcessJobTemplate(st drmaa2interface.JobTemplate) drmaa2interface.JobTemplate {
	st.RemoteCommand, st.Args = createCommandAndArgs(st)
	return st
}

func setBooleanExtension(options []string, extensions map[string]string, name string) []string {
	if ext, exists := extensions[name]; exists {
		if ext != "FALSE" && ext != "false" {
			options = append(options, "--"+name)
		}
	}
	return options
}

func setExtension(options []string, extensions map[string]string, name string) []string {
	if ext, exists := extensions[name]; exists {
		options = append(options, "--"+name)
		options = append(options, ext)
	}
	return options
}

func createCommandAndArgs(jt drmaa2interface.JobTemplate) (string, []string) {
	options := make([]string, 0, 4)
	globalOptions := make([]string, 0, 4)

	if jt.ExtensionList != nil {
		/* global options */
		globalOptions = setBooleanExtension(globalOptions, jt.ExtensionList, "debug")
		globalOptions = setBooleanExtension(globalOptions, jt.ExtensionList, "silent")
		globalOptions = setBooleanExtension(globalOptions, jt.ExtensionList, "quite")
		globalOptions = setBooleanExtension(globalOptions, jt.ExtensionList, "verbose")

		/* exec options */
		options = setBooleanExtension(options, jt.ExtensionList, "writable")
		options = setBooleanExtension(options, jt.ExtensionList, "keep-privs")
		options = setBooleanExtension(options, jt.ExtensionList, "net")
		options = setBooleanExtension(options, jt.ExtensionList, "nv")
		options = setBooleanExtension(options, jt.ExtensionList, "overlay")
		options = setBooleanExtension(options, jt.ExtensionList, "pid")
		options = setBooleanExtension(options, jt.ExtensionList, "ipc")
		options = setBooleanExtension(options, jt.ExtensionList, "app")
		options = setBooleanExtension(options, jt.ExtensionList, "contain")
		options = setBooleanExtension(options, jt.ExtensionList, "containAll")
		options = setBooleanExtension(options, jt.ExtensionList, "userns")
		options = setBooleanExtension(options, jt.ExtensionList, "workdir")

		options = setExtension(options, jt.ExtensionList, "bind")
		options = setExtension(options, jt.ExtensionList, "add-caps")
		options = setExtension(options, jt.ExtensionList, "drop-cap")
		options = setExtension(options, jt.ExtensionList, "security")
		options = setExtension(options, jt.ExtensionList, "hostname")
		options = setExtension(options, jt.ExtensionList, "network")
		options = setExtension(options, jt.ExtensionList, "network-args")
		options = setExtension(options, jt.ExtensionList, "apply-cgroups")
		options = setExtension(options, jt.ExtensionList, "scratch")
		options = setExtension(options, jt.ExtensionList, "home")
	}
	args := []string{}
	if len(globalOptions) > 0 {
		args = append(args, globalOptions...)
	}
	args = append(args, "exec")
	if len(options) > 0 {
		args = append(args, options...)
	}
	args = append(args, jt.JobCategory, jt.RemoteCommand)
	return "singularity", append(args, jt.Args...)
}
