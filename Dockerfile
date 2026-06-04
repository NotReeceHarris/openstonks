# ---- build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY app/go.mod app/go.sum ./
RUN go mod download

COPY app/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /openstonks ./cmd/openstonks

# ---- final stage ----
FROM alpine:3.19

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /openstonks /usr/local/bin/openstonks

USER app

ENTRYPOINT ["/usr/local/bin/openstonks"]
