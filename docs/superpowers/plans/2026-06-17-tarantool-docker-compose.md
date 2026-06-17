# Tarantool Docker Compose Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Docker Compose setup that starts a local Tarantool 3.x instance on port 3301 with the `flukeproxy` user, matching the credentials in `tmp/testdrive.go`.

**Architecture:** Two files are added — `docker/config.yaml` (Tarantool 3.x native YAML config) and `docker-compose.yml` (single-service compose). No application code changes. No CI changes. Data is persisted to a named Docker volume.

**Tech Stack:** Docker Compose, Tarantool 3.x (`tarantool/tarantool:3` image)

## Global Constraints

- Tarantool image: `tarantool/tarantool:3` (exact tag prefix)
- Host port binding: `127.0.0.1:3301:3301` (loopback only — no `0.0.0.0` exposure)
- User: `flukeproxy`, password: `flukeproxy-pwd`, role: `super`
- Config path inside container: `/etc/tarantool/config.yaml`
- Data volume: named volume `tarantool-data` at `/var/lib/tarantool`

---

### Task 1: Add Tarantool config and Docker Compose

**Files:**
- Create: `docker/config.yaml`
- Create: `docker-compose.yml`

**Interfaces:**
- Consumes: nothing
- Produces: a running Tarantool 3.x instance accessible at `127.0.0.1:3301` with user `flukeproxy` / `flukeproxy-pwd`

- [ ] **Step 1: Create `docker/config.yaml`**

Create the file `docker/config.yaml` with this exact content:

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

- [ ] **Step 2: Create `docker-compose.yml`**

Create the file `docker-compose.yml` at the repo root with this exact content:

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

- [ ] **Step 3: Start the container and verify it is healthy**

Run:
```bash
docker compose up -d
```
Expected: image pull + container start with no errors, exit 0.

Then check the container is running:
```bash
docker compose ps
```
Expected: one row with `tarantool` in state `running` (or `Up`).

Then confirm the port is open:
```bash
docker compose logs tarantool
```
Expected: log lines showing Tarantool started and is listening (no `FATAL` or `ERROR` lines). Look for something like `entering the event loop` or `ready to accept requests`.

- [ ] **Step 4: Run testdrive against the live container**

```bash
cd tmp && go run .
```
Expected: exits 0 with no panic. If `DebugLogger` output is visible, you should see migration steps logged.

- [ ] **Step 5: Stop and clean up**

```bash
docker compose down
```
Expected: container stopped and removed, exit 0. Volume is preserved (intentional — `down` without `-v` keeps data).

- [ ] **Step 6: Commit**

```bash
git add docker/config.yaml docker-compose.yml
git commit -m "chore: add docker compose for local tarantool dev"
```
