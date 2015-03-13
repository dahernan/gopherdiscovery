# Gopherdiscovery: Simple Service Discovery for Go and nanomsg

This is a library to provides a simple way to do service discovery in Go, or other languages compatibles with [nanomsg](http://nanomsg.org/)/[mangos](https://github.com/gdamore/mangos)

# Install and Usage

```
go get github.com/dahernan/gopherdiscovery
```

```go
import "github.com/dahernan/gopherdiscovery"
```


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

## Subscribe to changes in clients connections/disconections
```go

var clients []string
urlServ := "tcp://127.0.0.1:40009"
urlPubSub := "tcp://127.0.0.1:50009"

server, err := Server(urlServ, urlPubSub, defaultOpts)

// 	"golang.org/x/net/context"
ctx, cancel := context.WithCancel(context.Background())
sub, err := NewSubscriber(ctx, urlPubSub)

Client(urlServ, "client1")
Client(urlServ, "client2")

clients = <-sub.Changes()
// clients = []string{"client1", "client2"}	

Client(urlServ, "client3")

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
peers := groupcache.NewHTTPPool(me)
client, err := gopherdiscovery.ClientWithSub(urlServer, urlPubSub, me)

peers, err = client.Peers()	
for nodes := ranges peers {
	peers.Set(nodes)	
}

```


## Update the proxies in a loadbalancer

// TODO

# Single Point of Failure
Yes, it is!, but you can spin up multiple servers if you want to try.

