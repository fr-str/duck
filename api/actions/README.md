# actions

### RequestID - response will have same ID as Request
### Action - what action to perform
### Args - Args needed to perform action

## Live feed of containers and logs
```json 
{
    "RequestID": "abc",
    "Action": "live",
    "Args": {
        "containers":{}, // no args needed
        "logs":{
            "ContainerNames":["alpine"],
            "Amount":100
        }
    }
}
```

## container start/stop/restart/kill/inspect
# action: inspect will return code 208 and an object in Response.Data field
```json 
{
    "RequestID": "abc",
    "Action": "container.start",
    "Args": {
        "Name": "alpine",
    }
}
```

## get logs since/before timestamp
```json 
{
    "RequestID": "abc",
    "Action": "logs.get",
    "Args": {
        "ContainerName": "alpine",
        "Amount": 100,
        "Before":0, // unixNano time stamp
        "Since": 1666809208436109537 // unixNano time stamp
    }
}
```

## Response 
```json
{
    "RequestID": "abc",
    "Code": 200,
    "Data": "ok" //can be a object
}
```