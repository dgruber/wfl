.PHONY: build run push

build:
	docker build -t bark .

cloudbuild:
	gcloud builds submit --config cloudbuild.yaml .

push:
	docker tag bark gcr.io/$(GOOGLE_PROJECT)/bark
	docker push gcr.io/$(GOOGLE_PROJECT)/bark

run:
	go run bark.go
