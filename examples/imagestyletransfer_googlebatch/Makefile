# Replace with your Google Cloud Project ID
PROJECT_ID := your-gcp-project
# Replace with your Google Cloud Storage bucket name
BUCKET_NAME := your-gcp-bucket

DOCKER_IMAGE_NAME := gcr.io/$(PROJECT_ID)/style-transfer

# Build and push the Docker image
build-image:
	docker build -t $(DOCKER_IMAGE_NAME) .
	docker push $(DOCKER_IMAGE_NAME)

# Run the wfl style transfer example
run: build-image
	GOOGLE_PROJECT=$(PROJECT_ID) GOOGLE_BUCKET=$(BUCKET_NAME) go run style_transfer.go

.PHONY: build-image run