# Gopherdiscovery: Simple Service Discovery for Go and nanomsg

This is a library to provides a simple way to do service discovery in Go, or other languages compatibles with nanomsg


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

# TODO

* Only client subscribe example
* Load balancer example


* Groupcache example

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



