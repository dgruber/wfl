package main

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"

	"github.com/dgruber/drmaa2interface"
	"github.com/dgruber/wfl"
)

type PhraseSentiment struct {
	Phrase    string
	Sentiment float64
}

func main() {
	phrases := []string{
		"Best product ever",
		"Amazing customer service",
		"High-quality and affordable",
		"User-friendly design",
		"Simply outstanding",
		"Could be better",
		"Disappointing experience",
		"Waste of money",
		"Hard to use",
		"Terrible support",
	}

	// Create a workflow context to execute jobs in parallel
	ctx := wfl.NewProcessContext().WithDefaultJobTemplate(drmaa2interface.JobTemplate{
		OutputPath: wfl.RandomFileNameInTempDir(),
		ErrorPath:  "/dev/stderr",
	}).WithSessionName("sentiment-analysis")

	flow := wfl.NewWorkflow(ctx).NewJob()

	// Analyze sentiment for each phrase in parallel

	for _, phrase := range phrases {
		flow.RunT(
			drmaa2interface.JobTemplate{
				JobEnvironment: map[string]string{"PHRASE": phrase},
				RemoteCommand:  "python3",
				Args: []string{"-c", fmt.Sprintf(`from textblob import TextBlob
phrase = "%s"
sentiment = TextBlob(phrase).sentiment.polarity
print(sentiment)`, phrase)},
			}).OnError(func(e error) { panic(e) })
	}

	// Wait for all jobs to finish
	flow.Synchronize()

	phraseSentiments := make([]PhraseSentiment, 0, len(phrases))

	getJobOutput := func(j drmaa2interface.Job, i interface{}) error {
		sentiments := i.(*[]PhraseSentiment)
		template, err := j.GetJobTemplate()
		if err != nil {
			return err
		}
		// print output path which is different for each task
		jobOutput, err := ioutil.ReadFile(template.OutputPath)
		if err != nil {
			return err
		}
		output := strings.TrimSpace(string(jobOutput))
		sentiment, err := strconv.ParseFloat(output, 64)
		if err != nil {
			return err
		}
		*sentiments = append(*sentiments, PhraseSentiment{
			Phrase:    template.JobEnvironment["PHRASE"],
			Sentiment: sentiment,
		})
		return nil
	}

	// Collect all outputs
	err := flow.ForEach(getJobOutput, &phraseSentiments)
	if err != nil {
		panic(err)
	}
	// Sort phrases by sentiment
	sort.Slice(phraseSentiments, func(i, j int) bool {
		return phraseSentiments[i].Sentiment > phraseSentiments[j].Sentiment
	})

	// Print sorted phrases with sentiment score
	fmt.Println("Marketing phrases sorted by sentiment:")
	for _, ps := range phraseSentiments {
		fmt.Printf("Phrase: %s | Sentiment: %.2f\n", ps.Phrase, ps.Sentiment)
	}
}
