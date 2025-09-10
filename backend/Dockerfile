FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM golang:1.24-alpine
RUN apk --no-cache add ca-certificates gettext curl docker-cli && \
    apk --no-cache update && \
    apk --no-cache upgrade ca-certificates && \
    update-ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE ${PROD_BACKEND_PORT:-6080}
CMD ["./main"]
