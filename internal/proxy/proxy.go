package proxy

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

const maxPacketSize = 1024 * 64

type Router interface {
	FromClient(client *Client, buf []byte) error
	FromRemote(client *Client, buf []byte) error
}

type Proxy struct {
	clients sync.Map
	router  Router
	logger  *slog.Logger
}

func New(logger *slog.Logger, router Router) *Proxy {
	return &Proxy{logger: logger, router: router}
}

func (p *Proxy) ListenAndServe(listenAddr, remoteAddrRaw string) error {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve listen address %q: %w", listenAddr, err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %q: %w", listenAddr, err)
	}
	defer conn.Close()

	remoteAddr, err := net.ResolveUDPAddr("udp", remoteAddrRaw)
	if err != nil {
		return fmt.Errorf("failed to resolve remote address %q: %w", remoteAddrRaw, err)
	}

	buf := make([]byte, maxPacketSize)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			p.logger.Error("failed to read from UDP", "error", err)
			continue
		}

		client, err := p.conn(clientAddr, remoteAddr, conn)
		if err != nil {
			p.logger.Error("failed to create proxy connection",
				"client", clientAddr.String(), "error", err)
			continue
		}

		if err := p.router.FromClient(client, buf[:n]); err != nil {
			p.logger.Error("failed to route from client",
				"client", clientAddr.String(), "error", err)
		}
	}
}

func (p *Proxy) conn(clientAddr, remoteAddr *net.UDPAddr, conn *net.UDPConn) (*Client, error) {
	key := clientAddr.String()

	if val, ok := p.clients.Load(key); ok {
		return val.(*Client), nil
	}

	remoteConn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial remote for %s: %w", key, err)
	}

	client := &Client{
		addr:       clientAddr,
		localConn:  conn,
		remoteConn: remoteConn,
	}

	p.clients.Store(key, client)
	p.logger.Debug("new client connection established", "client", key)

	go p.handleRemote(key, client)

	return client, nil
}

func (p *Proxy) handleRemote(key string, client *Client) {
	defer func() {
		p.logger.Debug("connection timeout, terminating", "client", key)
		client.remoteConn.Close()
		p.clients.Delete(key)
	}()

	buf := make([]byte, maxPacketSize)

	for {
		client.remoteConn.SetReadDeadline(time.Now().Add(10 * time.Second))

		n, err := client.remoteConn.Read(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				return
			}
			p.logger.Error("failed to read from remote",
				"client", client.addr.String(), "error", err)
			return
		}

		if err := p.router.FromRemote(client, buf[:n]); err != nil {
			p.logger.Error("failed to route from remote", "error", err)
		}
	}
}
