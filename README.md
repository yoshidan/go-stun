# go-stun 

* STUN (Session Traversal Utilities for NATs) Server.

<img src="https://blog.ivrpowers.com/postimages/technologies/ivrpowers-turn-stun-screen.005.jpeg" width="640px"/>

* [RFC 5389](https://tools.ietf.org/html/rfc5389) 
* Only `Binding Request` is supported.

## Requirement

* Go 1.15

## Server

```
go run cmd/server/main.go
```

## Deploy

If you run on docker or GKE `--network=host` is required.

