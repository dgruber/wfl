package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dgruber/wfl"
)

func main() {
	RunPExampleFlow()
}

func RunPExampleFlow() {
	flow := wfl.NewWorkflow(wfl.NewProcessContext()).WithLLMOpenAI(
		wfl.OpenAIConfig{
			Token:      os.Getenv("OPENAI_KEY"),
			RunPMethod: wfl.RunPBehaviorMacOSShellScript,
		}).OnErrorPanic()

	fmt.Println("WFL Bash Shell")

	for {
		fmt.Printf("Do: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Internal error: %s\n", err)
			continue
		}

		// create job template
		template, err := flow.TemplateP(input)
		if err != nil {
			fmt.Printf("Internal error: %s\n", err)
			continue
		}
		fmt.Printf("Applying: %s\n", template.Args[1])
		fmt.Printf("y/n: ")
		var applyTemplate string
		fmt.Scanln(&applyTemplate)
		if applyTemplate != "y" {
			fmt.Println("Do NOT apply command.")
			continue
		}

		job := flow.NewJob().RunT(template)
		if job.Errored() {
			fmt.Printf("Error: %s\n", job.ErrorP("Explain the job error and provide a solution"))
		} else {
			fmt.Println(job.Output())
		}
	}
}
