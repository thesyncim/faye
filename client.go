package fayec

import (
	"github.com/thesyncim/faye/message"
	"github.com/thesyncim/faye/subscription"
	"github.com/thesyncim/faye/transport"
	_ "github.com/thesyncim/faye/transport/websocket"
)

type options struct {
	transport transport.Transport
	tOpts     transport.Options
}

var defaultOpts = options{
	transport: transport.GetTransport("websocket"),
}

//https://faye.jcoglan.com/architecture.html
type client interface {
	Disconnect() error
	Subscribe(subscription string) (*subscription.Subscription, error)
	//Unsubscribe(subscription string) error
	Publish(subscription string, message message.Data) (string, error)
	OnPublishResponse(subscription string, onMsg func(message *message.Message))
	OnTransportError(onErr func(err error))
}

//Option set the Client options, such as Transport, message extensions,etc.
type Option func(*options)

var _ client = (*Client)(nil)

// Client represents a client connection to an faye server.
type Client struct {
	opts options
}

//NewClient creates a new faye client with the provided options and connect to the specified url.
func NewClient(url string, opts ...Option) (*Client, error) {
	var c Client
	c.opts = defaultOpts
	for _, opt := range opts {
		opt(&c.opts)
	}

	var err error
	if err = c.opts.transport.Init(url, &c.opts.tOpts); err != nil {
		return nil, err
	}

	if err = c.opts.transport.Handshake(); err != nil {
		return nil, err
	}

	if err = c.opts.transport.Connect(); err != nil {
		return nil, err
	}

	return &c, nil
}

//Subscribe informs the server that messages published to that channel are delivered to itself.
func (c *Client) Subscribe(subscription string) (*subscription.Subscription, error) {
	return c.opts.transport.Subscribe(subscription)
}

//Publish publishes events on a channel by sending event messages, the server MAY  respond to a publish event
//if this feature is supported by the server use the OnPublishResponse to get the publish status.
func (c *Client) Publish(subscription string, data message.Data) (id string, err error) {
	return c.opts.transport.Publish(subscription, data)
}

//OnPublishResponse sets the handler to be triggered if the server replies to the publish request.
//According to the spec the server MAY reply to the publish request, so its not guaranteed that this handler will
//ever be triggered.
//can be used to identify the status of the published request and for example retry failed published requests.
func (c *Client) OnPublishResponse(subscription string, onMsg func(message *message.Message)) {
	c.opts.transport.OnPublishResponse(subscription, onMsg)
}

func (c *Client) OnTransportError(onErr func(err error)) {
	c.opts.transport.OnError(onErr)
}

//Disconnect closes all subscriptions and inform the server to remove any client-related state.
//any subsequent method call to the client object will result in undefined behaviour.
func (c *Client) Disconnect() error {
	return c.opts.transport.Disconnect()
}

//WithOutExtension append the provided outgoing extension to the the default transport options
//extensions run in the order that they are provided
func WithOutExtension(extension message.Extension) Option {
	return func(o *options) {
		o.tOpts.Extensions.Out = append(o.tOpts.Extensions.Out, extension)
	}
}

//WithExtension append the provided incoming extension and outgoing to the list of incoming and outgoing extensions.
//extensions run in the order that they are provided
func WithExtension(inExt message.Extension, outExt message.Extension) Option {
	return func(o *options) {
		o.tOpts.Extensions.In = append(o.tOpts.Extensions.In, inExt)
		o.tOpts.Extensions.Out = append(o.tOpts.Extensions.Out, outExt)
	}
}

//WithInExtension append the provided incoming extension to the list of incoming extensions.
//extensions run in the order that they are provided
func WithInExtension(extension message.Extension) Option {
	return func(o *options) {
		o.tOpts.Extensions.In = append(o.tOpts.Extensions.In, extension)
	}
}

//WithTransport sets the client transport to be used to communicate with server.
func WithTransport(t transport.Transport) Option {
	return func(o *options) {
		o.transport = t
	}
}
