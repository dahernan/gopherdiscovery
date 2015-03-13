package gopherdiscovery

import (
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/gdamore/mangos"
	"github.com/gdamore/mangos/protocol/pub"
	"github.com/gdamore/mangos/protocol/surveyor"
	"github.com/gdamore/mangos/transport/ipc"
	"github.com/gdamore/mangos/transport/tcp"
)

type Options struct {
	SurveyTime   time.Duration
	RecvDeadline time.Duration
	PollTime     time.Duration
}

type DiscoveryServer struct {
	urlServer string
	urlPubSub string
	opt       Options

	services *Services

	ctx    context.Context
	cancel context.CancelFunc
	sock   mangos.Socket
}

type Services struct {
	nodes     StringSet
	publisher *Publisher

	ctx   context.Context
	addCh chan StringSet
}

type Publisher struct {
	url string

	ctx  context.Context
	sock mangos.Socket

	publishCh chan []string
}

func Server(urlServer string, urlPubSub string, opt Options) (*DiscoveryServer, error) {
	var sock mangos.Socket
	var err error
	var publisher *Publisher

	ctx, cancel := context.WithCancel(context.Background())
	pubCtx, _ := context.WithCancel(ctx)
	publisher, err = NewPublisher(pubCtx, urlPubSub)
	if err != nil {
		return nil, err
	}

	servicesCtx, _ := context.WithCancel(ctx)
	services := NewServices(servicesCtx, publisher)

	sock, err = surveyor.NewSocket()
	if err != nil {
		return nil, err
	}

	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tcp.NewTransport())

	err = sock.Listen(urlServer)
	if err != nil {
		return nil, err
	}
	err = sock.SetOption(mangos.OptionSurveyTime, opt.SurveyTime)
	if err != nil {
		return nil, err
	}
	err = sock.SetOption(mangos.OptionRecvDeadline, opt.RecvDeadline)
	if err != nil {
		return nil, err
	}

	server := &DiscoveryServer{
		services: services,

		urlServer: urlServer,
		urlPubSub: urlPubSub,
		opt:       opt,

		ctx:    ctx,
		cancel: cancel,
		sock:   sock,
	}

	go server.run()
	return server, nil
}

func (d *DiscoveryServer) Cancel() {
	d.cancel()
}

func (d *DiscoveryServer) run() {
	for {
		select {
		case <-time.After(d.opt.PollTime):
			d.poll()
		case <-d.ctx.Done():
			return
		}
	}
}

func (d *DiscoveryServer) poll() error {
	var err error
	var msg []byte
	var responses StringSet

	fmt.Println("SERVER: SENDING DATE SURVEY REQUEST")
	err = d.sock.Send([]byte("DATE"))
	if err != nil {
		return err
	}

	responses = NewStringSet()
	for {
		fmt.Println("SERVER: WAITING FOR RESPONSES")

		msg, err = d.sock.Recv()
		if err != nil {
			if err == mangos.ErrRecvTimeout {
				fmt.Println("SERVER: Timeout ", err.Error())

				d.services.Add(responses)
				return nil
			}
			fmt.Println("SERVER: ERR", err.Error())

		} else {
			fmt.Println("SERVER: new client", string(msg))
			responses.Add(string(msg))
		}
	}

}

func NewPublisher(ctx context.Context, url string) (*Publisher, error) {
	var sock mangos.Socket
	var err error

	sock, err = pub.NewSocket()
	if err != nil {
		return nil, err
	}
	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tcp.NewTransport())

	err = sock.Listen(url)
	if err != nil {
		return nil, err
	}

	publiser := &Publisher{
		ctx:  ctx,
		url:  url,
		sock: sock,

		publishCh: make(chan []string),
	}

	go publiser.run()
	return publiser, nil
}

func (p *Publisher) Publish(msg []string) {
	p.publishCh <- msg
}

func (p *Publisher) run() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case msg := <-p.publishCh:
			fmt.Println("PUB: publish ", msg)
			err := p.sock.Send([]byte(strings.Join(msg, "|")))
			if err != nil {
				log.Println("PUB: Failed publishing:", err.Error())
			}
		}
	}
}

func NewServices(ctx context.Context, publisher *Publisher) *Services {
	s := &Services{
		nodes:     NewStringSet(),
		ctx:       ctx,
		addCh:     make(chan StringSet),
		publisher: publisher,
	}

	go s.run()
	return s
}

func (s *Services) run() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-s.addCh:
			fmt.Println("===: addCh ")
			s.add(msg)
		}
	}
}

func (s *Services) Add(responses StringSet) {
	fmt.Println("===: Add ")

	s.addCh <- responses
}

func (s *Services) add(responses StringSet) {
	s.nodes = responses.Clone()
	// publish the changes
	s.publisher.Publish(s.nodes.ToSlice())
}
