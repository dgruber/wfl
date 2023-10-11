FROM tensorflow/tensorflow:2.7.0

RUN pip install -U keras

COPY style_transfer.py /style_transfer.py

ENTRYPOINT ["python", "/style_transfer.py"]