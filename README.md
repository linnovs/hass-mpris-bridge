# Home Assistant Media Player `MPRIS` Bridge

This is a simple application that bridge Home Assistant's `media_player` entities to `D-BUS`'s
`MPRIS` player.

This will also implement `MPRIS` player control method, which allow user to control Home Assistant's
media player directly from their desktop environment. e.g., `playerctl` or `MPRIS` controller.

## `systemd` auto start

```systemd
[Unit]
Description=Home Assistant Media Player MPRIS Bridge

[Service]
Type=simple
Environment=HASS_URI=wss://{{YOUR_HASS_URI}}/api/websocket
Environment=HASS_TOKEN={{YOUR_HASS_LONG_LIVED_ACCESS_TOKEN}}
ExecStart=%h/.local/bin/hassmpris

[Install]
WantedBy=default.target
```
