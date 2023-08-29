## castsponsorskip

Skip sponsored YouTube segments on local Cast devices

### Synopsis

Skip sponsored YouTube segments on local Cast devices.

When run, this program will watch all Google Cast devices on the LAN.
If a Cast device begins playing a YouTube video, sponsored segments are fetched from the SponsorBlock API.
When the device reaches a sponsored segment, the CastSponsorSkip will quickly seek to the end of the segment.

Additionally, CastSponsorSkip will look for skippable YouTube ads, and automatically hit the skip button when it becomes available.

```
castsponsorskip [flags]
```

### Options

```
      --action-types strings         SponsorBlock action types to handle. Shorter segments that overlap with content can be muted instead of skipped. (default [skip,mute])
  -c, --categories strings           Comma-separated list of SponsorBlock categories to skip (default [sponsor])
      --completion string            Output command-line completion code for the specified shell. Can be 'bash', 'zsh', 'fish', or 'powershell'.
      --discover-interval duration   Interval to restart the DNS discovery client (default 5m0s)
  -h, --help                         help for castsponsorskip
      --log-level string             Log level (debug, info, warn, error) (default "info")
      --mute-ads                     Enables experimental support for muting unskippable ads
  -i, --network-interface string     Network interface to use for multicast dns discovery
      --paused-interval duration     Interval to scan paused devices (default 1m0s)
      --playing-interval duration    Interval to scan playing devices (default 500ms)
  -v, --version                      version for castsponsorskip
      --youtube-api-key string       YouTube API key for fallback video identification (required on some Chromecast devices).
```

