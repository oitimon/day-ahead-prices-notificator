FROM golang:1.23-alpine AS builder

WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o da-price-notificator cmd/server/main.go

FROM alpine:3.21.3

WORKDIR /app

# Install tzdata
RUN apk add --no-cache tzdata=2025b-r0

# Create a non-root user
RUN addgroup -S nonroot && \
    adduser -S -G nonroot nonroot
USER nonroot

COPY --from=builder  /app/da-price-notificator .
COPY --from=builder  /app/VERSION .
COPY --from=builder  /app/static/ static/

# Expose necessary ports
EXPOSE 8080
EXPOSE 9090

CMD ["./da-price-notificator"]
