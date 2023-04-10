import argparse
import os
import numpy as np
import tensorflow as tf
from tensorflow.keras import datasets, layers, models, utils

def load_data(input_file):
    data = np.load(input_file)
    X, y = data['X'], data['y']
    X = X.astype('float32') / 255.0
    y = utils.to_categorical(y, num_classes=10)
    return X, y

def create_cnn_model():
    model = models.Sequential()
    model.add(layers.Conv2D(32, (3, 3), activation='relu', input_shape=(32, 32, 3)))
    model.add(layers.MaxPooling2D((2, 2)))
    model.add(layers.Conv2D(64, (3, 3), activation='relu'))
    model.add(layers.MaxPooling2D((2, 2)))
    model.add(layers.Conv2D(64, (3, 3), activation='relu'))
    model.add(layers.Flatten())
    model.add(layers.Dense(64, activation='relu'))
    model.add(layers.Dense(10, activation='softmax'))

    model.compile(optimizer='adam',
                  loss='categorical_crossentropy',
                  metrics=['accuracy'])
    return model

def main(args):
    X_train, y_train = load_data(args.input_file)
    X_val, y_val = load_data(args.validation_data)

    model = create_cnn_model()
    model.fit(X_train, y_train, epochs=10, batch_size=64)

    val_loss, val_acc = model.evaluate(X_val, y_val)
    print(f'Validation accuracy: {val_acc}')

    model.save(args.output_file)

    # write accuracy to anouther output file
    with open(args.accuracy_file, 'w') as f:
        f.write(str(val_acc))

    
if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--input-file', required=True, help='Path to the input dataset file')
    parser.add_argument('--output-file', required=True, help='Path to the output file for the trained model')
    parser.add_argument('--validation-data', required=True, help='Path to the validation dataset file')
    parser.add_argument('--accuracy-file', required=True, help='Path to the output file for the accuracy')
    args = parser.parse_args()

    main(args)
