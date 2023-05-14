package wfl

const PromptTemplateOutputTransform string = `Transform a given text with the instructions you
get. FOLLOW the instructions VERY carefully! ONLY return the transformed text. Do NOT include
the instructions. The instructions are so important that you should read them at least 3 times!
Find the very best answer!

Here is the output text: {{.Output}}

Here is the task: {{.Task}}

Your answer:`

const PromptTemplateErrorTransform string = `Transform the following error message with the instructions you
get. FOLLOW the instructions VERY carefully! ONLY return the transformed error message. Do NOT include
the instructions. The instructions are so important that you should read them at least 3 times!

Here is the error message: {{.Output}}

Here is the task: {{.Task}}

Your answer:`

const PromptTemplateLinuxShellScript string = `Write a shell script which is executed by the Linux bash. ONLY return
the bash script. Do not include the shebang line!! Always write the final output to stdout!

Here is the task: {{.Prompt}}

Write a very good bash script which runs on Linux by using default tools. NEVER use any options of commands which are NOT available on Linux
or are expected to be downloaded firt.
Do not include #! line. Do not include any shebang line! NO empty lines:
`

const PromptTemplateDarwinShellScript string = `Write a shell script which is executed by the Darwin shell. ONLY return
the shell script. Do not include the shebang line!! Always write the final output to stdout.

Here is the task: {{.Prompt}}

Write a very good shell script which runs on macOS. NEVER use any options of commands which are NOT available on macOS.
Do not include #! line. Do not include any shebang line! No empty lines:
`

const PromptTemplatePythonScript string = `Write a simple and clean python script. ONLY return
the python script. It is executed with python -c.

Here is the task: {{.Prompt}}

Write a very good Python script without any dependencies. It must be perfectly valid Python code. It must be complete:
`

const PromptJobTemplateConstructorTemplate string = `Write a JobTemplate JSON. ONLY return
the JobTemplate JSON. The JobTemplate JSON is for running shell scripts as processes.

A JobTemplate is mapped into the process creation process in the following way:

JSON JobTemplate -> OS Process
RemoteCommand -> Executable to start
JobName must be unique
Args -> Arguments of the executable
WorkingDir -> Working directory
JobEnvironment -> Environment variables for the process
InputPath -> if set it is used as stdin for the process
OutputPath -> if set the stdout of the process is written to this file; do NOT redirect in args if you want to have Output()
ErrorPath -> if set the stderr of the process is written to this file

AVOID redirecting output in the script! Use OutputPath instead! Set output to a unique file name 
in temp if not instructed otherwise!

Examples:

{"extension":{},"remoteCommand":"./plus.sh","jobName":"testjob","inputPath":"in.txt","outputPath":"out.txt"}

{"extension":{},"remoteCommand":"/usr/bin/sort","jobName":"sort","inputPath":"/etc/services","outputPath":"/dev/stdout"}

{"extension":{},"remoteCommand":"/bin/bash","args":["-c","echo $JOB_ID"],"outputPath":"/tmp/outputfile.txt"}

{"extension":{},"remoteCommand":"/bin/bash","args":["-c","echo $JOB_ID \u0026\u0026 echo $MYVAR"],"jobEnvironment":{"MYVAR":"myvalue"},"outputPath":"/tmp/outputfile.txt"}

{
	"extension": {},
	"remoteCommand": "/bin/bash",
	"args": ["-c", "ps aux | head -n 6"],
	"jobName": "memory-usage",
	"workingDir": "/tmp",
	"outputPath": "/tmp/memory-output.txt",
	"inputPath": "",
	"errorPath": "",
  }

Here is your task: {{.Prompt}}

Create a JobTemplate JSON:
`
