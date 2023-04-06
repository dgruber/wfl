# Parallel Hyperparameter Tuning for Neural Networks

This repository contains an example of how to perform parallel hyperparameter tuning for a neural network using Docker containers and the wfl library in Go. The example trains a simple feed-forward neural network on the classic MNIST handwritten digit recognition task and explores different combinations of hyperparameters to find the best-performing model.

## Prerequisites

- Docker
- Go (>= 1.20)

## Setup

1. Clone the repository.

```bash
git clone https://github.com/dgruber/wfl.git
cd wfl/examples/hyperparametertuning
```

2. Build the custom Docker container image for training the MNIST model.

```bash
make build
```

This command creates a Docker container image named `parallel-training` that includes the necessary dependencies for training the neural network and the `train_mnist.py` script.

## Run

Execute the parallel hyperparameter tuning workflow using:

```bash
make run
```

This command starts the wfl workflow, which runs multiple training jobs concurrently with different combinations of hyperparameters. The workflow leverages the custom `parallel-training` Docker container image to ensure reproducibility and isolation between training jobs.

Once the workflow completes, it will print the best hyperparameters and the corresponding accuracy achieved for the MNIST model.
