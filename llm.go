package wfl

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/dgruber/drmaa2interface"
	openai "github.com/sashabaranov/go-openai"
)

type llmConfig struct {
	openAPIClient *openai.Client
	runPBehavior  RunPBehavior
	model         string
}

// OpenAIConfig is used to configure the OpenAI API client to enable
// the xP() methods which expects a prompt.
type OpenAIConfig struct {
	// Token is the access token to OpenAI API (to create one:
	// https://platform.openai.com/account/api-keys)
	Token string `json:"token"`
	// RunPMethod defines how the RunP() method should be executed. Currently
	// there is a MacOS shell and Linux bash variant. The Linux shell variant
	// is the default.
	RunPMethod RunPBehavior `json:"runPBehavior"`
	// Model allows you to change from gpt-3.5-turbo-0301 to a larger model.
	// Like one with 32k tokens for processing larger job outputs.
	Model string `json:"model"`
}

// WithLLMOpenAI adds an OpenAI API client to the workflow. This enables
// the OutputP(), and ErrorP() methods defined on the Job struct and
// the TemplateP() method defined on the Workflow struct.
//
// TemplateP("prompt") creates a job template based on the description
// of the prompt. Before executing the job template it should be checked
// for its safety! This is for research (or fun)! PLEASE EXECUTE THE JOB
// TEMPLATES ONLY WITH CARE IN AN ISOLATED ENVIRONMENT! THIS CAN BE DANGEROUS!
//
// OutputP("What to do with the output") can be used to process the job
// output (internally retrieved by Output()) by using a textual description
// of the transformation applied to the output.
//
// Note, that the Output() must be available, i.e. for process workflows the
// stdout of the process must be written to a unique persistent file (OutputPath).
// Please check the Ouput() documentation for more details.
//
// ErrorP() transforms the job error (retrieved by Error()) by using a textual
// description of the transformation task. For example you can let explain
// the error and an possible solution through the LLM.
//
// Note that this is an experimental feature! It might be dropped or reworked
// in the future!
func (w *Workflow) WithLLMOpenAI(config OpenAIConfig) *Workflow {
	if config.Token == "" {
		w.workflowCreationError = fmt.Errorf("no OpenAI token given")
		return w
	}
	if config.RunPMethod == "" {
		config.RunPMethod = RunPBehaviorLinuxShellScript
	}
	w.llmConfig = &llmConfig{
		openAPIClient: openai.NewClient(config.Token),
		runPBehavior:  config.RunPMethod,
	}
	// check if token is valid by sending a first request
	_, err := w.llmConfig.openAPIClient.CreateCompletion(
		context.Background(),
		openai.CompletionRequest{
			Model:       "davinci",
			Prompt:      "Hello",
			MaxTokens:   5,
			Temperature: 0.5,
			TopP:        1,
			N:           1,
			Stream:      false,
			LogProbs:    0,
			Stop:        []string{"\n"},
		},
	)
	if err != nil {
		if w.log != nil {
			w.log.Errorf(context.Background(), "OpenAI token is not valid: %v", err)
		}
		w.workflowCreationError = fmt.Errorf("OpenAI token is not valid: %v", err)
	}
	return w
}

type TemplatePromptType int

const (
	// PromptTemplateLinuxShellScript is the default template for creating
	// a Linux shell script.
	TemplatePromptTypeLinuxShellScript TemplatePromptType = iota
	// PromptTemplateDarwinShellScript is the template for creating a macOS
	// shell script.
	TemplatePromptTypeDarwinShellScript
	// PromptTemplatePythonScript is the template for creating a Python
	// script.
	TemplatePromptTypePythonScript
)

// TemplateP creates a job template based on the description of the prompt.
// It creates a job template which executes the prompt as a shell script.
// Before executing the job template it MUST be checked for its safety!!!
// The template parameter is optional and can be used to define a template
// for the shell script. The default template is for generating a Linux
// shell script (alternative for macOS: wfl.PromptTemplateDarwinShellScript).
func (flow *Workflow) TemplateP(prompt string, templateType ...TemplatePromptType) (drmaa2interface.JobTemplate, error) {
	if flow.llmConfig == nil || flow.llmConfig.openAPIClient == nil {
		return drmaa2interface.JobTemplate{},
			fmt.Errorf("no LLM configuration given")
	}
	var promptTemplate string
	shell := "/bin/bash"

	if len(templateType) == 0 {
		promptTemplate = PromptTemplateLinuxShellScript
	} else {
		switch templateType[0] {
		case TemplatePromptTypeLinuxShellScript:
			promptTemplate = PromptTemplateLinuxShellScript
		case TemplatePromptTypeDarwinShellScript:
			promptTemplate = PromptTemplateDarwinShellScript
		case TemplatePromptTypePythonScript:
			promptTemplate = PromptTemplatePythonScript
			shell = "python3"
		default:
			return drmaa2interface.JobTemplate{},
				fmt.Errorf("unknown template type %v", templateType[0])
		}
	}
	result, err := flow.llmConfig.applyPrompt(prompt, promptTemplate)
	if err != nil {
		return drmaa2interface.JobTemplate{}, err
	}
	jt := drmaa2interface.JobTemplate{
		RemoteCommand:    shell,
		Args:             []string{"-c", result},
		OutputPath:       RandomFileNameInTempDir(),
		ErrorPath:        "/dev/stderr",
		WorkingDirectory: "/tmp",
	}
	return jt, nil
}

// OutputP applies the given prompt to the output of the previous job if
// there is any. The prompt is a textual description of the transformation
// which is applied to the output. The transformation is done by using the
// OpenAI API. Note, that the context size is limited. If the output is too
// large the output might not be useful.
//
// Examples for prompts:
// - "Translate the output in Schwäbisch (kind of German)"
// - "Create a summary of the output with max. 30 words"
// - use some output specifc questions...
func (j *Job) OutputP(prompt string) string {
	j.begin(j.ctx, "OutputP()")
	if j.wfl.llmConfig == nil || j.wfl.llmConfig.openAPIClient == nil {
		j.lastError = fmt.Errorf("no LLM configuration given")
	}

	output := j.Output()
	if j.lastError != nil {
		return ""
	}
	if output == "" {
		j.errorf(context.Background(), "no output from previous job to apply LLM prompt")
		j.lastError = fmt.Errorf("no output from previous job to apply LLM prompt")
		return ""
	}

	// create prompt
	input, err := mergeOutputAndTaskWithTemplate(output, prompt, PromptTemplateOutputTransform)
	if err != nil {
		j.errorf(context.Background(), "error merging output with template: %v", err)
		j.lastError = fmt.Errorf("error merging prompt with template: %v", err)
		return ""
	}
	model := "gpt-3.5-turbo-0301"
	if j.wfl.llmConfig.model != "" {
		model = j.wfl.llmConfig.model
	}
	resp, err := j.wfl.llmConfig.openAPIClient.CreateChatCompletion(context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: input},
			},
			MaxTokens:   1024,
			Temperature: 0.5,
			Stop:        []string{"\n\n"},
			User:        "wfl",
		},
	)
	if err != nil {
		j.lastError = fmt.Errorf("error creating completion: %v", err)
		return ""
	}
	return resp.Choices[0].Message.Content
}

// ErrorP applies a prompt to the last error message. The prompt is a textual
// description of the transformation which is applied to the error message.
// The transformation is done by using the OpenAI API. Note, that the output
// size is limited. If the output is too large the output might not be useful.
//
// Examples for prompts:
// - "What is the reason for the error?"
// - "Explain the error and provide a solution"
// - "Translate the error message into Bayerisch (kind of German)"
func (j *Job) ErrorP(prompt string) string {
	j.begin(j.ctx, "ErrorP()")
	if j.lastError == nil {
		return "There is no error visible."
	}

	input, err := mergeOutputAndTaskWithTemplate(j.lastError.Error(),
		prompt, PromptTemplateErrorTransform)
	if err != nil {
		return fmt.Sprintf(`Can not analyse original error.
Internal error merging error with prompt template: %v`, err)
	}

	model := "gpt-3.5-turbo-0301"
	if j.wfl.llmConfig.model != "" {
		model = j.wfl.llmConfig.model
	}

	resp, err := j.wfl.llmConfig.openAPIClient.CreateChatCompletion(context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: input},
			},
			MaxTokens:   512,
			Temperature: 0.1,
			User:        "wfl2",
		},
	)
	if err != nil {
		// do not override error as we want to analyze the error
		return fmt.Sprintf(`Can not analyse original error. 
Internal error creating LLM completion: %v`, err)
	}
	return resp.Choices[0].Message.Content
}

func (llmc *llmConfig) applyPrompt(prompt, promptTemplate string) (string, error) {
	// create prompt
	input, err := mergePromptWithTemplate(prompt, promptTemplate)
	if err != nil {
		return input, fmt.Errorf("error merging prompt with template: %v", err)
	}

	model := openai.GPT40314
	if llmc.model != "" {
		model = llmc.model
	}

	resp, err := llmc.openAPIClient.CreateChatCompletion(context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: input},
			},
			MaxTokens:   512,
			Temperature: 0.5,
			Stop:        []string{"\n\n"},
			User:        "wfl",
			N:           1,
		},
	)
	if err != nil {
		return "", fmt.Errorf("error creating completion: %v", err)
	}
	return resp.Choices[0].Message.Content, nil
}

type RunPBehavior string

const (
	RunPBehaviorLinuxShellScript RunPBehavior = "linuxshellscript"
	RunPBehaviorMacOSShellScript RunPBehavior = "darwinshellscript"
)

func mergeOutputAndTaskWithTemplate(output, task, promptTemplate string) (string, error) {
	type Input struct {
		Output string
		Task   string
	}
	tmpl, err := template.New("outputTransform").Parse(promptTemplate)
	if err != nil {
		panic(err)
	}
	buff := bytes.Buffer{}
	err = tmpl.Execute(&buff, Input{
		Output: output,
		Task:   task})
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}

func mergePromptWithTemplate(prompt, promptTemplate string) (string, error) {
	type Input struct {
		Prompt string
	}
	tmpl, err := template.New("mergeTemplate").Parse(promptTemplate)
	if err != nil {
		panic(err)
	}
	buff := bytes.Buffer{}
	err = tmpl.Execute(&buff, Input{Prompt: prompt})
	if err != nil {
		return "", err
	}
	return buff.String(), nil
}
