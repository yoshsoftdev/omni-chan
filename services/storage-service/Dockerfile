FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY shared/ shared/
COPY services/storage-service/ services/storage-service/
WORKDIR /src/services/storage-service
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /storage-service ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /storage-service /app/storage-service
COPY services/storage-service/migrations /app/migrations
ENV STORAGE_LOCAL_PATH=/data/storage
RUN mkdir -p /data/storage
EXPOSE 8082
CMD ["/app/storage-service"]
