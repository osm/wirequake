# WireQuake

Hijack QWFWD for general-purpose traffic tunneling.

## Example

WireQuake is a proxy that can be used, for example, as an intermediary between
WireGuard and QWFWD proxies. Its purpose is to relay arbitrary traffic through
one or more QWFWD proxies.

In the following example, we’ll relay WireGuard traffic through a single QWFWD
proxy.

`[WireGuard] -> [WireQuake] -> [QWFWD #1] -> [WireQuake] -> [WireGuard]`

The WireQuake entry establishes the connection to the first QWFWD in the
chain, while the WireQuake exit responds to the handshake request from the
last QWFWD to complete the tunnel.

In this example, two WireQuake instances are started to create a tunnel for
WireGuard traffic through a QWFWD proxy chain.

Together, these two WireQuake proxies enable WireGuard traffic to be tunneled
through the QWFWD proxies, providing an additional relay layer before reaching
the actual WireGuard endpoint.

```
$ wirequake -role entry \
	-listen-addr 10.0.0.1:41820 \
	-target-addrs 10.0.0.2:30000@10.0.0.3:41820
```

This command starts the WireQuake entry proxy, which listens on `10.0.0.1:41820`.
It is configured to forward traffic to a QWFWD proxy at 10.0.0.2:30000, and
from there, continue to the WireQuake exit proxy at 10.0.0.3:41820. In effect,
the entry proxy initiates the tunnel by sending encapsulated WireGuard traffic
into the QWFWD chain.


```sh
$ wirequake -role exit \
	-listen-addr 10.0.0.3:41820 \
	-target-addrs 127.0.0.1:51820
```

This command starts the WireQuake exit proxy, which listens on `10.0.0.3:41820`.
It receives traffic from the last QWFWD proxy in the chain, decapsulates it,
and forwards the original WireGuard traffic to the WireGuard endpoint at
`127.0.0.1:51820`.

## QWFWD

I recommend using the `whitelist` feature of QWFWD to ensure that your proxy
only forwards traffic to servers you explicitly want to allow.

## Disclaimer

This is just a proof of concept to demonstrate how a QWFWD proxy could be
abused. Please don’t use it for anything malicious or in real setups.

## Links

- [QWFWD](https://github.com/QW-Group/qwfwd)
