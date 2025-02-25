FROM golang:1.21.5-alpine3.18 as builder
WORKDIR /md2htmlapi
RUN apk update && apk upgrade --available && sync && apk add --no-cache --virtual .build-deps
COPY . .
RUN go build -ldflags="-w -s" .
FROM alpine:3.19.0
RUN apk update && apk upgrade --available && sync
COPY --from=builder /md2htmlapi/md2htmlapi /md2htmlapi
ENTRYPOINT ["/md2htmlapi"]
