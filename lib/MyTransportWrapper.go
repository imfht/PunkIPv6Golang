package lib

import (
  "context"
  "net"
  "net/http"
  "time"
)

type MyTransportWrapper struct {
  http.RoundTripper
  transport *http.Transport
  IPv6Only  bool
  IPv4Only  bool
}

func (t MyTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
  transport := http.Transport{DialContext: t.dial}
  return transport.RoundTrip(req)
}
func (t MyTransportWrapper) getNetworkString() string {
  if t.IPv6Only {
    return "tcp6"
  } else if t.IPv4Only {
    return "tcp4"
  } else {
    return "tcp"
  }
}
func (t *MyTransportWrapper) dial(ctx context.Context, network, addr string) (net.Conn, error) {
  if t.transport == nil {
    dialContext := (&net.Dialer{
      Timeout:   30 * time.Second,
      KeepAlive: 30 * time.Second,
      DualStack: true,
    }).DialContext
    return dialContext(ctx, t.getNetworkString(), addr)
  }
  return t.transport.DialContext(ctx, t.getNetworkString(), addr)
}
