FROM golang:1.16 as builder

WORKDIR /go/src/github.com/alexlast/stock-notifier

COPY go.mod .
COPY go.sum .
COPY cmd/ cmd/
COPY internal/ internal/

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GO111MODULE=on GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o notifier github.com/alexlast/stock-notifier/cmd/notifier

FROM alpine:3.12

WORKDIR /opt/notifier

COPY --from=builder /go/src/github.com/alexlast/stock-notifier/notifier .

ENTRYPOINT ["./notifier"]
