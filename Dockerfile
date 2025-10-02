FROM golang:1.25.0-alpine3.22 as build

ARG version
ARG gitVersion

RUN apk update && \
    apk add git

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify

COPY . .

RUN go build \
		-trimpath \
    -ldflags="-X main.version=${version}" \
		-o ../bin/hka-2fa-proxy \
		.

FROM alpine:3.22.1

COPY --from=build /go/bin/hka-2fa-proxy /usr/local/bin/hka-2fa-proxy

RUN addgroup -S proxy && \
    adduser -S -G proxy proxy && \
    mkdir -p /var/lib/hka-2fa-proxy && \
    chown proxy:proxy /var/lib/hka-2fa-proxy

WORKDIR /var/lib/hka-2fa-proxy
USER proxy

EXPOSE 8080

ENTRYPOINT [ "hka-2fa-proxy" ]
