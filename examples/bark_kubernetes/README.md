# Bark on Kubernetes with wfl

This project demonstrates how to harness the power of Bark to generate engaging audio from text prompts on Kubernetes  (here GKE). Bark is a transformer-based text-to-audio model from Suno that can generate realistic speech, background noise, music, and even nonverbal communications like laughter and sighs.

This example requires a GKE cluster with a V100 GPU. Feel free to test it on AKS or EKS by setting "distribution" in the JobTemplate accordingly.

## Project Structure

- `bark.go`: The main Go application that uses wfl and Bark on Kubernetes.
- `Dockerfile`: Used to create a container with Bark.
- `Makefile`: Helps build, push, and run the Bark container.
- `cloudbuild.yaml`: Configuration file for Google Cloud Build.
- `bark.py`: Embedded Bark script for text-to-audio conversion.

## Getting Started

1. Build the Docker image:

   ```
   make build
   ```

2. Push the Docker image to Google Container Registry:

   ```
   export GOOGLE_PROJECT=your-google-project-id
   make push
   ```

3. Run the Go application:

   Set the other environment variables (see bark.go).

   ```
   go run bark.go
   ```

4. Check the generated audio file at [https://raw.githubusercontent.com/dgruber/wfl/master/examples/bark_kubernetes/cow.wav](https://raw.githubusercontent.com/dgruber/wfl/master/examples/bark_kubernetes/cow.wav).

## Resources

- [wfl GitHub Repository](https://github.com/dgruber/wfl)
- [Bark GitHub Repository](https://github.com/suno-ai/bark)

## Support and Troubleshooting

For any issues or questions related to wfl, please refer to the [wfl GitHub repository](https://github.com/dgruber/wfl) for support and troubleshooting.