# Running `hreysi watch` as a background service

`hreysi watch` runs in the foreground by default. To keep can't-miss capture
running for a repo across logout/reboot, install it as a service.

> Single-repo today. A global multi-repo daemon (`hreysi watch --all`) is a
> planned fast-follow; until then, install one service per repo you care about.

## macOS — launchd

Save as `~/Library/LaunchAgents/io.hreysi.watch.<repo>.plist` (replace the paths):

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>            <string>io.hreysi.watch.myrepo</string>
    <key>ProgramArguments</key>
    <array>
        <string>/opt/homebrew/bin/hreysi</string>
        <string>watch</string>
    </array>
    <key>WorkingDirectory</key> <string>/Users/you/code/myrepo</string>
    <key>RunAtLoad</key>        <true/>
    <key>KeepAlive</key>        <true/>
    <key>StandardOutPath</key>  <string>/tmp/hreysi-myrepo.log</string>
    <key>StandardErrorPath</key><string>/tmp/hreysi-myrepo.log</string>
</dict>
</plist>
```

```sh
launchctl load ~/Library/LaunchAgents/io.hreysi.watch.myrepo.plist    # start
launchctl unload ~/Library/LaunchAgents/io.hreysi.watch.myrepo.plist  # stop
```

## Linux — systemd (user service)

Save as `~/.config/systemd/user/hreysi-watch@.service`:

```ini
[Unit]
Description=hreysi watch for %I

[Service]
ExecStart=/usr/local/bin/hreysi watch
WorkingDirectory=/home/you/code/%I
Restart=always

[Install]
WantedBy=default.target
```

```sh
systemctl --user enable --now hreysi-watch@myrepo.service   # start + on boot
journalctl --user -u hreysi-watch@myrepo -f                 # tail
```
