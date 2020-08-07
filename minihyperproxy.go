package minihyperproxy

import (
	"log"
	"net/url"
	"os"
)

type MinihyperProxy struct {
	ErrorLog *log.Logger
	WarnLog  *log.Logger
	InfoLog  *log.Logger

	Servers map[string]*Server
}

func NewMinihyperProxy() (m *MinihyperProxy) {
	m = &MinihyperProxy{ErrorLog: log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		WarnLog: log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		InfoLog: log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		Servers: make(map[string]*Server)}
	return
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (m *MinihyperProxy) startHopperServer(serverName string, port string) {}

func (m *MinihyperProxy) stopHopperServer(serverName string, port string) {}

func (m *MinihyperProxy) startProxyServer(serverName string) {
	tempServer := Server(NewProxyServer(serverName, getEnv("PROXY_SERVER", "7052"), m.InfoLog, m.WarnLog, m.ErrorLog))
	m.Servers[serverName] = &tempServer
	(*m.Servers[serverName]).Serve()
}

func (m *MinihyperProxy) addProxyRedirect(serverName string, path string, target *url.URL) {
	if s, ok := m.Servers[serverName]; ok {
		if proxyServer, ok := (*s).(*ProxyServer); ok {
			proxyServer.NewProxy(&url.URL{RawPath: path}, target)
		} else {
			m.ErrorLog.Printf("Server %s is not of the right type", serverName)
		}
	} else {
		m.ErrorLog.Printf("Server %s doesn't exist", serverName)
	}
}

func (m *MinihyperProxy) stopProxyServer(serverName string) {
	if _, ok := m.Servers[serverName]; ok {
		(*m.Servers[serverName]).Stop()
	}
}

func (m *MinihyperProxy) importConfig() {}
