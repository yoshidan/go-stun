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
    client := stun.NewClient(context.Background(), "stun.l.google.com:19302", nil)
    result := client.Discover()
}

// keep alive
func keepalive() {
    client := stun.NewClient(context.Background(), "stun.l.google.com:19302", nil)
    err := client.Keepalive()
    if err !=nil {
        // handle err
    }
    defer client.Close()

    addr, err := client.Discover()         

    //reuse connection
    addr2, err2 := client.Discover()         
}
```


