package gopherdiscovery

import (
	"testing"
	"time"

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

		client, err := Client(urlServ, urlPubSub, "client1")
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

		client, err := Client(urlServ, urlPubSub, "client1")
		So(err, ShouldBeNil)

		clients := <-client.Nodes()

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
		clientOne, err := Client(urlServ, urlPubSub, "client1")
		So(err, ShouldBeNil)

		// client2
		clientTwo, err := Client(urlServ, urlPubSub, "client2")
		So(err, ShouldBeNil)

		// client3
		clientThree, err := Client(urlServ, urlPubSub, "client3")
		So(err, ShouldBeNil)

		clients := <-clientOne.Nodes()

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		clients = <-clientTwo.Nodes()

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		clients = <-clientThree.Nodes()

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
		clientOne, err := Client(urlServ, urlPubSub, "client1")
		So(err, ShouldBeNil)

		clients := <-clientOne.Nodes()

		So(clients, ShouldContain, "client1")

		// client2
		clientTwo, err := Client(urlServ, urlPubSub, "client2")
		So(err, ShouldBeNil)

		clients = <-clientOne.Nodes()

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")

		// client3
		clientThree, err := Client(urlServ, urlPubSub, "client3")
		So(err, ShouldBeNil)

		clients = <-clientOne.Nodes()

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
		clientOne, err := Client(urlServ, urlPubSub, "client1")
		So(err, ShouldBeNil)

		// client2
		clientTwo, err := Client(urlServ, urlPubSub, "client2")
		So(err, ShouldBeNil)

		// client3
		clientThree, err := Client(urlServ, urlPubSub, "client3")
		So(err, ShouldBeNil)

		clients := <-clientOne.Nodes()

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		clientThree.Cancel()

		clients = <-clientOne.Nodes()

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")

		clientTwo.Cancel()

		clients = <-clientOne.Nodes()

		So(clients, ShouldContain, "client1")

		server.Cancel()
		clientOne.Cancel()

	})
}

func TestServerDiscoveryOnlyChanges(t *testing.T) {
	Convey("Discover multiple clients", t, func() {
		urlServ := "tcp://127.0.0.1:40007"
		urlPubSub := "tcp://127.0.0.1:50007"

		server, err := Server(urlServ, urlPubSub, defaultOpts)
		So(err, ShouldBeNil)

		// client1
		clientOne, err := Client(urlServ, urlPubSub, "client1")
		So(err, ShouldBeNil)

		// client2
		clientTwo, err := Client(urlServ, urlPubSub, "client2")
		So(err, ShouldBeNil)

		// client3
		clientThree, err := Client(urlServ, urlPubSub, "client3")
		So(err, ShouldBeNil)

		clients := <-clientOne.Nodes()

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		clients = <-clientTwo.Nodes()

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		clients = <-clientThree.Nodes()

		So(clients, ShouldContain, "client1")
		So(clients, ShouldContain, "client2")
		So(clients, ShouldContain, "client3")

		server.Cancel()
		clientOne.Cancel()
		clientTwo.Cancel()
		clientThree.Cancel()

	})
}
