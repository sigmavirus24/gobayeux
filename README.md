# gobayeux

[![PkgGoDev](https://pkg.go.dev/badge/github.com/L11R/gobayeux)](https://pkg.go.dev/github.com/L11R/gobayeux)
[![Test](https://github.com/L11R/gobayeux/workflows/Test/badge.svg)](https://github.com/L11R/gobayeux/actions?query=workflow:Test)
[![Build](https://github.com/L11R/gobayeux/workflows/Build/badge.svg)](https://github.com/L11R/gobayeux/actions?query=workflow:Build)
[![Lint](https://github.com/L11R/gobayeux/workflows/Lint/badge.svg)](https://github.com/L11R/gobayeux/actions?query=workflow:Lint)

Bayeux protocol library compatible with [CometD](https://cometd.org/)
and [Faye](https://faye.jcoglan.com/) servers.

### Documentation

- [API Reference](https://pkg.go.dev/github.com/L11R/gobayeux)
- [Protocol specification](https://docs.cometd.org/current/reference/#_bayeux)

### Installation

```bash
go get github.com/L11R/gobayeux
```

### Status

Library provides a basic set of features to start getting notification over `long-polling` transport.

### Protocol compliance

- [x] Handshake
- [x] Connect/Disconnect
- [x] Subscribe/Unsubscribe
- [ ] Publish and Delivery event messages
- [x] The long-polling transport
- [ ] The callback-polling transport
- [ ] The websocket transport

### Authors

- @sigmavirus24
- @L11R

### License

Apache 2.0