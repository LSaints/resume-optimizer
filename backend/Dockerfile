FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o server ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates curl

RUN ARCH=$(uname -m) && \
    case "$ARCH" in \
        x86_64)  TYPST_ARCH="x86_64" ;; \
        aarch64) TYPST_ARCH="aarch64" ;; \
        armv7l)  TYPST_ARCH="arm" ;; \
        *) echo "unsupported arch: $ARCH" && exit 1 ;; \
    esac && \
    curl -fsL "https://github.com/typst/typst/releases/latest/download/typst-${TYPST_ARCH}-unknown-linux-musl.tar.xz" \
      -o /tmp/typst.tar.xz && \
    tar -xJf /tmp/typst.tar.xz -C /tmp && \
    mv /tmp/typst-${TYPST_ARCH}-unknown-linux-musl/typst /usr/local/bin/typst && \
    rm -rf /tmp/typst*

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
