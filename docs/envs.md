# Environment Variables

| Name | Usage | Default |
| --- | --- | --- |
| `CSS_ACTION_TYPES` | SponsorBlock action types to handle. Shorter segments that overlap with content can be muted instead of skipped. | `skip,mute` |
| `CSS_CATEGORIES` | Comma-separated list of SponsorBlock categories to skip | `sponsor` |
| `CSS_DEVICES` | Comma-separated list of device addresses. This will disable discovery and is not recommended unless discovery fails | ` ` |
| `CSS_DISCOVER_INTERVAL` | Interval to restart the DNS discovery client | `5m0s` |
| `CSS_IGNORE_SEGMENT_DURATION` | Ignores the previous sponsored segment for a set amount of time. Useful if you want to to go back and watch a segment. | `1m0s` |
| `CSS_LOG_FORMAT` | Log format (one of: auto, color, plain, json) | `auto` |
| `CSS_LOG_LEVEL` | Log level (one of: debug, info, warn, error, none) | `info` |
| `CSS_MUTE_ADS` | Mutes the device while an ad is playing | `true` |
| `CSS_NETWORK_INTERFACE` | Network interface to use for multicast dns discovery. (default all interfaces) | ` ` |
| `CSS_PAUSED_INTERVAL` | Interval to scan paused devices | `1m0s` |
| `CSS_PLAYING_INTERVAL` | Interval to scan playing devices | `500ms` |
| `CSS_SKIP_DELAY` | Delay skipping the start of a segment | `0s` |
| `CSS_SKIP_SPONSORS` | Skip sponsored segments with SponsorBlock | `true` |
| `CSS_YOUTUBE_API_KEY` | YouTube API key for fallback video identification (required on some Chromecast devices). | ` ` |