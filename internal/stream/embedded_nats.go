package stream

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type EmbeddedNats struct {
	Server *server.Server
	Client *nats.Conn
	Stream nats.JetStreamContext
}

var embeddedIns = &EmbeddedNats{
	Server: nil,
	Client: nil,
	Stream: nil,
}

func StartEmbeddedServer(nodeName, bindAddress string) (*EmbeddedNats, error) {
	host, port, err := parseHostAndPort(bindAddress)
	if err != nil {
		return nil, err
	}

	opts := &server.Options{
		Host:               host,
		Port:               port,
		ServerName:         nodeName,
		StoreDir:           "/",
		NoSigs:             true,
		JetStream:          true,
		JetStreamDomain:    "embedded",
		JetStreamMaxMemory: -1,
		JetStreamMaxStore:  -1,
	}

	ns, err := server.NewServer(opts)
	if err != nil {
		return nil, err
	}

	ns.SetLogger(
		&natsLogger{log.With().Str("from", "nats").Logger()},
		opts.Debug,
		opts.Trace,
	)

	go ns.Start()
	if !ns.ReadyForConnections(10 * time.Second) {
		return nil, errors.New("NATS Server time out")
	}

	embeddedIns.Server = ns

	clientOpts := []nats.Option{
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(20),
		nats.ReconnectWait(3 * time.Second),
		nats.ConnectHandler(func(conn *nats.Conn) {
			log.Debug().Msgf("Connected to NATs")
		}),
		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			log.Debug().Msgf("Disconnected from NATs %v", err)
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			log.Debug().Msgf("Reconnected to NATs %v", conn.ConnectedUrl())
		}),
	}

	clientOpts = append(clientOpts, nats.InProcessServer(ns))

	nc, err := nats.Connect(ns.ClientURL(), clientOpts...)
	if err != nil {
		return nil, err
	}

	embeddedIns.Client = nc

	jsOpts := []nats.JSOpt{}
	js, err := nc.JetStream(jsOpts...)

	embeddedIns.Stream = js

	return embeddedIns, nil
}

func parseHostAndPort(adr string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(adr)
	if err != nil {
		return "", 0, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}

	return host, port, nil
}

func parseRemoteLeafOpts(isRoot bool, rootNatsURL string) []*server.RemoteLeafOpts {
	rootURL, _ := url.Parse(rootNatsURL)

	opts := []*server.RemoteLeafOpts{
		{
			URLs: []*url.URL{rootURL},
			Hub:  true,
		},
	}

	return opts
}
