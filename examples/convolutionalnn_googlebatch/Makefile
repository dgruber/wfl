.PHONY: build push run

export GOOGLE_PROJECT=customer-nest
# without gs:// prefix
export GOOGLE_BUCKET=uc_experimental

build:
	docker build -t gcr.io/${GOOGLE_PROJECT}/cifar10-parallel-training .

push:
	docker push gcr.io/$(GOOGLE_PROJECT)/cifar10-parallel-training

run:
	go run cifar.go
