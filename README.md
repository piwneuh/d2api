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
TIME_TO_CANCEL= //Time to cancel in seconds
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

# Game Mode List (send as game mode string in *ScheduleMatch*)
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