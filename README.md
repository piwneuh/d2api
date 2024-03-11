# d2api

Test for go-dota2 ad go-steam

1. need to have **inventory.json** in main `d2api` folder with credentials, as well as **.env** file

2. need to have **REDIS** (docker compose or your machine) running

3. run `go run main.go`

## .ENV format

```
REDIS_PASSWORD=
REDIS_HOST=
REDIS_PORT=
REDIS_DB=
SERVER_PORT=
INVENTORY_PATH=
```

## inventory.json format

```
[
  {
    "Username": "", // Username of steam account
    "Password": "" // Password of steam account
  }
]
```
