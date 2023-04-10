.PHONY: build push run

build:
	docker build -t parallel-training .

push:
	docker tag parallel-training gcr.io/$(GOOGLE_PROJECT)/parallel-training
	docker push gcr.io/$(GOOGLE_PROJECT)/parallel-training

run:
	go run 21.go
