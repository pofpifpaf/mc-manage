# MC Manage

MC Manage is a Go-based CLI interface for creating, running and managing minecraft servers.

It relies on a file-based system, scanning a server folder and starting, screening, stopping or killing the servers found in that folder. Its aim is to be as simple and reliable as possible.

GitHub page: https://github.com/pofpifpaf/mc-manage

Server types available in this version are :
 - Vanilla
 - Paper
 - Neoforge
 - Purpur
 - Fabric
 
## Features

- Auto restarts
- Start on boot
- Managing multiple servers
	- Create (with automatic jar downloading!)
	- Start/Stop/Kill
	- Screen
	- Import
- Docker installation, with no docker-in-docker
- Streamlining argument management
- Lightweight
- Multiple Java versions concurrently

Example:
```
Info for server: server-name

status: running

Type              paper
Version           1.20.6
Version Arg       144
Java Version      21
Memory Allocated  512M
Memory Max        4G
Jar file name     server.jar
Server port       25565
Level name used   world
Auto restarts     false
Boot              false
Size              209.4 MiB
Uptime            1m48s
Players           0/20
Memory Used       1.7 GiB

Additional JVM Args     1 - -XX:+AlwaysPreTouch
                        2 - -XX:+DisableExplicitGC
                        3 - -XX:+ParallelRefProcEnabled
                        4 - -XX:+PerfDisableSharedMem
                        5 - -XX:+UnlockExperimentalVMOptions
                        6 - -XX:+UseG1GC
                        7 - -XX:G1HeapRegionSize=8M
                        8 - -XX:G1HeapWastePercent=5
                        9 - -XX:G1MaxNewSizePercent=40
                        10 - -XX:G1MixedGCCountTarget=4
                        11 - -XX:G1MixedGCLiveThresholdPercent=90
                        12 - -XX:G1NewSizePercent=30
                        13 - -XX:G1RSetUpdatingPauseTimePercent=5
                        14 - -XX:G1ReservePercent=20
                        15 - -XX:InitiatingHeapOccupancyPercent=15
                        16 - -XX:MaxGCPauseMillis=200
                        17 - -XX:MaxTenuringThreshold=1
                        18 - -XX:SurvivorRatio=32
Additional Server Args  1 - example-argument


```

## Installation

### Docker run

```
docker run -dit \
  --name mc-manager \
  --restart unless-stopped \
  -p 25565:25565 \
  -v /path/to/server/files:/servers:rw \
  ghcr.io/pofpifpaf/manager:latest
```

### Docker Compose

```
services:
  mc-manage:
    tty: true
    container_name: mc-manager
    restart: unless-stopped
    image: ghcr.io/pofpifpaf/manager:latest
    ports:
      - 25565:25565
    volumes:
      - /path/to/server/files:/servers:rw
networks: {}
```

## Roadmap / Potential future features

- Other server type support
- MOTD generator and/or viewer
- Backups