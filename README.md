# postman
A mail to threaded messaging microservice

## Requirements

### Inbound
`POST  /inbound`

callback that manage inbound messages come from outside service

in:
``` 
```
out: 
``` 
```

### Users
`GET   /user/threads`

return all threads of 'authenticated' user

out: 
``` 
```

### Threads
`GET   /threads/:id`

return detail of a thread identified with id

out: 
``` 
```

`POST   /threads`

create a new thread

in:
``` 
```
out: 
``` 
```

`POST   /threads/<id>/reply`

reply with a new message

in:
``` 
```
out: 
``` 
```

