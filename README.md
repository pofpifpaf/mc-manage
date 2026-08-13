# MC Manage

MC Manage is a Go-based CLI interface for creating, running and managing minecraft servers.

It relies on a file-based system, scanning a server folder and starting, screening, stopping or killing the servers found in that folder. Its aim is to be as simple and reliable as possible.

GitHub page: https://github.com/pofpifpaf/mc-manage

Server types available in this version are :
 - Vanilla
 - Paper
 - Neoforge
 
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

```
status: not running

Type              vanilla
Version           1.16.5
Java Version      21
Memory Allocated  512M
Memory Max        4G
Jar file name     server.jar
Server port       25565
Level name used   world
Auto restarts     false
Boot              false
Size              36.2 MiB
Uptime            -
Players           -/-
Memory Used       -

Additional JVM Args     -
Additional Server Args  -
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

- Other server type support (Neoforge, Forge, Paper...)
- MOTD generator and/or viewer
- Backups