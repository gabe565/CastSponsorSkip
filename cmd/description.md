Skip sponsored YouTube segments on local Cast devices.

When run, this program will watch all Google Cast devices on the LAN.
If a Cast device begins playing a YouTube video, sponsored segments are fetched from the SponsorBlock API.  
When the device reaches a sponsored segment, CastSponsorSkip will quickly seek to the end of the segment.  
CastSponsorSkip will also mute YouTube ads and automatically hit the skip button when it becomes available.

All flags can be set using environment variables.  
To use an env, capitalize all characters, replace `-` with `_`, and prefix with `CSS_`.  
For example, `--paused-interval=1m` would become `CSS_PAUSED_INTERVAL=1m`.
