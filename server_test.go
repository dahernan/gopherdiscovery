package gopherdiscovery

import (
	"testing"
	"time"

	"golang.org/x/net/context"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	defaultOpts = Options{
		SurveyTime:   10 * time.Millisecond,
		RecvDeadline: 10 * time.Millisecond,
		PollTime:     20 * time.Millisecond,
	}
)

func TestServerCancel(t *testing.T) {
	Convey("Discovery server can be canceled", t, func() {
		urlServ := "tcp://127.0.0.1:40001"
		urlPubSub := "tcp://127.0.0.1:50001"

		server, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldBeNil)

		server.Cancel()

	})
}

func TestClientCancel(t *testing.T) {
	Convey("Discovery server and client can be canceled", t, func() {

		urlServ := "tcp://127.0.0.1:40002"
		urlPubSub := "tcp://127.0.0.1:50002"

		server, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldBeNil)

		client, err := ClientWithSub(urlServ, urlPubSub, "client1")
		So(err, ShouldBeNil)

		server.Cancel()
		client.Cancel()

	})
}

func TestServerDiscovery(t *testing.T) {
	Convey("Discover one client", t, func() {

		urlServ := "tcp://127.0.0.1:40003"
		urlPubSub := "tcp://127.0.0.1:50003"

		server, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldBeNil)

		client, err := ClientWithSub(urlServ, urlPubSub, "client1")
		So(err, ShouldBeNil)

		peers, err := client.Peers()
		So(err, ShouldBeNil)
		clients := <-peers

		So(clients, ShouldResemble, []string{"client1"})

		server.Cancel()
		client.Cancel()

	})
}

func TestServerDiscoveryMultipleClients(t *testing.T) {
	Convey("Discover multiple clients", t, func() {
		urlServ := "tcp://127.0.0.1:40004"
		urlPubSub := "tcp://127.0.0.1:50004"

		server, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldBeNil)

		// client1
		clientOne, err := ClientWithSub(urlServ, urlPubSub, "client1")
		So(err, ShouldBeNil)

		// client2
		clientTwo, err := ClientWithSub(urlServ, urlPubSub, "client2")
		So(err, ShouldBeNil)

		// client3
		clientThree, err := ClientWithSub(urlServ, urlPubSub, "client3")
		So(err, ShouldBeNil)

		peers, err := clientOne.Peers()
		So(err, ShouldBeNil)
		clients := <-peers

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		peers, err = clientTwo.Peers()
		So(err, ShouldBeNil)
		clients = <-peers

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		peers, err = clientThree.Peers()
		So(err, ShouldBeNil)
		clients = <-peers

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		server.Cancel()
		clientOne.Cancel()
		clientTwo.Cancel()
		clientThree.Cancel()

	})
}

func TestServerDiscoveryAddClients(t *testing.T) {
	Convey("Discover when you add more clients", t, func() {
		urlServ := "tcp://127.0.0.1:40005"
		urlPubSub := "tcp://127.0.0.1:50005"

		server, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldBeNil)

		// client1
		clientOne, err := ClientWithSub(urlServ, urlPubSub, "client1")
		So(err, ShouldBeNil)

		peers, err := clientOne.Peers()
		So(err, ShouldBeNil)
		clients := <-peers

		So(clients, ShouldContain, "client1")

		// client2
		clientTwo, err := ClientWithSub(urlServ, urlPubSub, "client2")
		So(err, ShouldBeNil)

		clients = <-peers

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")

		// client3
		clientThree, err := ClientWithSub(urlServ, urlPubSub, "client3")
		So(err, ShouldBeNil)

		clients = <-peers

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		server.Cancel()
		clientOne.Cancel()
		clientTwo.Cancel()
		clientThree.Cancel()

	})
}

func TestServerDiscoveryRemoveClients(t *testing.T) {
	Convey("Discover when you remove clients", t, func() {
		urlServ := "tcp://127.0.0.1:40006"
		urlPubSub := "tcp://127.0.0.1:50006"

		server, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldBeNil)

		// client1
		clientOne, err := ClientWithSub(urlServ, urlPubSub, "client1")
		So(err, ShouldBeNil)

		// client2
		clientTwo, err := ClientWithSub(urlServ, urlPubSub, "client2")
		So(err, ShouldBeNil)

		// client3
		clientThree, err := ClientWithSub(urlServ, urlPubSub, "client3")
		So(err, ShouldBeNil)

		peers, err := clientOne.Peers()
		So(err, ShouldBeNil)
		clients := <-peers

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		clientThree.Cancel()

		clients = <-peers

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")

		clientTwo.Cancel()

		clients = <-peers

		So(clients, ShouldContain, "client1")

		server.Cancel()
		clientOne.Cancel()

	})
}

func TestServerDiscoveryOnlyChanges(t *testing.T) {
	Convey("Publish msg only with changes", t, func() {
		urlServ := "tcp://127.0.0.1:40007"
		urlPubSub := "tcp://127.0.0.1:50007"

		server, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldBeNil)

		// client1
		clientOne, err := ClientWithSub(urlServ, urlPubSub, "client1")
		So(err, ShouldBeNil)

		// client2
		clientTwo, err := ClientWithSub(urlServ, urlPubSub, "client2")
		So(err, ShouldBeNil)

		// client3
		clientThree, err := ClientWithSub(urlServ, urlPubSub, "client3")
		So(err, ShouldBeNil)

		peers, err := clientOne.Peers()
		So(err, ShouldBeNil)
		clients := <-peers

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		time.Sleep(100 * time.Millisecond)

		select {
		case clients = <-peers:
			t.Fail()
		default:

		}

		server.Cancel()
		clientOne.Cancel()
		clientTwo.Cancel()
		clientThree.Cancel()

	})
}

func TestClientAndSubError(t *testing.T) {
	Convey("Client without subscribe gives an error if you call Peers", t, func() {

		urlServ := "tcp://127.0.0.1:40008"
		urlPubSub := "tcp://127.0.0.1:50008"

		server, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldBeNil)

		client, err := Client(urlServ, "client1")
		So(err, ShouldBeNil)

		_, err = client.Peers()
		So(err, ShouldNotBeNil)

		server.Cancel()
		client.Cancel()

	})
}

func TestClientAndIndependentSub(t *testing.T) {
	Convey("Gets the changes from a Subscriber", t, func() {

		urlServ := "tcp://127.0.0.1:40009"
		urlPubSub := "tcp://127.0.0.1:50009"

		server, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldBeNil)

		ctx, cancel := context.WithCancel(context.Background())
		sub, err := NewSubscriber(ctx, urlPubSub)
		So(err, ShouldBeNil)

		client, err := Client(urlServ, "client1")
		So(err, ShouldBeNil)

		clients := <-sub.Changes()
		So(clients, ShouldContain, "client1")

		server.Cancel()
		cancel()
		client.Cancel()

	})
}

func TestBadUrlServer(t *testing.T) {
	Convey("Discovery with bad url", t, func() {
		urlServ := "tcp://xxx"
		urlPubSub := "tcp://127.0.0.1:50010"

		_, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldNotBeNil)

		_, err = Client(urlServ, "client1")
		So(err, ShouldNotBeNil)

	})
}

func TestBadUrlPubSub(t *testing.T) {
	Convey("Discovery with bad url", t, func() {
		urlServ := "tcp://127.0.0.1:40011"
		urlPubSub := "tcp://xxx"

		_, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldNotBeNil)

		_, err = ClientWithSub(urlServ, urlPubSub, "client1")
		So(err, ShouldNotBeNil)

	})
}
