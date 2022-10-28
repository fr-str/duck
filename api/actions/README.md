# actions

### RequestID - response will have same ID as Request
### Action - what action to perform
### Data - data needed to perform action

## Live feed of containers
```json 
{
    "RequestID": "abc",
    "Action": "live.containers",
    "Data": {}
}
```

## Stream of given containers logs
```json 
{
    "RequestID": "abc",
    "Action": "live.logs",
    "Data": {
        "ContainerName":"alpine",
        "Amount":100
    }
}
```


## get logs since/before timestamp
```json 
{
    "RequestID": "abc",
    "Action": "logs.get",
    "Data": {
        "ContainerName": "alpine",
        "Amount": 100,
        "Before":0, // unixNano time stamp
        "Since": 1666809208436109537 // unixNano time stamp
    }
}
```
