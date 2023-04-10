FROM tensorflow/tensorflow:latest

COPY prepare_cifar10.py /prepare_cifar10.py
COPY train_cifar10.py /train_cifar10.py

ENTRYPOINT ["python3", "/train_cifar10.py"]
