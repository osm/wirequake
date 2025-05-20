package proxy

import (
	"net"
	"sync"
)

type Client struct {
	addr       *net.UDPAddr
	localConn  *net.UDPConn
	remoteConn *net.UDPConn

	connectedMu sync.Mutex
	connected   bool

	queueMu sync.Mutex
	queue   [][]byte
}

func (c *Client) Enqueue(buf []byte) {
	tmp := make([]byte, len(buf))
	copy(tmp, buf)

	c.queueMu.Lock()
	defer c.queueMu.Unlock()
	c.queue = append(c.queue, tmp)
}

func (c *Client) Dequeue() [][]byte {
	c.queueMu.Lock()
	defer c.queueMu.Unlock()
	q := c.queue
	c.queue = c.queue[:0]
	return q
}

func (c *Client) SetConnected() {
	c.connectedMu.Lock()
	defer c.connectedMu.Unlock()
	c.connected = true
}

func (c *Client) IsConnected() bool {
	c.connectedMu.Lock()
	defer c.connectedMu.Unlock()
	return c.connected
}

func (c *Client) WriteLocal(buf []byte) error {
	if _, err := c.localConn.WriteToUDP(buf, c.addr); err != nil {
		return err
	}

	return nil
}

func (c *Client) WriteRemote(buf []byte) error {
	if _, err := c.remoteConn.Write(buf); err != nil {
		return err
	}

	return nil
}
