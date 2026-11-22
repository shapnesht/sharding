# Sharding

A small URL shortener in Go that spreads its data across four PostgreSQL shards using consistent hashing. It is a learning project for database sharding.

## How it works

```
POST /?url=<long-url>
        |
        v
  sha256(url)[0:5]  -->  short id (5 hex chars)
        |
        v
  consistent hash ring (xxhash)  -->  shard "5433" | "5434" | "5435" | "5436"
        |
        v
  INSERT INTO Urls(url, url_id)  on that shard
```

- The short id is the first 5 hex characters of the SHA-256 of the URL.
- The short id is the shard key. [`buraksezer/consistent`](https://github.com/buraksezer/consistent) picks the shard, so adding or removing a shard only moves a fraction of the keys.
- Each shard is an independent Postgres instance with the same schema (see [init.sql](init.sql)).
- Shards are named after the host port they listen on.

## Project layout

| File | Purpose |
|---|---|
| [main.go](main.go) | HTTP server, shard connections, consistent hash ring |
| [Dockerfile](Dockerfile) | `postgres` image with `init.sql` baked in, used for every shard |
| [init.sql](init.sql) | Creates the `Urls` table on first start of a shard |
| `.env` | Database settings (not committed) |

## Prerequisites

- Go 1.27 or later (see [go.mod](go.mod))
- Podman or Docker

## Setup

### 1. Build the shard image

```sh
podman build -t pgshard .
```

### 2. Start four shards

Each shard needs a superpassword, otherwise the Postgres image exits on first start. The host ports must match the ones hard-coded in [main.go](main.go): 5433 to 5436.

```sh
podman run -d --name pgshard1 -p 5433:5432 -e POSTGRES_PASSWORD=postgres pgshard
podman run -d --name pgshard2 -p 5434:5432 -e POSTGRES_PASSWORD=postgres pgshard
podman run -d --name pgshard3 -p 5435:5432 -e POSTGRES_PASSWORD=postgres pgshard
podman run -d --name pgshard4 -p 5436:5432 -e POSTGRES_PASSWORD=postgres pgshard
```

`init.sql` runs only when a shard starts with an empty data directory, so every new shard gets the `Urls` table automatically.

### 3. Configure the app

Create a `.env` file in the project root. All four shards share these settings:

```
DB_HOST=localhost
DB_NAME=postgres
DB_USERNAME=postgres
DB_PASSWORD=postgres
```

### 4. Run

```sh
go run .
```

The server listens on `:8080`.

## API

### Create a short URL

```sh
curl -X POST "http://localhost:8080/?url=https://example.com/some/long/path"
```

```json
{
  "url": "https://example.com/some/long/path",
  "shortUrl": "a1b2c",
  "server": "5434"
}
```

`server` shows which shard stored the row.

### Resolve a short URL

`GET /{shortUrl}` is **not implemented yet**. The route and the lookup are still being built.

## Resetting the shards

Removing a container with `-v` also deletes its data volume:

```sh
podman rm -f -v pgshard1 pgshard2 pgshard3 pgshard4
```

## Known limitations

- No `GET /{shortUrl}` redirect yet.
- Shard ports and the shard list are hard-coded in [main.go](main.go).
- Each shard uses a single connection, which is not safe for concurrent requests. A pool (`pgxpool`) would fix this.
- There is no unique constraint on `url_id`, so a hash collision or a repeated URL inserts a duplicate row.
- Changing the number of shards moves keys between shards, and there is no migration tooling.
