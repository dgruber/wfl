# Parallel Hyperparameter Tuning for Neural Networks on Google Batch

This repository contains an example of how to perform parallel hyperparameter tuning for a neural network using Google Batch and the wfl library in Go. The example trains a simple feed-forward neural network on the classic MNIST handwritten digit recognition task and explores different combinations of hyperparameters to find the best-performing model.

This example is similar to hyperpamatertuning but instead of running locally it runs on Google cloud. For that you need to be logged into Google cloud when starting the application (gcloud auth login / gcloud auth application-default login). For building the container image and running the application please set the _GOOGLE_PROJECT_ (to your Google cloud project ID) and _GOOGLE_BUCKET_ (to your bucket name with "gs://" prefix, like "gs://experiment") environment variable.

## Prerequisites

- Docker
- Go (>= 1.20)

## Setup

1. Clone the repository.

```bash
git clone https://github.com/dgruber/wfl.git
cd wfl/examples/hyperparametertuning
```

2. Build the custom Docker container image for training the MNIST model and push to Google Container Registry (GCR).

```bash
export GOOGLE_PROJECT=projectid
export GOOGLE_BUCKET=gs://experiment"

# log into Google cloud
gcloud auth application-default login

make build
make push
```

This command creates a Docker container image named `parallel-training` that includes the necessary dependencies for training the neural network and the `train_mnist.py` script.

## Run

Execute the parallel hyperparameter tuning workflow using:

```bash
make run
```

This command starts the wfl workflow, which runs multiple (27) training jobs concurrently with different combinations of hyperparameters. The workflow leverages the custom `parallel-training` Docker container image to ensure reproducibility and isolation between training jobs.

Once the workflow completes, it will print the best hyperparameters and the corresponding accuracy achieved for the MNIST model.
