IMAGE_NAME ?= "sheixinsheisb/ssh-server"

image:
	@echo $(IMAGE_NAME)
	docker build . -t $(IMAGE_NAME)

push: image
	docker push $(IMAGE_NAME)