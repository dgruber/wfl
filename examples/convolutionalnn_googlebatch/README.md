# Cifar10 Convolutional Neural Network Parallel Training on Google Cloud

This application demonstrates parallel training of a convolutional neural network (CNN) model using the Cifar10 dataset on Google Cloud. It utilizes _wfl_ and the Google Cloud Batch API for training multiple model instances concurrently and stores the output in a Google Cloud Storage bucket. It uses Spot instances for reducing the cloud costs involved.

## Prerequisites

- A Google Cloud account with billing and APIs enabled
- A Google Cloud Storage bucket
- [Google Cloud SDK](https://cloud.google.com/sdk) installed and configured on your local machine
- [Go](https://golang.org/dl/) installed on your local machine

## Setup

1. Clone this repository:

```bash
git clone https://github.com/dgruber/wfl.git
cd wfl/examples/convolutionalnn_googlebatch/cifar10-parallel-training
```

2. Set the required environment variables:

```bash
export GOOGLE_PROJECT="your-google-project-id"
export GOOGLE_BUCKET="your-google-bucket-name"
```

Replace `your-google-project-id` and `your-google-bucket-name` with your Google Cloud project ID and Google Cloud Storage bucket name, respectively.

3. Build the container image and push to Google Container Registry:

```bash
make build
make push
```

## Run

Execute the wfl application:

```bash
make run
```

The application will perform the following steps:

1. Create a Google Batch context for running training jobs using your specified Google Cloud project and bucket.
2. Run a data preparation job to split the Cifar10 dataset into multiple parts for parallel training.
3. Submit multiple parallel training jobs using different parts of the dataset.
4. Wait for all training jobs to complete.
5. Print the accuracy and runtime of each training job.

## Output

The application will print the progress, status, and results of each job in the console. The trained model files, accuracy results, and logs will be stored in the specified Google Cloud Storage bucket.

## Customization

You can customize the number of parallel training jobs, machine types, and other job parameters by modifying the `cifar.go` file and rebuilding the application.
