# Animeman

[![Release](https://github.com/sonalys/animeman/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/sonalys/animeman/actions/workflows/goreleaser.yml)
[![Tests](https://github.com/sonalys/animeman/actions/workflows/tests.yml/badge.svg)](https://github.com/sonalys/animeman/actions/workflows/tests.yml)

Animeman is a service for synchronizing your anime list currently watching with a torrent source and QBittorrent.  
Currently it manages qBittorrent through it's WebUI, creating and managing a category of torrents.  
It automatically parses the torrent titles for tagging the show, season and episodes, while also searching for new releases.  
Torrent sources are pluggable; Animeman currently supports **Nyaa.si** and **NekoBT**, configured via `torrentSource.type`.

## Features

* **Automatic Downloads** weekly releases from your WatchList
* **Downloads batch releases**: from complete series from your WatchList
* **Tags**: all torrent entries under the configured category with [`!Serie name`, `S01E01`] as an example
* **Source and quality filter**: you can specify prioritized matching based on source and quality filters
* **Smart episode detection**: you don't need to worry about downloading the same episode twice
* **Custom torrent renaming logic**: you are able to write exactly how your torrents should be named using [expr-lang](https://expr-lang.org/docs/language-definition)

## How does it work?

0. Tag existing torrents in the configured category in **qBittorrent**
1. Fetch your **Currently Watching** entries from **MAL** or **Anilist**
2. Search the configured **torrent source** (Nyaa.si or NekoBT) for episodes for each anime list entry
3. Scan through results searching for newer episodes than the existing ones in **qBittorrent**  
  It doesn't search for specific episodes, it uses all results from a single page to retrieve new episodes.  
  This might not work well for shows with more episodes than the source's page size.  
  Changing this approach is not viable at the moment, 
  as it would require Animeman more requests about each episode information.
4. Add torrent to qBittorrent via the WebUI API

The purpose of this tool is to download the latest RSS entry for each episode.
It prioritizes the highest provided quality, respecting your filter.
If there are multiple sources for the same quality, it should choose the one with the highest number of seeders.

## Configuration

Animeman will generate a boilerplate config for the first time.  
You can set your own config path with the env `CONFIG_PATH`.

```yaml
# config.yaml
logLevel: info # Accetps: debug,info,error.
animeList:
  type: myanimelist # Accepts: myanimelist,anilist.
  username: YOUR_USERNAME
torrentSource:
  type: nyaa # Accepts: nyaa,nekobt. Must also define the corresponding config.
  # https://nyaa.si/help
  nyaa:
    customParameters: # Configures custom query parameters for the nyaa list call.
      c: 1_2 # Defines english only anime sources.
  # https://wiki.nekobt.to/technical-details/torznab
  # nekobt:
  #   apiKey: YOUR_API_KEY # Required when type is nekobt.
  #   customParameters: # Configures extra torznab query parameters, they override the defaults.
  #     audio_lang: ja # Japanese audio only, filters out dubs.
  #     sub_lang: en # English subtitles only.
  #     mtl: "false" # Excludes machine translated subtitles.
  #     hardsub: "false" # Excludes hardcoded subtitles.
  #     batch: "false" # Excludes batch torrents.
discovery:
  pollFrequency: 5m0s # Minimum 1m0s.
  sources: # OR filter. Specifies which sources to use, and in which priority. Keep empty to accept all.
      - source1
      - source2
  qualities: # OR filter. Example 1080 HEVC or 1080. It keeps priority.
      - 1080 HEVC # Can be used to prioritize HEVC for example.
      - 1080
  searchSuffix: "" # Specifies search suffixes like negative match on keywords.
  category: Animes # Animeman managed category, will be used to create tags / identify latest ep.
  downloadPath: /downloads/animes
  createShowFolder: true # Default: false. Creates a folder to for the show inside downloadPath.
  renameTorrent: true # Default: false. Renames the torrent entry on qbittorrent.
  renameScript: "" # Optional. See example below.
torrentClient:
  type: qbittorrent
  qbittorrent:
    host: http://ip:port
    username: username
    password: password
# Optional. Used for linking files when Shoko is unable to find suitable matches.
# Requires nekobt set, since it's the only way I have to correlate anilist or mal ids to anidb.
shoko:
  host: http://ip:port
  apiKey: YOUR_API_KEY
```

### Docs for renameScript

Language builtins are found under [expr-lang](https://expr-lang.org/docs/language-definition).  
Available envs: `[title, releaseGroup, labels, tag, verticalResolution]`.  
Available funcs:
- `format("mask", value)`: Use format masks like `%s` or `%v`, to render a value. Returns empty string on zero value

#### Example 1
```
// Joins parts with a single space.
join(
  // Filters only non-empty parts.
  filter(
    [
      // Put release group inside square brackets if any.
      format("[%s]", releaseGroup),
      title,
      // Format tag as a string. e.g. S1E4.
      tag.String(), 
      // Put vertical resolution inside square brackets + p, if any.
      format("[%s]", verticalResolution),
      // Format array of labels as [v1,v2,v3...]. 
      format("%v", labels),
    ], 
    # != "",
  ), 
  " ",
)
```

This script formats torrents as: `[release-group] My Anime Title S1E1 [1080p] [HEVC 10BIT]`

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

```docker run -it -e CONFIG_PATH=/config/config.yaml -v ./config:/config ghcr.io/sonalys/animeman:v3```

### Docker Compose

```yaml
# docker-compose.yaml
services:
  animeman:
    image: ghcr.io/sonalys/animeman:v3
    container_name: animeman
    environment:
      - CONFIG_PATH=/config/config.yaml
    volumes:
      - ./config:/config
```

`docker compose -f docker-compose.yaml up -d animeman`

## Building

### Dependencies

You will need at least go 1.27 for building the binary.  
For the image you will need docker.  
To build you can simply run `make build`  
For the image you can run `make image`

## Contribution

Please open issues first. Let's discuss what you need and then go for development.

## Disclaimer

This tool is not intended for any illegal activities.  
It's your responsibility to check your own jurisdiction.
