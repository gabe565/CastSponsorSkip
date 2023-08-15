# CastSponsorSkip

A Go program that skips sponsored YouTube content and skippable ads on all local Google Cast devices, using the [SponsorBlock](https://github.com/ajayyy/SponsorBlock) API. This project was inspired by [sponsorblockcast](https://github.com/nichobi/sponsorblockcast), but written from scratch to decrease memory and CPU usage, and to work around some of its problems (see [Differences from sponsorblockcast](#differences-from-sponsorblockcast)).

This program will scan for all Google Cast devices on the LAN, and runs a Goroutine for each one to efficiently poll its status every minute. If a Cast device is found to be playing a YouTube video, the poll interval is increased to once every second, sponsored segments are fetched from the SponsorBlock API and stored in memory. Whenever the Cast device reaches a sponsored segment, the program tells it to seek to the end of the segment.

Additionally, CastSponsorSkip will look for skippable YouTube ads, and automatically hit the skip button when it becomes available.

## Installation

### Docker Image

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


### APT Repository (Ubuntu, Debian)

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

### RPM Repository (CentOS, RHEL)

<details>
  <summary>Click to expand</summary>

1. If you don't have it already, install the `ca-certificates` package
   ```shell
   sudo yum install ca-certificates
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
   sudo yum install castsponsorskip
   ```
</details>

### Homebrew (macOS, Linux)

<details>
  <summary>Click to expand</summary>

  ```shell
  brew install gabe565/tap/castsponsorskip
  ```
</details>

### Manual Installation

<details>
  <summary>Click to expand</summary>

  #### Instructions

  Download and run the [latest release binary](https://github.com/gabe565/CastSponsorSkip/releases/latest) for your system and architecture.
</details>

## Usage
Run `castsponsorskip` from a terminal or activate the service with `systemctl enable --now castsponsorskip`.

## Configuration
You can configure the following parameters by setting the appropriate command line flag or environment variable:

| Flag                 | Env                                   | Description                                                                                                                                                        | Default        |
|----------------------|---------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------|
| `--paused-interval`  | `SBC_PAUSED_INTERVAL`                 | Time to wait between each poll of the Cast device status when paused.                                                                                              | `1m`           |
| `--playing-interval` | `SBC_PLAYING_INTERVAL`                | Time to wait between each poll of the Cast device status when playing.                                                                                             | `1s`           |
| `--categories`       | `SBC_CATEGORIES` (or `SBCCATEGORIES`) | Comma-separated (or space-separated) SponsorBlock categories to skip, see [category list](https://github.com/ajayyy/SponsorBlock/blob/master/config.json.example). | `sponsor`      |
| `--interface`        | `SBC_INTERFACE`                       | Optionally configure the network interface to use.                                                                                                                 | All interfaces |

To modify the variables when running as a systemd service, create an override for the service with:

```shell
sudo systemctl edit castsponsorskip.service
```

This will open a blank override file where you can specify environment variables like so:
```
[Service]
Environment="SBC_PAUSED_INTERVAL=1m"
Environment="SBC_PLAYING_INTERVAL=1s"
Environment="SBC_CATEGORIES=sponsor,selfpromo"
```

To modify the variables when running as a Docker container, you can add arguments to the `docker run` command like so:

```shell
docker run --network=host --env SBC_PAUSED_INTERVAL=5m --env SBC_PLAYING_INTERVAL=2s --name=castsponsorskip ghcr.io/gabe565/castsponsorskip
```

When using `docker-compose.yaml`, you can simply edit the `environment` directive as shown in the example file.

## Differences from sponsorblockcast
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
