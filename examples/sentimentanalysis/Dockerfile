FROM golang:1.20

RUN apt-get update && apt-get install -y python3 python3-pip
RUN pip install textblob

WORKDIR /app
COPY . .

RUN go mod init app
RUN go get github.com/dgruber/drmaa2interface
RUN go get github.com/dgruber/wfl

CMD ["go", "run", "19.go"]

