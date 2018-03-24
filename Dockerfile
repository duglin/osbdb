FROM golang as builder
WORKDIR /tmp
COPY broker.go .
RUN go get -d .
RUN GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 go build \
	-ldflags "-w -extldflags -static" \
	-tags netgo -installsuffix netgo \
	-o broker broker.go

FROM scratch
COPY --from=builder /tmp/broker /broker
ENTRYPOINT [ "/broker" ]
