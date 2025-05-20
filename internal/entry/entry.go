package entry

import (
	"errors"
	"strings"
	"time"

	"github.com/osm/wirequake/internal/proxy"
	"github.com/osm/wirequake/internal/qw"
)

type Entry struct {
	name    string
	team    string
	targets []string
}

func New(name, team string, targets []string) *Entry {
	return &Entry{name: name, team: team, targets: targets}
}

func (e *Entry) FromClient(client *proxy.Client, buf []byte) error {
	if client.IsConnected() {
		return client.WriteRemote(buf)
	}

	// When not connected, enqueue the packets so they can be retransmitted
	// as soon as the full tunnel connection is established.
	client.Enqueue(buf)

	return client.WriteRemote(qw.GetChallengeBytes())
}

func (e *Entry) FromRemote(client *proxy.Client, buf []byte) error {
	if client.IsConnected() {
		return client.WriteLocal(buf)
	}

	cmd, payload := qw.Parse(buf)
	switch cmd {
	case qw.Challenge:
		if payload == nil {
			return errors.New("failed to extract challenge id")
		}
		return client.WriteRemote(qw.ConnectBytes(payload,
			e.name, e.team, strings.Join(e.targets, "@")))
	case qw.Accept:
		// Give the proxies in the chain a while to finish establishing
		// the full tunnel before we dequeue our initial packets.
		time.Sleep(100 * time.Millisecond)

		for _, p := range client.Dequeue() {
			client.WriteRemote(p)
		}

		client.SetConnected()
	}

	return nil
}
