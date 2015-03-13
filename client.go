package gopherdiscovery

import (
	"errors"
	"log"
	"strings"

	"github.com/gdamore/mangos"
	"github.com/gdamore/mangos/protocol/respondent"
	"github.com/gdamore/mangos/protocol/sub"

	"github.com/gdamore/mangos/transport/ipc"
	"github.com/gdamore/mangos/transport/tcp"
	"golang.org/x/net/context"
)

type DiscoveryClient struct {
	// url for the survey heartbeat
	// for example tcp://127.0.0.1:40007
	urlServer string
	// url for the Pub/Sub
	// in this url you are going to get the changes on the set of nodes
	// for example tcp://127.0.0.1:50007
	urlPubSub string

	// Service that needs to be discovered, for example for a web server could be
	// http://192.168.1.1:8080
	service string

	ctx    context.Context
	cancel context.CancelFunc
	sock   mangos.Socket

	subscriber *Subscriber
}

type Subscriber struct {
	// url for the Pub/Sub
	url string

	ctx  context.Context
	sock mangos.Socket

	changes chan []string
}

func Client(urlServer string, service string) (*DiscoveryClient, error) {
	return ClientWithSub(urlServer, "", service)
}

func ClientWithSub(urlServer string, urlPubSub string, service string) (*DiscoveryClient, error) {
	var sock mangos.Socket
	var err error
	var subscriber *Subscriber

	ctx, cancel := context.WithCancel(context.Background())

	if urlPubSub != "" {
		subCtx, _ := context.WithCancel(ctx)
		subscriber, err = NewSubscriber(subCtx, urlPubSub)
		if err != nil {
			return nil, err
		}
	}

	sock, err = respondent.NewSocket()
	if err != nil {
		return nil, err
	}

	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tcp.NewTransport())
	err = sock.Dial(urlServer)
	if err != nil {
		return nil, err
	}

	client := &DiscoveryClient{
		urlServer:  urlServer,
		urlPubSub:  urlPubSub,
		service:    service,
		ctx:        ctx,
		cancel:     cancel,
		sock:       sock,
		subscriber: subscriber,
	}

	go client.run()
	return client, nil
}

func (d *DiscoveryClient) Peers() (chan []string, error) {
	if d.subscriber == nil {
		return nil, errors.New("No subscribe url is provided to discover the Peers")
	}
	return d.subscriber.Changes(), nil
}

func (d *DiscoveryClient) Cancel() {
	d.cancel()
}

func (d *DiscoveryClient) run() {
	var err error
	for {
		_, err = d.sock.Recv()
		if err != nil {
			log.Println("DiscoveryClient: Cannot receive the SURVEY", err.Error())
		} else {
			select {
			case <-d.ctx.Done():
				return

			default:
				err = d.sock.Send([]byte(d.service))
				if err != nil {
					log.Println("DiscoveryClient: Cannot send the SURVEY response", err.Error())
				}
			}
		}
	}
}

func NewSubscriber(ctx context.Context, url string) (*Subscriber, error) {
	var sock mangos.Socket
	var err error

	sock, err = sub.NewSocket()
	if err != nil {
		return nil, err
	}
	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tcp.NewTransport())

	err = sock.Dial(url)
	if err != nil {
		return nil, err
	}
	// subscribes to everything
	err = sock.SetOption(mangos.OptionSubscribe, []byte(""))
	if err != nil {
		return nil, err
	}

	subscriber := &Subscriber{
		url:     url,
		ctx:     ctx,
		sock:    sock,
		changes: make(chan []string, 8),
	}

	go subscriber.run()
	return subscriber, nil
}

func (s *Subscriber) Changes() chan []string {
	return s.changes
}

func (s *Subscriber) run() {
	var msg []byte
	var err error

	for {
		select {
		case <-s.ctx.Done():
			close(s.changes)
			return
		default:
			msg, err = s.sock.Recv()
			if err != nil {
				log.Println("DiscoveryClient: Cannot SUBSCRIBE to the changes", err.Error())

			}

			// non-blocking send to the channel, discards changes if the channel is not ready
			select {
			case s.changes <- strings.Split(string(msg), "|"):
			default:
			}

		}
	}
}
