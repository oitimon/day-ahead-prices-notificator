FROM golang:1.23-alpine AS builder

WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o da-price-notificator main.go

FROM alpine:latest

WORKDIR /app

# Install tzdata
RUN apk add --no-cache tzdata

# Create a non-root user
RUN addgroup -S nonroot && \
    adduser -S -G nonroot nonroot
USER nonroot

COPY --from=builder  /app/da-price-notificator .

# Expose necessary ports
EXPOSE 8080

CMD ["./da-price-notificator"]
