# Gopherdiscovery: Simple Service Discovery for Go and nanomsg

This is a library to provide a simple way to do service discovery in Go, or other languages compatibles with [nanomsg](http://nanomsg.org/)/[mangos](https://github.com/gdamore/mangos)

The library was inspired by this blog post http://www.bravenewgeek.com/fast-scalable-networking-in-go-with-mangos/

# Install and Usage

```
go get github.com/dahernan/gopherdiscovery
```

```go
import "github.com/dahernan/gopherdiscovery"
```

## Design
[Design of the protocol](https://github.com/dahernan/gopherdiscovery/wiki/Design-of-the-protocol)

# Use cases

## Discover peers in a cluster

```go
	
var peers chan []string	
urlServer := "tcp://127.0.0.1:40007"
urlPubSub := "tcp://127.0.0.1:50007"

opts := Options{
		SurveyTime:   1 * time.Second,
		RecvDeadline: 1 * time.Second,
		PollTime:     2 * time.Second,
}

server, err := gopherdiscovery.Server(urlServer, urlPubSub, opts)

// client1
clientOne, err := gopherdiscovery.ClientWithSub(urlServer, urlPubSub, "client1")

// client2
clientTwo, err := gopherdiscovery.ClientWithSub(urlServer, urlPubSub, "client2")

// client3
clientThree, err := gopherdiscovery.ClientWithSub(urlServer, urlPubSub, "client3")

peers, err = clientOne.Peers()	
nodes <- peers
// nodes = []string{"client1", "client2", "client3"}

// Cancel client2
clientTwo.Cancel()

nodes <- peers
// nodes = []string{"client1", "client3"}

```

Read from the peers for changes

```go
peers, err = clientOne.Peers()	
for nodes := range peers {
	AddNodesToCluster(nodes)
}

```

## Subscribe to clients changes (new connections/disconnections)
```go

var clients []string
urlServ := "tcp://127.0.0.1:40009"
urlPubSub := "tcp://127.0.0.1:50009"

server, err := gopherdiscovery.Server(urlServ, urlPubSub, defaultOpts)

// 	"golang.org/x/net/context"
ctx, cancel := context.WithCancel(context.Background())
sub, err := gopherdiscovery.NewSubscriber(ctx, urlPubSub)

gopherdiscovery.Client(urlServ, "client1")
gopherdiscovery.Client(urlServ, "client2")

clients = <-sub.Changes()
// clients = []string{"client1", "client2"}	

gopherdiscovery.Client(urlServ, "client3")

clients = <-sub.Changes()
// clients = []string{"client1", "client2", "client3"}

cancel() // stops subscribe

```


## Update the peers in [groupcache](https://github.com/golang/groupcache)


```go

// Using gopherdiscovery to update the peers in groupcache

urlServer := "tcp://10.0.0.100:40007"
urlPubSub := "tcp://10.0.0.100:50007"
me := "http://10.0.0.1"

// on the server
server, err := gopherdiscovery.Server(urlServer, urlPubSub, opts)


// any of the peers
pool := groupcache.NewHTTPPool(me)
client, err := gopherdiscovery.ClientWithSub(urlServer, urlPubSub, me)

peers, err = client.Peers()	
for nodes := ranges peers {
	pool.Set(nodes...)	
}

```


## Update the proxies in a loadbalancer

// TODO

# Single Point of Failure
Yes, it is!, but you can spin up multiple servers if you want to try.

