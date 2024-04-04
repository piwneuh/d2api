# d2api

Api for scheduling and managing Dota 2 matches.

## Requirements

1. need to have **.env** file in main `d2api` with below format
2. need to have **inventory.json** in main `d2api` folder with credentials, 2FA should be disabled for all accounts, one account should be enough for testing
3. need to have **REDIS** instance, either local or remote
4. go **1.21** or higher

Optional:

1. **Docker** and **docker-compose** for running redis and server in containers

### .ENV format

```
REDIS_PASSWORD=
REDIS_HOST=
REDIS_PORT=
REDIS_DB=
SERVER_PORT=
INVENTORY_PATH=
TIME_TO_CANCEL= //Time to cancel in seconds
```

### inventory.json format

```
[
  {
    "Username": "", // Username of steam account
    "Password": "" // Password of steam account
  }
]
```

## Start the server

### For development

```bash
# Start redis inside docker container
$ docker compose up redis -d

# Run the server
$ go run /cmd/main.go
```

If you have docker compose version 1.22 or higher

```bash
# Start both server and redis inside docker container
$ docker compose watch
```

Docker compose watch is setup so that any changes in the code will automatically rebuild the server, effectively giving you a hot reload.

### For production

```bash
# Start both server and redis inside docker container
$ docker compose up -d
```

## Endpoint documentation

### 1. POST /match

This endpoint schedules a match.

```json
{
  "teamA": [],
  "teamB": [],
  "lobbyConfig": {
    "gameName": "",
    "passKey": "",
    "serverRegion": 0,
    "gameMode": ""
  },
  "startTime": ""
}
```

- **teamA**: Array of Steam IDs for Team A.
- **teamB**: Array of Steam IDs for Team B.
- **lobbyConfig**:
  - **gameName**: Lobby name, send as string.
  - **passKey**: Password for the lobby, send as string, this is required so lobby is shown as public.
  - **serverRegion**: Server region of the match, send as number.
  - **gameMode**: Game mode of the match, send as string, see **Game Mode List** below.
- **startTime**: Start time of the match, format: `"2021-07-01T12:00:00Z"`, if not provided, match will be scheduled immediately.

returns `200` if match is scheduled successfully, `400` if match is not scheduled.

```json
{
  "matchIdx": 0
}
```

### 2. GET /match/:matchIdx

This endpoint returns the match details for the given match index `matchIdx`.
Keep in mind that match index is returned when a match is scheduled.
If match is not begun it will return lobby details, and match details can be fetched only when the match is finished.

**It will either return lobby details or match details based on the match status.**

returns `200` if match is found, `404` if match is not found.

### Game Mode List (send as game mode string in _ScheduleMatch_)

- **AP**: All Pick
- **CM**: Captain Mode
- **RD**: Random Draft
- **SD**: Single Draft
- **AR**: All Random
- **REVERSE_CM**: Reverse Captain Mode
- **MO**: Mid Only
- **LP**: Least played
- **CD**: Captains Draft
- **ABILITY_DRAFT**: Ability Draft
- **ARDM**: All Random Deathmatch
- **1V1MID**: 1v1 Solo Mid
- **ALL_DRAFT**: Ranked All Pick
- **TURBO**: Turbo
