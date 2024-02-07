# Animeman

[![Build](https://github.com/sonalys/animeman/actions/workflows/build.yml/badge.svg)](https://github.com/sonalys/animeman/actions/workflows/build.yml)
[![Tests](https://github.com/sonalys/animeman/actions/workflows/tests.yml/badge.svg)](https://github.com/sonalys/animeman/actions/workflows/tests.yml)

Animeman is a service for fetching your MyAnimeList currently watching animes from Nyaa.si RSS feed.

Currently it manages qBitTorrent through it's WebUI, creating and managing a category of torrents.

It automatically parses the torrent titles for tagging the show, season and episodes, while also searching in Nyaa.si for new releases.

## How does it work?

We fetch your currently watching anime list from MAL and search Nyaa.si with some extra parameters for entries.

You can specify several sources and quality.

It will fetch from the first result found, for it to download as soon as possible.

Currently it only fetches the latest entry, so if you missed an episode, you will have to download it yourself.

We avoid duplication from other sources by tagging downloads with season and episode tags, and checking if it already exists.

## Configuration

You can run animeman for it to generate a boilerplate config file, then you configure it yourself.

### Config.yaml

It will read config.yaml either from the current work directory or from the env `CONFIG_PATH`.

```yaml
sources:
    - source1
    - source2
qualities:
    - "1080 HEVC" # Only downloads HEVC that are 1080p
    - "720"
category: Animes
downloadPath: /downloads/animes
createShowFolder: true
malUser: YOUR_USER
qBitTorrentHost: http://192.168.1.240:8088 # qBittorrent WebUI Host.
qBitTorrentUsername: admin # change with qBitTorrent credentials.
qBitTorrentPassword: adminadmin
pollFrequency: 15m0s # How often should we seek for updates?
```

## Building

### Dependencies

You will need at least go 1.16 for building the binary.

For the image you will need docker.

To build you can simply run `make build`

For the image you can run `make image`

## Running

You can run a first time for generating a boilerplate config, then you configure your `config.yaml`.

### CLI

Simply run `CONFIG_PATH=./config.yaml ./animeman`

### Docker

```docker run -it -e CONFIG_PATH=/config/config.yaml -v ./config:/config ghcr.io/sonalys/animeman:latest```

### Docker Compose

```yaml
# docker-compose.yaml
version: "2.1"
services:
  animeman:
    image: ghcr.io/sonalys/animeman:latest
    container_name: animeman
    environment:
      - CONFIG_PATH=/config/config.yaml
    volumes:
      - ./config:/config
```

`docker compose -f docker-compose.yaml up -d animeman`

## Roadmap

There are a couple things that will be iterated:

* Improve batch torrent validation ( example: batches containing only part of the episodes of a season are blocking the download of the subsequent episodes )
* Use some calendar service like anilist.co for scanning Nyaa only when close to the release date
* Improve behavior for adding shows that are already released
* Improve interfaces for allowing other torrent clients
* Improve interfaces for allowing other RSS feeds and anime lists

## Contribution

Feel free to fork and open pull requests

Tests or roadmap features are very welcome, thanks.