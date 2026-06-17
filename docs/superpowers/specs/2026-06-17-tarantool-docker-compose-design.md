# Tarantool Docker Compose — Design Spec

**Date:** 2026-06-17
**Status:** Approved

## Goal

Add a Docker Compose setup so developers can spin up a local Tarantool 3.x instance for manual testing via `tmp/testdrive.go`.

## Scope

- Local development and testdrive use only. No integration tests are added.
- Not used by CI (existing tests are mock-based and need no container).

## Files

```
docker/
  config.yaml        # Tarantool 3.x native config
docker-compose.yml   # Single-service compose definition
```

## Architecture

One Docker Compose service: `tarantool`, based on `tarantool/tarantool:3`.

- Listens on `0.0.0.0:3301`, mapped to host `127.0.0.1:3301`
- User `flukeproxy` with password `flukeproxy-pwd` and `super` role (matches `tmp/testdrive.go`)
- Data persisted to a named Docker volume (`tarantool-data`)
- Config delivered via volume mount of `docker/config.yaml`

## Config (Tarantool 3.x YAML format)

```yaml
credentials:
  users:
    flukeproxy:
      password: 'flukeproxy-pwd'
      roles: [super]

iproto:
  listen:
    - uri: '0.0.0.0:3301'
```

Tarantool 3.x reads this file as its primary configuration when mounted to the expected path inside the container.

## Compose

```yaml
services:
  tarantool:
    image: tarantool/tarantool:3
    ports:
      - "127.0.0.1:3301:3301"
    volumes:
      - ./docker/config.yaml:/etc/tarantool/config.yaml
      - tarantool-data:/var/lib/tarantool
    restart: unless-stopped

volumes:
  tarantool-data:
```

## Usage

```bash
docker compose up -d
cd tmp && go run .
docker compose down
```
