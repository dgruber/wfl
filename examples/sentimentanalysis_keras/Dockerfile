FROM golang:1.20

RUN apt-get update && \
    apt-get install -y python3 python3-pip && \
    pip3 install protobuf==3.20.1 keras==2.6.0 numpy tensorflow==2.6.0

WORKDIR /app
COPY . .

RUN go mod init app
RUN go get github.com/dgruber/drmaa2interface
RUN go get github.com/dgruber/wfl@v1.3.0

CMD ["go", "run", "20.go"]
