FROM golang:alpine AS builder
WORKDIR /go/src/spider-go
COPY . .
RUN go install -v ./...

FROM alpine
WORKDIR /spider-go
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /go/bin/spider-go .
COPY ./config.yml ./
ENTRYPOINT ["./spider-go"]