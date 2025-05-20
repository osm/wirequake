package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/osm/wirequake/internal/entry"
	"github.com/osm/wirequake/internal/exit"
	"github.com/osm/wirequake/internal/proxy"
)

const (
	name = "wire"
	team = "quake"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging for verbose output")
	listenAddr := flag.String("listen-addr", "localhost:41820", "Local address to listen on")
	role := flag.String("role", "", "Node role in the forwarding chain: 'entry' or 'exit'")
	targetAddrsRaw := flag.String("target-addrs", "", "Target addresses (ip1:port1@ip2:port2)")
	flag.Parse()

	if *listenAddr == "" {
		exitf("Missing required flag: -listen-addr. Use -h for usage\n")
	}

	if *role == "" {
		exitf("Missing required flag: -role. Use -h for usage\n")
	}
	if *role != "entry" && *role != "exit" {
		exitf("Invalid value for -role: must be either 'entry' or 'exit'\n")
	}

	if *targetAddrsRaw == "" {
		exitf("Missing required flag: -target-addrs\n")
	}
	targetAddrs, err := parseTargetAddrs(*targetAddrsRaw, *role)
	if err != nil {
		exitf("Failed to parse target addresses: %v\n", err)
	}

	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))

	var router proxy.Router
	if *role == "entry" {
		router = entry.New(name, team, targetAddrs[1:])
	} else {
		router = exit.New()
	}

	logger.Info(fmt.Sprintf("WireQuake %s", *role),
		"listen-addr", *listenAddr,
		"target-addrs", *targetAddrsRaw)
	proxy := proxy.New(logger, router)
	if err := proxy.ListenAndServe(*listenAddr, targetAddrs[0]); err != nil {
		exitf("Failed to start listener: %v\n", err)
	}
}

func parseTargetAddrs(targetAddrs string, role string) ([]string, error) {
	parts := strings.Split(targetAddrs, "@")
	if role == "entry" && len(parts) < 2 {
		return nil, errors.New("too few target addresses specified")
	}

	addrs := make([]string, len(parts))

	for i, p := range parts {
		a := strings.TrimSpace(p)
		if !strings.Contains(a, ":") {
			return nil, fmt.Errorf("address %q does not contain a port", a)
		}

		addrs[i] = a
	}

	return addrs, nil
}

func exitf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
