# CastSponsorSkip

<img src="./assets/icon.svg" alt="CastSponsorSkip Icon" align="right" width="90">

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/gabe565/CastSponsorSkip)](https://github.com/gabe565/CastSponsorSkip/releases)
[![Build](https://github.com/gabe565/CastSponsorSkip/actions/workflows/build.yml/badge.svg)](https://github.com/gabe565/CastSponsorSkip/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gabe565/CastSponsorSkip)](https://goreportcard.com/report/github.com/gabe565/CastSponsorSkip)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=gabe565_CastSponsorSkip&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=gabe565_CastSponsorSkip)

A Go program that skips sponsored YouTube content and skippable ads on all local Google Cast devices, using the [SponsorBlock](https://github.com/ajayyy/SponsorBlock) API. This project was inspired by [sponsorblockcast](https://github.com/nichobi/sponsorblockcast), but written from scratch to decrease memory and CPU usage, and to work around some of its problems (see [Differences from sponsorblockcast](#differences-from-sponsorblockcast)).

When run, this program will watch all Google Cast devices on the LAN. If a Cast device begins playing a YouTube video, sponsored segments are fetched from the SponsorBlock API. When the device reaches a sponsored segment, the CastSponsorSkip will quickly seek to the end of the segment.

Additionally, CastSponsorSkip will look for skippable YouTube ads, and automatically hit the skip button when it becomes available.

## Installation

### Docker

<details>
  <summary>Click to expand</summary>

  You can [install Docker](https://docs.docker.com/engine/install/) directly or use [Docker Compose](https://docs.docker.com/compose/install/) (Or use Podman, Portainer, etc). Please note you *MUST* use the `host` network as shown below for cli or in the example `docker-compose` file.

  #### Docker
  Run the below commands as root or a member of the `docker` group:
  ```shell
  docker run --network=host --name=castsponsorskip ghcr.io/gabe565/castsponsorskip
  ```

  #### Docker Compose
  First you will need a `docker-compose.yaml` file, such as the [one included in this repo](docker-compose.yaml). Run the below commands as root or a member of the `docker` group:
  ```shell
  docker compose up -d
  ```
</details>


### Home Assistant Addon

<details>
  <summary>Click to expand</summary>

A Home Assistant addon is provided by @alexbelgium and @bruvv. See [alexbelgium/hassio-addons](https://github.com/alexbelgium/hassio-addons/tree/master/sponsorblockcast) for installation instructions.
</details>


### APT (Ubuntu, Debian)

<details>
  <summary>Click to expand</summary>

1. If you don't have it already, install the `ca-certificates` package
   ```shell
   sudo apt install ca-certificates
   ```

2. Add gabe565 apt repository
   ```
   echo 'deb [trusted=yes] https://apt.gabe565.com /' | sudo tee /etc/apt/sources.list.d/gabe565.list
   ```

3. Update apt repositories
   ```shell
   sudo apt update
   ```

4. Install CastSponsorSkip
   ```shell
   sudo apt install castsponsorskip
   ```
</details>

### RPM (CentOS, RHEL)

<details>
  <summary>Click to expand</summary>

1. If you don't have it already, install the `ca-certificates` package
   ```shell
   sudo dnf install ca-certificates
   ```

2. Add gabe565 rpm repository to `/etc/yum.repos.d/gabe565.repo`
   ```ini
   [gabe565]
   name=gabe565
   baseurl=https://rpm.gabe565.com
   enabled=1
   gpgcheck=0
   ```

3. Install CastSponsorSkip
   ```shell
   sudo dnf install castsponsorskip
   ```
</details>

### AUR (Arch Linux)

<details>
  <summary>Click to expand</summary>

Install [castsponsorskip-bin](https://aur.archlinux.org/packages/castsponsorskip-bin) with your [AUR helper](https://wiki.archlinux.org/index.php/AUR_helpers) of choice.
</details>

### Homebrew (macOS, Linux)

<details>
  <summary>Click to expand</summary>

Install CastSponsorSkip from [gabe565/homebrew-tap](https://github.com/gabe565/homebrew-tap):
```shell
brew install gabe565/tap/castsponsorskip
```
</details>

### Kubernetes

<details>
  <summary>Click to expand</summary>

CastSponsorSkip doesn't require a Service or Ingress, so it's simple to host in Kubernetes with a single Deployment. The only gotcha is that `hostNetwork` must be enabled for device discovery to work.

A Helm chart is available to make Kubernetes deployment even simpler.
For more information, see
[charts.gabe565.com](https://charts.gabe565.com/charts/castsponsorskip/) or
[Artifact Hub](https://artifacthub.io/packages/helm/gabe565/castsponsorskip).
</details>

### Manual Installation

<details>
  <summary>Click to expand</summary>

Download and run the [latest release binary](https://github.com/gabe565/CastSponsorSkip/releases/latest) for your system and architecture.
</details>

## Usage
Run `castsponsorskip` from a terminal or activate the service with systemd:
```shell
systemctl enable --now castsponsorskip
````

<details>
  <summary>Homebrew Instructions</summary>

  Use [brew services](https://github.com/Homebrew/homebrew-services) to start CastSponsorSkip:
  ```shell
  brew services start castsponsorskip
  ```
</details>

## Configuration
CastSponsorSkip can be configured with envs, command-line flags, or a config file. Some notable envs are listed below, but all [flags](./docs/castsponsorskip.md) can be set with envs.  
To use an env that is not listed here, capitalize all characters, replace `-` with `_`, and prefix with `CSS_`. For example, `--paused-interval=1m` would become `CSS_PAUSED_INTERVAL=1m`.

### Notable Envs
| Env                     | Description                                                                                                                                                  | Default   |
|-------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------|
| `CSS_DISCOVER_INTERVAL` | Interval to restart the DNS discovery client.                                                                                                                | `5m`      |
| `CSS_PAUSED_INTERVAL`   | Time to wait between each poll of the Cast device status when paused.                                                                                        | `1m`      |
| `CSS_PLAYING_INTERVAL`  | Time to wait between each poll of the Cast device status when playing.                                                                                       | `500ms`   |
| `CSS_CATEGORIES`        | Comma-separated list of SponsorBlock categories to skip, see [category list](https://wiki.sponsor.ajay.app/w/Types#Category).                                | `sponsor` |
| `CSS_YOUTUBE_API_KEY`   | [YouTube API key](https://developers.google.com/youtube/registering_an_application) for fallback video identification (required on some Chromecast devices). | ` `       |
| `CSS_MUTE_ADS`          | Mutes the device while an ad is playing.                                                                                                                     | `true`    |
| `CSS_DEVICES`           | Comma-separated list of device addresses. This will disable discovery and is not recommended unless discovery fails.                                         | `[]`      |

> **Note**
> [sponsorblockcast envs](https://github.com/nichobi/sponsorblockcast#configuration) are also supported to simplify the migration to CastSponsorSkip. When used, a deprecation warning will be logged with an updated env key and value. There are currently no plans to remove these envs.

### Flags
See command-line flag documentation [here](./docs/castsponsorskip.md).

### Systemd
To modify the variables when running as a systemd service, create an override for the service with:

```shell
sudo systemctl edit castsponsorskip.service
```

This will open a blank override file where you can specify environment variables like so:
```
[Service]
Environment="CSS_PAUSED_INTERVAL=1m"
Environment="CSS_PLAYING_INTERVAL=1s"
Environment="CSS_CATEGORIES=sponsor,selfpromo"
```

To modify the variables when running as a Docker container, you can add arguments to the `docker run` command like so:

```shell
docker run --network=host --env CSS_PAUSED_INTERVAL=5m --env CSS_PLAYING_INTERVAL=2s --name=castsponsorskip ghcr.io/gabe565/castsponsorskip
```

### Docker Compose
When using `docker-compose.yaml`, you can simply edit the `environment` directive as shown in the [example file](./docker-compose.yaml).

## Differences from sponsorblockcast
- Uses the SponsorBlock [enhanced privacy endpoint](https://wiki.sponsor.ajay.app/w/API_Docs#GET_/api/skipSegments/:sha256HashPrefix). When searching for sponsored segments, the video ID is hashed and only the first 4 characters of the hash are passed to SponsorBlock. This allows CastSponsorSkip to fetch segments without telling SponsorBlock what video is being watched.
- Compiles to a single binary. No dependencies are required other than CastSponsorSkip.
- Scans Cast device status much less frequently when a YouTube video is not playing, resulting in decreased CPU usage and less stress on the Cast device.
- Written Go, which is the same language as `go-chromecast`. This means `go-chromecast` functions can be called directly instead of relying on shell scripts, child commands, or string parsing.
- `go-chromecast` only needs to be loaded once within a single Go program, resulting on lower memory usage.
- Dependency updates are automated with Renovate.

I own 12 Google Cast devices, and have compared CPU and memory usage of the two programs. Note that CPU usage is measured in "milliCPU", meaning that 1m is equal to 1/1000 of a CPU. Here are the averages:

| Program             | CPU | Memory |
|---------------------|-----|--------|
| sponsorblockcast    | 75m | 70Mi   |
| castsponsorskip | 1m  | 10Mi   |
