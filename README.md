# postman 
A mail to threaded messaging microservice in Go and SCALA

## Requirements
Every request to the ```threads``` endpoint must be authenticated with basic HTTP auth header, the password **must not be empty** but actually it is not used/checked.
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

return all threads of 'authenticated' user based on the auth header user

out: 
```
[
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
] 
```

`GET   /threads/:id`

return detail of a thread identified with id verifying the owner of the thread is actually the authenticated user

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

`POST   /threads`

create a new thread, upon creation postman sends a mail containing the message *msg* to the *to* email address setting the sender as the *from* mail address and the *reply-to* field to the email address of the mail node (inbound.yourdomain.com).

in:
``` 
{
  
  from: "pinco@random.com",
  to: "pinco@random.com",
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
