# Sentiment Analyzer

This sentiment analyzer app analyzes the sentiment of marketing phrases and sorts them based on their sentiment scores. The app runs in a Docker container using a Makefile for building the image and running the container.

## Prerequisites

- Docker

## Building the Docker Image

To build the Docker image for the sentiment analyzer, first make sure you have the Docker daemon running on your system. Then, use the provided Makefile by running the following command:

```sh
make build
```

This command builds a Docker image named `sentiment-analyzer`.

## Running the Sentiment Analyzer

To run the sentiment analyzer in a Docker container, use the following command:

```sh
make run
```

This command runs a Docker container using the `sentiment-analyzer` image and removes the container after execution. The sentiment analyzer app will analyze the sentiment of the marketing phrases and sort them based on their sentiment scores. The sorted phrases along with their sentiment scores will be printed to the console.


