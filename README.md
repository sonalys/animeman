# Animeman

[![Release](https://github.com/sonalys/animeman/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/sonalys/animeman/actions/workflows/goreleaser.yml)
[![Tests](https://github.com/sonalys/animeman/actions/workflows/tests.yml/badge.svg)](https://github.com/sonalys/animeman/actions/workflows/tests.yml)

Animeman is a service for synchronizing your anime list currently watching with Nyaa and QBittorrent.  
Currently it manages qBittorrent through it's WebUI, creating and managing a category of torrents.  
It automatically parses the torrent titles for tagging the show, season and episodes, while also searching in Nyaa.si for new releases.

## Features

* **Automatic Downloads** weekly releases from your WatchList
* **Downloads batch releases**: from complete series from your WatchList
* **Tags**: all torrent entries under the configured category with [`!Serie name`, `Serie name S01E01`] as an example
* **Source and quality filter**: you can specify resolution and HEVC tag
* **Smart episode detection**: you don't need to worry about downloading the same episode twice

## How does it work?

0. Tag existing torrents in the configured category
1. Fetch your **Currently Watching** entries from **MAL** or **Anilist**
2. Search in the RSS feed for each entry in **Nyaa.si**
3. Validate if the episode / season is already present in **qBittorrent**
4. Add torrent to qBittorrent via the WebUI API

The purpose of this tool is to download the latest RSS entry for each episode.
It prioritizes the highest provided quality, respecting your filter.
If there are multiple sources for the same quality, it should choose the one with the highest number of seeders.

## Configuration

Animeman will generate a boilerplate config for the first time.  
You can set your own config path with the env `CONFIG_PATH`.

```yaml
# config.yaml
logLevel: info # (debug,info,error).
animeList:
  type: myanimelist # (myanimelist|anilist).
  username: YOUR_USERNAME # Replace with your username.
rssConfig:
  type: nyaa
  pollFrequency: 5m0s # min 1m0s.
  sources:
      - source1 # replace with your sources or remove the sources field to fetch all.
      - source2
  qualities:
      - 1080 # filter for 1080, 720, HEVC or remove to fetch all.
  customParameters:
    c: 1_2 # you can configure custom query parameters for the rss list call. In this example it will set ?c=1_2.
torrentConfig:
  type: qbittorrent
  category: Animes
  downloadPath: /downloads/animes
  createShowFolder: true # creates a folder to for the show inside downloadPath.
  renameTorrent: true # will rename the torrent in qBittorrent avoiding conflict between multiple sources with different names for the show.
  host: http://192.168.1.240:8088 # replace with your qBittorrent WebUI address.
  username: admin # replace credentials with your own
  password: adminadmin
```

## Installation

### Download

You can download the latest release [here](https://github.com/sonalys/animeman/releases).  
You can run a first time for generating a boilerplate config, then you configure your `config.yaml`.

### Linux CLI

Simply run `CONFIG_PATH=./config.yaml ./animeman`

### Windows

Simply run `animeman.exe` on the `cmd`.

### Docker

Support for `linux/amd64` and `linux/arm64`.

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

## Building

### Dependencies

You will need at least go 1.22 for building the binary.  
For the image you will need docker.  
To build you can simply run `make build`  
For the image you can run `make image`

## Roadmap

There are a couple things that will be iterated:

* Use some calendar service like anilist.co for scanning Nyaa only when close to the release date
* Improve interfaces for allowing other RSS feeds

## Contribution

Feel free to fork and open pull requests  
Tests or roadmap features are very welcome, thanks.

## Disclaimer

This tool is intended as a proof-of-concept, and is not intended for any illegal activities.
