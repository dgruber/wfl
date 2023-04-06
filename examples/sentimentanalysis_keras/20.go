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
		OutputPath: wfl.RandomFileNameInTempDir() + "{{.ID}}.out",
		ErrorPath:  "/dev/stderr",
	}).WithSessionName("wfl-example")

	flow := wfl.NewWorkflow(ctx).NewJob()

	// Analyze sentiment for each phrase in parallel using ANN
	annCodeTrain := `import numpy as np
from keras.models import Sequential
from keras.layers import Dense
from keras.preprocessing.text import Tokenizer
from keras.preprocessing.sequence import pad_sequences
from keras.models import save_model

np.random.seed(7)
	
# Define example training data
#texts = ["This product is great", "I love this item", "Terrible experience", "Worst purchase ever"]
texts = ["This product is great", "I love this item", "Amazing quality", "Impressive performance",
"Fantastic design", "Highly recommended", "Excellent value", "Superior experience",
"Terrible experience", "Worst purchase ever", "Complete waste of money", "Very disappointing",
"Poor customer service", "Not worth the price", "Hard to use", "Can't recommend"]

#labels = np.array([1, 1, 0, 0], dtype='float32')
labels = np.array([1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0], dtype='float32')

# Tokenize and pad input phrases
tokenizer = Tokenizer()
tokenizer.fit_on_texts(texts)
sequences = tokenizer.texts_to_sequences(texts)
data = np.array(pad_sequences(sequences, maxlen=4), dtype='float32')
	
# Build a simple ANN model
model = Sequential()
#model.add(Dense(8, input_dim=4, activation='relu'))
#model.add(Dense(1, activation='sigmoid'))
#model.compile(loss='binary_crossentropy', optimizer='adam', metrics=['accuracy'])

model.add(Dense(16, input_dim=4, activation='relu'))
model.add(Dense(1, activation='sigmoid'))
model.compile(loss='binary_crossentropy', optimizer='adam', metrics=['accuracy'])

# Train the model
model.fit(data, labels, epochs=150, batch_size=16, verbose=0)

# Save the trained model
model.save('/tmp/ann_model.h5')
`

	annCodeInference := `import numpy as np
import numpy as np
from keras.preprocessing.text import Tokenizer
from keras.preprocessing.sequence import pad_sequences
from keras.models import load_model
	
np.random.seed(7)
	
loaded_model = load_model('/tmp/ann_model.h5')

texts = ["This product is great", "I love this item", "Terrible experience", "Worst purchase ever"]
tokenizer = Tokenizer()
tokenizer.fit_on_texts(texts)

# Classify the given phrase using the trained model
input_phrase = "%s"
input_sequence = tokenizer.texts_to_sequences([input_phrase])
input_data = np.array(pad_sequences(input_sequence, maxlen=4), dtype='float32')
sentiment = loaded_model.predict(input_data)[0][0]
		
print(sentiment)
`

	// Train the ANN model
	flow.RunT(
		drmaa2interface.JobTemplate{
			RemoteCommand: "python3",
			Args:          []string{"-c", annCodeTrain},
			JobEnvironment: map[string]string{
				"TRAINING": "true",
			},
		}).OnError(func(e error) { panic(e) }).Wait()

	// Run inference for each phrase in parallel
	for _, phrase := range phrases {
		flow.RunT(
			drmaa2interface.JobTemplate{
				JobEnvironment: map[string]string{"PHRASE": phrase, "INFERENCE": "true"},
				RemoteCommand:  "python3",
				Args:           []string{"-c", fmt.Sprintf(annCodeInference, phrase)},
				ErrorPath:      "/dev/stderr",
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
		// skip training job
		if template.JobEnvironment["INFERENCE"] != "true" {
			return nil
		}
		// print output path which is different for each task
		jobOutput, err := ioutil.ReadFile(template.OutputPath)
		if err != nil {
			return err
		}
		fmt.Printf("job %s output: %s\n", j.GetID(), string(jobOutput))
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
	err := flow.ForAll(getJobOutput, &phraseSentiments)
	if err != nil {
		panic(err)
	}
	// Sort phrases by sentiment
	sort.Slice(phraseSentiments, func(i, j int) bool {
		return phraseSentiments[i].Sentiment > phraseSentiments[j].Sentiment
	})

	// Print sorted phrases with sentiment score
	fmt.Println("Marketing phrases sorted by sentiment (using ANN):")
	for _, ps := range phraseSentiments {
		fmt.Printf("Phrase: %s | Sentiment: %.2f\n", ps.Phrase, ps.Sentiment)
	}
}
