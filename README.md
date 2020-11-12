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

```
go build cmd/server/main.go
./main
```

If you run on docker or GKE `--network=host` is required.

## Client

```go
import "github.com/yoshidan/go-stun/stun"

// discover once
func onetime() {
    client := NewClient(context.Background(), "stun.l.google.com:19302", laddr)
    result := client.Discover()
}

// keep alive
func keepalive() {
    client := NewClient(context.Background(), "stun.l.google.com:19302", laddr)
    err := client.Keepalive()
    if err !=nil {
        // handle err
    }
    defer client.Close()

    result := client.Discover()         

    //reuse connection
    result := client.Discover()         
}
```


