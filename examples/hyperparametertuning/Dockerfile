# Use the official Python 3.8 image as the base image
FROM python:3.8

# Install any additional dependencies, e.g., TensorFlow or PyTorch
RUN pip install tensorflow

# Copy the train_mnist.py script to the container
COPY train_mnist.py /train_mnist.py

# Set the default command to execute the train_mnist.py script
CMD ["python", "/train_mnist.py"]
