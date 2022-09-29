# Clex

Utility to automatically cleanup watched shows and movies to save disk space and your time.

## Features:

-   Job interval is configurable (daily, weekly, monthly or once)
-   Background job to automatically detect watched media
-   Unmonitor and delete media from Sonarr and Radarr
-   Adds watched movies to exclusion list to prevent re-downloads from watched lists

## Installation

If you have go installed just run `make install`

Or just download the binary from the release and copy it to `/usr/local/bin/` if you want to use systemd service files. Other directories tend be problematic on systems with selinux enabled. If you use a different directory adjust the clex.service file.

## Usage

You can change the default interval and time (weekly every monday at 5:00AM) to daily or monthly.

E.g. `clex -interval daily -time 5:00PM` or `clex -interval monthly -time 17:00`
