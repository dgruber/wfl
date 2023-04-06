# ANN Sentiment Analysis Container

This container runs an Artificial Neural Network (ANN) based sentiment analysis using Keras and TensorFlow. The ANN is trained on a dataset of phrases and their corresponding sentiment labels (positive or negative). Once the model is trained, it predicts the sentiment of new phrases and prints the results.

## Requirements

- Docker

## Building the Container

1. Make sure the `Dockerfile` and the `20.go` file are in the same directory.
2. Open a terminal and navigate to the directory containing the `Dockerfile`.
3. Run the following command to build the Docker container:

```bash
docker build -t ann-sentiment-analysis .
```

This command builds a Docker container with the tag `ann-sentiment-analysis`.

## Running the Container

After building the container, you can run it using the following command:

```bash
docker run --rm ann-sentiment-analysis
```

This command runs the container and automatically removes it after execution. The container will train the ANN model on the provided dataset, predict the sentiment of new phrases, and print the results.

## Customizing the Container

You can add more training examples and modify the model parameters by editing the `20.go` file. After making any changes, rebuild the container using the `docker build` command as described above.