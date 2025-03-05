FROM --platform=linux/amd64 golang:1.22-alpine as builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY download/go.mod download/go.sum ./download/
COPY scan/go.mod scan/go.sum ./scan/
COPY db/go.mod db/go.sum ./db/
COPY exporter/go.mod exporter/go.sum ./exporter/

RUN go mod download
COPY . .

RUN go build -o progdl ./download/cmd/download.go

FROM --platform=linux/amd64 mcr.microsoft.com/playwright:v1.49.1-jammy

LABEL org.opencontainers.image.description="Utility for downloading 2000AD"

RUN wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
RUN tar -xvf go1.24.0.linux-amd64.tar.gz -C /usr/local

RUN USER=pwuser && \
    GROUP=pwuser && \
    curl -SsL https://github.com/boxboat/fixuid/releases/download/v0.6.0/fixuid-0.6.0-linux-amd64.tar.gz | tar -C /usr/local/bin -xzf - && \
    chown root:root /usr/local/bin/fixuid && \
    chmod 4755 /usr/local/bin/fixuid && \
    mkdir -p /etc/fixuid && \
    printf "user: pwuser\ngroup: pwuser\n" > /etc/fixuid/config.yml

WORKDIR /app

ENV PATH="${PATH}:/usr/local/go/bin"
ENV BROWSER_DIR /opt/cache
ENV DOWNLOAD_DIR /opt/downloads
ENV DOWNLOAD_COUNT 1

RUN mkdir -p /opt/downloads
RUN mkdir -p /opt/cache
RUN chown pwuser /opt/downloads
RUN chown pwuser /opt/cache

USER pwuser:pwuser
RUN go run github.com/playwright-community/playwright-go/cmd/playwright@v0.4101.1 install firefox

RUN mkdir -p ~/.config/progger/

VOLUME /opt/cache
VOLUME /opt/downloads

COPY --from=builder /app/progdl ./

ENTRYPOINT ["fixuid"]

CMD ./progdl \
    --browser-dir $BROWSER_DIR \
    --download-dir $DOWNLOAD_DIR \
    --download-count $DOWNLOAD_COUNT
