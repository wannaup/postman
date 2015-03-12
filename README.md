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
{
  owner: {
    id: "..."
  },
  to: "pallo@random.com",
  from: "pinco@random.com",
  msg: "hello!"
}
```
out: 
``` 
{
  id: "...",
  owner: {
    id: "..."
  },
  msgs: [
    {
      from: "pinco@random.com",
      to: "pinco@random.com",
      msg: "hello!"
    }
  ]
}
```

`POST   /threads/<id>/reply`

reply with a new message

in:
``` 
{
  from: "pallo@random.com",
  msg: "Hi!"
}
```
out: 
``` 
{
  id: "...",
  owner: {
    id: "..."
  },
  msgs: [
    {
      from: "pinco@random.com",
      to: "pallo@random.com",
      msg: "hello!"
    },
    {
      from: "pallo@random.com",
      to: "pinco@random.com",
      msg: "Hi!"
    }
  ]
}
```

