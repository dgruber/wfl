from bark import SAMPLE_RATE, generate_audio, preload_models, set_seed
from IPython.display import Audio
from scipy.io.wavfile import write as write_wav
import numpy as np

# download and load all models
preload_models()

# set seed for reproducibility
set_seed(131)

# override seed by Go
#$$$SEEDS$$$

prompts = [ "I am a dog.",  "I am a cat.", "I am a human.", ]

# override prompts by Go
#$$$PROMPTS$$$

audio_arrays = []
history_prompt = None

for idx, prompt in enumerate(prompts):
    full_generation, audio_array = generate_audio(prompt, history_prompt=history_prompt, output_full=True)
    if idx == 0:
        history_prompt = full_generation
    audio_arrays.append(audio_array)

combined_audio = np.concatenate(audio_arrays)
write_wav("/home/bark/output/#$$$OUTPUTFILENAME$$$", SAMPLE_RATE, combined_audio)

set_seed(-1)