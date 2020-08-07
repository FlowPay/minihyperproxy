package minihyperproxy

import (
	"net/url"
	"os"
	"testing"
	"time"
)

func TestGenerale(t *testing.T) {
	os.Setenv("PROXY_SERVER", "7052")
	m := NewMinihyperProxy()
	m.startProxyServer("prova")
	time.Sleep(time.Second * 5)
	m.addProxyRedirect("prova", "/prova1", &url.URL{Host: "localhost:3000", RawPath: "/prova1"})
	m.stopProxyServer("prova")
	t.Logf("Sono qui")
}
