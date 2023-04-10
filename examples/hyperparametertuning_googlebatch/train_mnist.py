import argparse
import tensorflow as tf

def train_mnist(learning_rate, batch_size, num_epochs):
    # Load the MNIST dataset
    (x_train, y_train), (x_test, y_test) = tf.keras.datasets.mnist.load_data()

    # Preprocess the data
    x_train, x_test = x_train / 255.0, x_test / 255.0

    # Create a simple feed-forward neural network model
    model = tf.keras.models.Sequential([
        tf.keras.layers.Flatten(input_shape=(28, 28)),
        tf.keras.layers.Dense(128, activation='relu'),
        tf.keras.layers.Dropout(0.2),
        tf.keras.layers.Dense(10, activation='softmax')
    ])

    # Compile the model
    model.compile(optimizer=tf.keras.optimizers.Adam(learning_rate=learning_rate),
                  loss='sparse_categorical_crossentropy',
                  metrics=['accuracy'])

    # Train the model
    model.fit(x_train, y_train, batch_size=batch_size, epochs=num_epochs)

    # Evaluate the model
    _, accuracy = model.evaluate(x_test, y_test)
    return accuracy

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--learning-rate", type=float, default=0.01)
    parser.add_argument("--batch-size", type=int, default=32)
    parser.add_argument("--num-epochs", type=int, default=10)
    parser.add_argument("--output-file", type=str, default="output.txt")
    args = parser.parse_args()

    accuracy = train_mnist(args.learning_rate, args.batch_size, args.num_epochs)
    print(accuracy)
    with open(args.output_file, 'w') as f:
        f.write(str(accuracy))