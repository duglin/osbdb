all: image

IMAGE_NAME?=duglin/osdb

broker: broker.go
	GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 go build \
		-ldflags "-w -extldflags -static" \
		-tags netgo -installsuffix netgo \
		-o broker broker.go

image: .image

.image: broker
	docker build -t $(IMAGE_NAME) .
	touch .image

push: image
	docker push $(IMAGE_NAME)

clean:
	rm -f broker
	docker rmi osdb 2> /dev/null || true
	rm -f .image
