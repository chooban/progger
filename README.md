# Progger


## Intro

A collection of libraries and utilities related to reading 2000AD.

## Projects 


### Downloader

A Docker container to allow for automated downloads of Progs. 

Image available [on docker hub](https://hub.docker.com/repository/docker/chooban/progger-downloader/general) for AMD64 only
due to Playwright dependency. The image requires your Rebellion username and password so you
should probably treat it with great suspicion. These details are provided via env vars because I run it in a
single tenant environment and I'm not too concerned about leaking details to myself.

Example docker-compose:

```yaml
services:
  progger:
    image: "chooban/progger-downloader:latest"
    environment:
      - REBELLION_USERNAME="mylogin@test.com"
      - REBELLION_PASSWORD="notmypassword"
    volumes:
      - ./cache:/opt/cache
      - ./downloads:/opt/downloads
```


### Exporter

If I want to reread a series, or do a catchup, I want to read one series at a time. The [exporter](./exporter/) lets me do that.


## Motivation

Learn more Go. Read more comics.
