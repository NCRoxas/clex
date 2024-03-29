# Clex

Utility to automatically cleanup watched shows and movies to save disk space and your time. \

Only Linux is supported at the moment!

## Features:

-   Job interval is configurable (daily, weekly, monthly or once)
-   Background job to automatically detect watched media
-   Unmonitor and delete media from Sonarr and Radarr
-   Adds watched movies to exclusion list to prevent re-downloads from watched lists

## Installation

If you have go installed clone this repository and run `make install`

Or just download the binary from the release and copy it to `~/.local/bin/` if you want to use systemd service files. Other directories tend be problematic on systems with selinux enabled. If you use a different directory adjust the clex.service file.
You can find an example service in this repository "clex.service". Copy it to `~/.config/systemd/user/clex.service`, enable it with `systemctl --user enable clex` and start it with `systemctl --user start clex`

## Config example

```yaml
hosts:
    plex: https://plex.example.com
    sonarr: https://sonarr.example.com
    radarr: https://radarr.example.com
tokens:
    plex: ...
    sonarr: ...
    radarr: ...
delete: true
exclude: true
client_id: ...
libraries:
    - Movies
    - TV Shows
```

## Usage

First time running the app will create a config file in the directory `~/.config/clex`. Just fill in the urls of your plex, sonarr, radarr instances, the sonarr/radarr api keys and which libraries you want to watch. The plex token will be automatically added when you click the authentication link after starting the app again. The client id can be ignored.

You can change the default interval (once).

E.g. `clex -interval daily -time 5:00PM` or `clex -interval monthly -time 17:00`
