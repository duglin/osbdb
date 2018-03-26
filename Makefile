all: broker test

IMAGE_NAME?=duglin/osbdb

broker: broker.go
	GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 go build \
		-ldflags "-w -extldflags -static" \
		-tags netgo -installsuffix netgo \
		-o broker broker.go

image: .image

.image: broker
	docker build -t $(IMAGE_NAME) .
	@touch .image

push: image
	docker push $(IMAGE_NAME)

test: .test

.test: broker $(shell sh -c "find . -name *_test.go")
	@# Old stuff
	@# Kill any existing broker that's running
	@# sh -c "kill -9 $$(ps -e | grep broker | awk '{print $$1}')"
	@# This assumes the tests will finish in 5 seconds
	@echo && echo "** Starting the tests..."
	go test -v broker*.go
	@touch .test

clean:
	rm -f broker
	docker rmi $(IMAGE_NAME) 2> /dev/null || true
	rm -f .image .test
