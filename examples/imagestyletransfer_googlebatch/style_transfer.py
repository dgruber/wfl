import argparse
import numpy as np
import tensorflow as tf
from tensorflow.keras.applications import vgg19

def preprocess_image(image_path):
    img = tf.keras.preprocessing.image.load_img(image_path, target_size=(512, 512))
    img = tf.keras.preprocessing.image.img_to_array(img)
    img = np.expand_dims(img, axis=0)
    img = vgg19.preprocess_input(img)
    return img

def deprocess_image(x):
    x[:, :, 0] += 103.939
    x[:, :, 1] += 116.779
    x[:, :, 2] += 123.68
    x = x[:, :, ::-1]
    x = np.clip(x, 0, 255).astype('uint8')
    return x

def gram_matrix(x):
    x = tf.transpose(x, (2, 0, 1))
    features = tf.reshape(x, (tf.shape(x)[0], -1))
    gram = tf.matmul(features, tf.transpose(features))
    return gram

def style_loss(style, combination):
    S = gram_matrix(style)
    C = gram_matrix(combination)
    channels = 3
    size = 512 * 512
    return tf.reduce_sum(tf.square(S - C)) / (4.0 * (channels ** 2) * (size ** 2))

def content_loss(base, combination):
    return tf.reduce_sum(tf.square(combination - base))

def total_variation_loss(x):
    a = tf.square(x[:, :511, :511, :] - x[:, 1:, :511, :])
    b = tf.square(x[:, :511, :511, :] - x[:, :511, 1:, :])
    return tf.reduce_sum(tf.pow(a + b, 1.25))

def main(args):
    content_image = preprocess_image(args.content_image)
    style_image = preprocess_image(args.style_image)

    content_layer = 'block5_conv2'
    style_layers = [
        'block1_conv1',
        'block2_conv1',
        'block3_conv1',
        'block4_conv1',
        'block5_conv1',
    ]

    content_model = vgg19.VGG19(weights='imagenet', include_top=False)
    content_output = content_model.get_layer(content_layer).output
    content_model = tf.keras.Model(content_model.inputs, content_output)

    style_model = vgg19.VGG19(weights='imagenet', include_top=False)
    style_outputs = [style_model.get_layer(layer).output for layer in style_layers]
    style_model = tf.keras.Model(style_model.inputs, style_outputs)

    content_target = tf.constant(content_model(content_image))
    style_targets = [gram_matrix(output) for output in style_model(style_image)]

    combination_image = tf.Variable(content_image)
    combination_output = content_model(combination_image)
    content_loss_value = content_loss(content_target, combination_output)

    combination_outputs = style_model(combination_image)
    style_loss_values = [style_loss(style_target, combination_output) for style_target, combination_output in zip(style_targets, combination_outputs)]
    style_loss_value = tf.add_n(style_loss_values) * args.style_weight / len(style_layers)

    total_variation_loss_value = total_variation_loss(combination_image)

    loss = content_loss_value + style_loss_value + total_variation_loss_value

    opt = tf.optimizers.Adam(learning_rate=5.0)
    epochs = 100

    @tf.function
    def train_step(image):
        with tf.GradientTape() as tape:
            tape.watch(image)
            loss_value = loss
        grads = tape.gradient(loss_value, image)
        opt.apply_gradients([(grads, image)])

    for epoch in range(epochs):
        train_step(combination_image)
        print(f'Epoch {epoch + 1}/{epochs}')

    output_image = deprocess_image(combination_image.numpy()[0])
    tf.keras.preprocessing.image.save_img(args.output_image, output_image)

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--content-image', required=True, help='Path to the content image')
    parser.add_argument('--style-image', required=True, help='Path to the style image')
    parser.add_argument('--output-image', required=True, help='Path to the output image')
    parser.add_argument('--style-weight', type=float, default=1e-6, help='Weight for the style loss')
    args = parser.parse_args()

    main(args)