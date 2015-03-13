package gopherdiscovery

import (
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
	// SurveyTime is used to indicate the deadline for survey
	// responses
	SurveyTime time.Duration
	// RecvDeadline is the time until the next recived of the SURVEY times out.
	RecvDeadline time.Duration
	// PollTime is minimal time between SURVEYS (The time between SURVEYS could be greater than this time
	// if the SURVEY process takes longer than that time)
	PollTime time.Duration
}

type DiscoveryServer struct {
	// url for the survey heartbeat
	// for example tcp://127.0.0.1:40007
	urlServer string
	// url for the Pub/Sub
	// in this url you are going to get the changes on the set of nodes
	// for example tcp://127.0.0.1:50007
	urlPubSub string

	// Time options
	opt Options

	// Set of the services that has been discovered
	services *Services

	ctx    context.Context
	cancel context.CancelFunc
	sock   mangos.Socket
}

type Services struct {
	// set of nodes discovered
	nodes StringSet
	// publisher, we are going to publish the changes of the set here
	publisher *Publisher
}

type Publisher struct {
	// url for pub/sub
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

	pubCtx, pubCancel := context.WithCancel(ctx)
	publisher, err = NewPublisher(pubCtx, urlPubSub)
	if err != nil {
		pubCancel()
		return nil, err
	}
	services := NewServices(publisher)

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

// Shutdown the server
func (d *DiscoveryServer) Cancel() {
	d.cancel()
}

// Waits until the server finish running
func (d *DiscoveryServer) Wait() {
	<-d.ctx.Done()
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

func (d *DiscoveryServer) poll() {
	var err error
	var msg []byte
	var responses StringSet

	err = d.sock.Send([]byte(""))
	if err != nil {
		log.Println("DiscoveryServer: Error sending the SURVEY", err.Error())
		return
	}

	responses = NewStringSet()
	for {
		msg, err = d.sock.Recv()
		if err != nil {
			if err == mangos.ErrRecvTimeout {
				// Timeout means I can add the current responses to the SET
				d.services.Add(responses)
				return
			}
			log.Println("DiscoveryServer: Error reading SURVEY responses", err.Error())
		} else {
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
			close(p.publishCh)
			return
		case msg := <-p.publishCh:
			err := p.sock.Send([]byte(strings.Join(msg, "|")))
			if err != nil {
				log.Println("DiscoveryServer: Error PUBLISHING changes to the socket", err.Error())
			}
		}
	}
}

func NewServices(publisher *Publisher) *Services {
	s := &Services{
		nodes:     NewStringSet(),
		publisher: publisher,
	}

	return s
}

func (s *Services) Add(responses StringSet) {
	removed := s.nodes.Difference(responses)
	added := responses.Difference(s.nodes)

	// Do not publish anything if there is no changes
	if removed.Cardinality() == 0 && added.Cardinality() == 0 {
		return
	}

	s.nodes = responses
	// publish the changes
	s.publisher.Publish(s.nodes.ToSlice())
}
