from opennsfw2 import make_open_nsfw_model

model = make_open_nsfw_model()

model.export('./model/nsfw_model')
print("Model exported in SavedModel format!")
