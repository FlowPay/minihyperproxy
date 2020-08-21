package minihyperproxy

import (
	"log"
	"net/url"
	"os"
	"strconv"
)

type MinihyperProxy struct {
	ErrorLog                   *log.Logger
	WarnLog                    *log.Logger
	InfoLog                    *log.Logger
	latestHopperServerIncoming string
	latestHopperServerOutgoing string
	Servers                    map[string]*Server
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

func (m *MinihyperProxy) startHopperServer(serverName string) (incomingPort, outgoingPort string) {
	if m.latestHopperServerIncoming != "" {
		incTemp, err := strconv.Atoi(m.latestHopperServerIncoming)
		if err != nil {
			m.ErrorLog.Print(err)
			return "", ""
		}
		outTemp, err := strconv.Atoi(m.latestHopperServerOutgoing)
		if err != nil {
			m.ErrorLog.Print(err)
			return "", ""
		}
		incTemp += 2
		outTemp += 2
		m.latestHopperServerIncoming = strconv.Itoa(incTemp)
		m.latestHopperServerOutgoing = strconv.Itoa(outTemp)
	} else {
		m.latestHopperServerIncoming = getEnv("HOPPER_SERVER_INCOMING", "7052")
		m.latestHopperServerOutgoing = getEnv("HOPPER_SERVER_OUTGOING", "7053")
	}
	tempServer := Server(NewHopperServer(serverName, m.latestHopperServerIncoming, m.latestHopperServerOutgoing))
	m.Servers[serverName] = &tempServer
	(*m.Servers[serverName]).Serve()
	return m.latestHopperServerIncoming, m.latestHopperServerOutgoing
}

func (m *MinihyperProxy) AddHop(serverName string, target *url.URL, hop *url.URL) {
	if s, ok := m.Servers[serverName]; ok {
		if hopperServer, ok := (*s).(*HopperServer); ok {
			hopperServer.BuildNewOutgoingHop(target, hop)
		} else {
			m.ErrorLog.Printf("Server %s is not of the right type", serverName)
		}
	} else {
		m.ErrorLog.Printf("Server %s doesn't exist", serverName)
	}
}

func (m *MinihyperProxy) ReceiveHop(serverName string, target *url.URL, hop *url.URL) {
	if s, ok := m.Servers[serverName]; ok {
		if hopperServer, ok := (*s).(*HopperServer); ok {
			hopperServer.BuildNewIncomingHop(target, hop)
		} else {
			m.ErrorLog.Printf("Server %s is not of the right type", serverName)
		}
	} else {
		m.ErrorLog.Printf("Server %s doesn't exist", serverName)
	}
}

func (m *MinihyperProxy) GetOutgoingHops(serverName string) map[string]*url.URL {
	if s, ok := m.Servers[serverName]; ok {
		if hopperServer, ok := (*s).(*HopperServer); ok {
			return hopperServer.getOutgoingHops()
		} else {
			m.ErrorLog.Printf("Server %s is not of the right type", serverName)
		}
	} else {
		m.ErrorLog.Printf("Server %s doesn't exist", serverName)
	}
	return nil
}

func (m *MinihyperProxy) GetIncomingHops(serverName string) map[string]*url.URL {
	if s, ok := m.Servers[serverName]; ok {
		if hopperServer, ok := (*s).(*HopperServer); ok {
			return hopperServer.getIncomingHops()
		} else {
			m.ErrorLog.Printf("Server %s is not of the right type", serverName)
		}
	} else {
		m.ErrorLog.Printf("Server %s doesn't exist", serverName)
	}
	return nil
}

func (m *MinihyperProxy) GetProxyMap(serverName string) map[string]string {
	if s, ok := m.Servers[serverName]; ok {
		if proxyServer, ok := (*s).(*ProxyServer); ok {
			return proxyServer.getProxyMap()
		} else {
			m.ErrorLog.Printf("Server %s is not of the right type", serverName)
		}
	} else {
		m.ErrorLog.Printf("Server %s doesn't exist", serverName)
	}
	return nil
}

func (m *MinihyperProxy) startProxyServer(serverName string) {
	tempServer := Server(NewProxyServer(serverName, getEnv("PROXY_SERVER", "7052")))
	m.Servers[serverName] = &tempServer
	(*m.Servers[serverName]).Serve()
}

func (m *MinihyperProxy) addProxyRedirect(serverName string, path *url.URL, target *url.URL) {
	if s, ok := m.Servers[serverName]; ok {
		if proxyServer, ok := (*s).(*ProxyServer); ok {
			proxyServer.NewProxy(&url.URL{Path: path.EscapedPath()}, target)
		} else {
			m.ErrorLog.Printf("Server %s is not of the right type", serverName)
		}
	} else {
		m.ErrorLog.Printf("Server %s doesn't exist", serverName)
	}
}

func (m *MinihyperProxy) stopServer(serverName string) {
	if s, ok := m.Servers[serverName]; ok {
		m.InfoLog.Printf("Stopping %s", serverName)
		if _, ok := (*s).(*HopperServer); ok {

		}
		(*s).Stop()
	}
}

func (m *MinihyperProxy) importConfig() {}
