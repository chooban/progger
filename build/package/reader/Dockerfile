FROM golang:1.24-alpine AS builder

ARG TARGETARCH

RUN apk add --no-cache gcc musl-dev wget pkgconfig

WORKDIR /app

ARG PDFIUM_VERSION=6406
RUN PDFIUM_ARCH=$(case ${TARGETARCH} in amd64) echo "x64" ;; arm64) echo "arm64" ;; esac) && \
    wget -q https://github.com/bblanchon/pdfium-binaries/releases/download/chromium%2F${PDFIUM_VERSION}/pdfium-linux-${PDFIUM_ARCH}.tgz -O /tmp/pdfium.tgz \
    && mkdir -p /usr/local/lib/pdfium \
    && tar -C /usr/local/lib/pdfium -xf /tmp/pdfium.tgz \
    && rm /tmp/pdfium.tgz

COPY build/pdfium.pc /usr/local/lib/pkgconfig/pdfium.pc

COPY go.work go.work.sum ./
COPY download/go.mod download/go.sum ./download/
COPY scan/go.mod scan/go.sum ./scan/
COPY reader/go.mod reader/go.sum ./reader/
COPY exporter/go.mod exporter/go.sum ./exporter/

RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o reader ./reader/cmd/main.go

FROM alpine:3.21

LABEL org.opencontainers.image.description="Komga-compatible web reader for 2000 AD"

RUN apk add --no-cache ca-certificates

COPY --from=builder /usr/local/lib/pdfium/lib/libpdfium.so /usr/local/lib/
ENV LD_LIBRARY_PATH=/usr/local/lib

COPY --from=builder /app/reader /usr/local/bin/reader

ENV HOST=:8080
ENV DATABASE_PATH=/data/reader.db
ENV LOG_LEVEL=info
ENV SCAN_ON_STARTUP=false
ENV SCAN_DIRS=

EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["reader"]
