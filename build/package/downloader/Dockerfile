FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY download/go.mod download/go.sum ./download/
COPY scan/go.mod scan/go.sum ./scan/
COPY exporter/go.mod exporter/go.sum ./exporter/
COPY reader/go.mod reader/go.sum ./reader/

RUN go mod download
COPY . .

RUN go build -o progdl ./download/cmd/download.go

FROM mcr.microsoft.com/playwright:v1.51.1-jammy

LABEL org.opencontainers.image.description="Utility for downloading 2000AD"

ARG TARGETARCH

RUN GO_ARCH=$(case "${TARGETARCH}" in amd64) echo "amd64" ;; arm64) echo "arm64" ;; *) echo "${TARGETARCH}" ;; esac) && \
    wget "https://go.dev/dl/go1.24.0.linux-${GO_ARCH}.tar.gz"
RUN GO_ARCH=$(case "${TARGETARCH}" in amd64) echo "amd64" ;; arm64) echo "arm64" ;; *) echo "${TARGETARCH}" ;; esac) && \
    tar -xvf "go1.24.0.linux-${GO_ARCH}.tar.gz" -C /usr/local

RUN USER=pwuser && \
    GROUP=pwuser && \
    FIXUID_ARCH=$(case "${TARGETARCH}" in amd64) echo "amd64" ;; arm64) echo "arm64" ;; *) echo "${TARGETARCH}" ;; esac) && \
    curl -SsL "https://github.com/boxboat/fixuid/releases/download/v0.6.0/fixuid-0.6.0-linux-${FIXUID_ARCH}.tar.gz" | tar -C /usr/local/bin -xzf - && \
    chown root:root /usr/local/bin/fixuid && \
    chmod 4755 /usr/local/bin/fixuid && \
    mkdir -p /etc/fixuid && \
    printf "user: pwuser\ngroup: pwuser\n" > /etc/fixuid/config.yml

WORKDIR /app

ENV PATH="${PATH}:/usr/local/go/bin"
ENV BROWSER_DIR=/opt/cache
ENV DOWNLOAD_DIR=/opt/downloads
ENV DOWNLOAD_COUNT=1
ENV DOWNLOAD_PAGE=0

RUN mkdir -p /opt/downloads
RUN mkdir -p /opt/cache
RUN chown pwuser /opt/downloads
RUN chown pwuser /opt/cache

USER pwuser:pwuser
RUN go run github.com/playwright-community/playwright-go/cmd/playwright@v0.5001.0 install firefox

RUN mkdir -p ~/.config/progger/

VOLUME /opt/cache
VOLUME /opt/downloads

COPY --from=builder /app/progdl ./

ENTRYPOINT ["fixuid"]

CMD ./progdl \
    --browser-dir $BROWSER_DIR \
    --download-dir $DOWNLOAD_DIR \
    --list-page $DOWNLOAD_PAGE \
    --download \
    --download-count $DOWNLOAD_COUNT
