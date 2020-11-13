FROM golang:1.15.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-s -w" -o /go/bin/ablab

FROM alpine

LABEL org.label-schema.description="ablab.io rest-server" \
    org.label-schema.name="go-rest-server" \
    org.label-schema.url="https://github.com/acim/go-rest-server/blob/master/README.md" \
    org.label-schema.vendor="ablab.io"

RUN adduser -D ablab

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/bin/ablab /usr/bin/ablab

EXPOSE 3000 3001

USER ablab

ENTRYPOINT ["/usr/bin/ablab"]
