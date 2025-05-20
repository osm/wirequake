package qw

import (
	"bytes"
	"crypto/rand"
	"fmt"
)

type Command uint8

const (
	Unknown Command = iota
	Challenge
	Accept
	GetChallenge
	Connect
)

var (
	header       = []byte{0xff, 0xff, 0xff, 0xff}
	accept       = []byte("j")
	challenge    = []byte("c")
	connect      = []byte("connect 28 ")
	getChallenge = []byte("getchallenge\n")
	space        = []byte(" ")
)

func Parse(buf []byte) (Command, []byte) {
	headerLen := len(header)
	if len(buf) < headerLen || !bytes.Equal(buf[:headerLen], header) {
		return Unknown, nil
	}

	payload := buf[headerLen:]
	switch {
	case len(payload) >= len(getChallenge) && bytes.HasPrefix(payload, getChallenge):
		return GetChallenge, nil
	case len(payload) >= len(connect) && bytes.HasPrefix(payload, connect):
		return Connect, nil
	case len(payload) >= len(challenge) && bytes.HasPrefix(payload, challenge):
		id := payload[len(challenge):]
		if len(id) > 0 && id[len(id)-1] == 0x00 {
			id = id[:len(id)-1]
		}
		return Challenge, id
	case len(payload) >= len(accept) && bytes.HasPrefix(payload, accept):
		return Accept, nil
	default:
		return Unknown, nil
	}
}

func GetChallengeBytes() []byte {
	return append(header, getChallenge...)
}

func ConnectBytes(challenge []byte, name, team, targets string) []byte {
	out := []byte{}
	out = append(out, header...)
	out = append(out, connect...)
	out = append(out, randNum(5)...)
	out = append(out, space...)
	out = append(out, challenge...)
	out = append(out, space...)
	out = append(out, []byte(fmt.Sprintf(`"\team\%s\name\%s\prx\%s"\n`, team, name, targets))...)
	return out
}

func ChallengeBytes() []byte {
	out := []byte{}
	out = append(out, header...)
	out = append(out, challenge...)
	out = append(out, randNum(8)...)
	return out
}

func AcceptBytes() []byte {
	return append(header, accept...)
}

func randNum(n int) []byte {
	num := make([]byte, n)

	for i := range num {
		var b [1]byte
		rand.Read(b[:])
		num[i] = '0' + (b[0] % 10)
	}

	return num
}
