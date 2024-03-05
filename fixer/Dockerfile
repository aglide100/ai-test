FROM golang:1.20-alpine3.17 AS builder

RUN apk update && apk add upx && apk --no-cache add ca-certificates

WORKDIR "/go/src/github.com/aglide100/ai-test"

COPY go.mod .
COPY go.sum .

RUN mkdir -p /opt/bin/fixer/

RUN go mod tidy \
    && go get -u -d -v ./...

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w' -o /opt/bin/fixer/main cmd/fixer/main.go \    
    && upx --lzma /opt/bin/fixer/main

FROM debian:stretch-slim AS runtime
COPY --from=builder /opt/bin/fixer /opt/bin/fixer
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD [ "/opt/bin/fixer/main" ]