FROM golang:1.15.5 as builder

WORKDIR /go/src/github.com/alexlast/stock-notifier

COPY go.mod .
COPY cmd/ cmd/
COPY internal/ internal/

RUN GO111MODULE=on GOOS=linux GOARCH=amd64 go build -a -o notifier github.com/alexlast/stock-notifier/cmd/notifier

FROM ubuntu:20.04

WORKDIR /opt/notifier

COPY --from=builder /go/src/github.com/alexlast/stock-notifier/notifier .

RUN apt-get update && \
    apt-get -y dist-upgrade && \
    apt-get install -y ca-certificates && \
    rm -rf /var/cache/apt/lists

RUN useradd -ms /bin/bash notifier

USER notifier

ENTRYPOINT ["./notifier"]
