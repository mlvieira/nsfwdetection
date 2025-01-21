import tensorflow as tf
from keras.layers import TFSMLayer
import numpy as np
from PIL import Image
import os

base_dir = os.path.dirname(os.path.dirname(__file__))
model_path = os.path.join(base_dir, "model", "nsfw_model")
image_dir = os.path.join(base_dir, "model_test")

model = TFSMLayer(model_path, call_endpoint='serving_default')

def preprocess_image(image_path):
    try:
        image = Image.open(image_path).convert('RGB')

        image = image.resize((256, 256), Image.LANCZOS)

        left = (256 - 224) / 2
        top = (256 - 224) / 2
        right = left + 224
        bottom = top + 224
        image = image.crop((left, top, right, bottom))

        image = np.array(image).astype(np.float32)

        image = image[:, :, ::-1]

        mean = [104.0, 117.0, 123.0]
        image -= mean

        image = np.expand_dims(image, axis=0)

        return image
    except Exception as e:
        print(f"Error preprocessing {image_path}: {str(e)}")
        return None


def process_image(image_path):
    try:
        image = preprocess_image(image_path)
        if image is None:
            return None

        output = model(image)
        output_tensor = output['output_0']
        output_array = output_tensor.numpy()

        nsfw_score = output_array[0][1]
        return nsfw_score

    except Exception as e:
        print(f"Error processing {image_path}: {str(e)}")
        return None


image_paths = [os.path.join(image_dir, fname) for fname in os.listdir(image_dir) if fname.endswith(('.jpg', '.png'))]

for img_path in image_paths:
    nsfw_score = process_image(img_path)
    if nsfw_score is not None:
        print(f"{os.path.basename(img_path)} - NSFW Score: {nsfw_score:.4f}")
