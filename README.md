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

### Threads
`GET   /threads`

return all threads of 'authenticated' user

out: 
```
{ result: [
  {
    id: "...",
    owner: {
      id: "..."
    },
    msgs: [
      {
        from: "pinco@random.com",
        to: "pallo@random.com",
        body: "hello!"
      },
      {
        from: "pallo@random.com",
        to: "pinco@random.com",
        body: "Hi!"
      }
    ]
  },
  ...
  ],
  paginator: {
    next: "url to next page",
    prev: "url to prev page",
  }
```

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
  msg: {
    to: "pallo@random.com",
    from: "pinco@random.com",
    body: "hello!"
  }
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
      body: "hello!"
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
      body: "hello!"
    },
    {
      from: "pallo@random.com",
      to: "pinco@random.com",
      body: "Hi!"
    }
  ]
}
```

