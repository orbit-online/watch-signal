FROM golang
COPY . /go/src/github.com/orbit-online/watch-signal
WORKDIR /go/src/github.com/orbit-online/watch-signal
RUN go get ./...
RUN go build -ldflags="-s -linkmode external -extldflags -static -w"

FROM scratch
COPY --from=0 /go/src/github.com/orbit-online/watch-signal/watch-signal /
ENTRYPOINT ["/watch-signal"]
