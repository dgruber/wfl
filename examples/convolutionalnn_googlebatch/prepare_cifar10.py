import argparse
import os
import numpy as np
from tensorflow.keras import datasets

def check_existing_parts(output_dir, num_parts):
    existing_parts = []
    for filename in os.listdir(output_dir):
        if filename.startswith("cifar10_part") and filename.endswith(".npz"):
            existing_parts.append(filename)

    return len(existing_parts) == num_parts

def delete_existing_parts(output_dir):
    for filename in os.listdir(output_dir):
        if filename.startswith("cifar10_part") and filename.endswith(".npz"):
            os.remove(os.path.join(output_dir, filename))

def split_cifar10(num_parts, output_dir):
    if not check_existing_parts(output_dir, num_parts):
        delete_existing_parts(output_dir)

    (X_train, y_train), (X_val, y_val) = datasets.cifar10.load_data()

    # Shuffle the training data
    indices = np.arange(X_train.shape[0])
    np.random.shuffle(indices)
    X_train = X_train[indices]
    y_train = y_train[indices]

    # Split into N parts
    num_samples = X_train.shape[0] // num_parts
    for i in range(num_parts):
        start = i * num_samples
        end = (i + 1) * num_samples
        np.savez(os.path.join(output_dir, f'cifar10_part{i + 1}.npz'), X=X_train[start:end], y=y_train[start:end])

    # Save the validation dataset
    np.savez(os.path.join(output_dir, 'cifar10_validation.npz'), X=X_val, y=y_val)

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--num-parts', type=int, default=4, help='Number of parts to split the dataset into')
    parser.add_argument('--output-dir', type=str, required=True, help='Directory to save the dataset parts')
    args = parser.parse_args()

    split_cifar10(args.num_parts, args.output_dir)