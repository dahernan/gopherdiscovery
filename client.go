package gopherdiscovery

import (
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
	urlServer string
	urlPubSub string
	service   string

	ctx    context.Context
	cancel context.CancelFunc
	sock   mangos.Socket

	subscriber *Subscriber
}

type Subscriber struct {
	url string

	ctx  context.Context
	sock mangos.Socket

	changes chan []string
}

func Client(urlServer string, urlPubSub string, service string) (*DiscoveryClient, error) {
	var sock mangos.Socket
	var err error
	var subscriber *Subscriber

	ctx, cancel := context.WithCancel(context.Background())
	subCtx, _ := context.WithCancel(ctx)
	subscriber, err = NewSubscriber(subCtx, urlPubSub)
	if err != nil {
		return nil, err
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

func (d *DiscoveryClient) Nodes() chan []string {
	return d.subscriber.Changes()
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
			return
		default:
			msg, err = s.sock.Recv()
			if err != nil {
				log.Println("DiscoveryClient: Cannot SUBSCRIBE to the changes", err.Error())

			}

			// non-blocking send to the channel
			select {
			case s.changes <- strings.Split(string(msg), "|"):
			default:
			}

		}
	}
}
