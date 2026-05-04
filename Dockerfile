FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY static/ static/
COPY VERSION .

RUN go build -o da-price-notificator cmd/server/main.go

FROM alpine:3.22

WORKDIR /app

# Install tzdata
RUN apk add --no-cache tzdata=2026b-r0

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
