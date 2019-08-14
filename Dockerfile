FROM golang:1.12-alpine3.10 as builder

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on
WORKDIR /go/src/github.com/axot/dnstress
COPY . .
RUN set -eux \
    && apk update \
    && apk add git make \
    && go mod download \
    && make build_bin

# runtime image
FROM alpine:3.10
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/axot/dnstress/dnstress /
ENTRYPOINT ["/dnstress"]