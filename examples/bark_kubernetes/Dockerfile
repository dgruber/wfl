FROM pytorch/pytorch:2.0.0-cuda11.7-cudnn8-devel

RUN apt-get update && apt-get install -y git curl

RUN groupadd -g 1300 barker

RUN useradd -u 1300 -g 1300 -m bark

USER bark

#RUN cd /home/bark && git clone https://github.com/suno-ai/bark.git

# I am using my own fork because I pulled a PR from the original repo.
# That supports setting seeds for the random number generator.
RUN cd /home/bark && git clone https://github.com/dgruber/bark.git

WORKDIR /home/bark

RUN cd /home/bark/bark && pip install .

