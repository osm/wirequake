package exit

import (
	"github.com/osm/wirequake/internal/proxy"
	"github.com/osm/wirequake/internal/qw"
)

type Exit struct{}

func New() *Exit {
	return &Exit{}
}

func (e *Exit) FromClient(client *proxy.Client, buf []byte) error {
	if client.IsConnected() {
		return client.WriteRemote(buf)
	}

	cmd, _ := qw.Parse(buf)
	switch cmd {
	case qw.GetChallenge:
		return client.WriteLocal(qw.ChallengeBytes())
	case qw.Connect:
		if err := client.WriteLocal(qw.AcceptBytes()); err != nil {
			return err
		}
		client.SetConnected()
	}

	return nil
}

func (e *Exit) FromRemote(client *proxy.Client, buf []byte) error {
	return client.WriteLocal(buf)
}
