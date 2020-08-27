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
	latestProxyServer          string
	latestServer               string
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

func (m *MinihyperProxy) getFreeServerAndIncrement(referenceEnv string, referenceDefault string) string {
	var tempServer int
	if m.latestServer != "" {
		tempServer, _ = strconv.Atoi(m.latestServer)
		tempServer++
	} else {
		tempServer, _ = strconv.Atoi(getEnv(referenceEnv, referenceDefault))
	}
	serverPort := strconv.Itoa(tempServer)
	m.latestServer = serverPort
	return serverPort
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

func (m *MinihyperProxy) startHopperServer(serverName string) (incomingPort, outgoingPort string) {
	m.latestHopperServerIncoming = m.getFreeServerAndIncrement("HOPPER_SERVER_INCOMING", "7053")
	m.latestHopperServerOutgoing = m.getFreeServerAndIncrement("HOPPER_SERVER_OUTGOING", "7054")
	tempServer := Server(NewHopperServer(serverName, m.latestHopperServerIncoming, m.latestHopperServerOutgoing))
	m.Servers[serverName] = &tempServer
	(*m.Servers[serverName]).Serve()
	return m.latestHopperServerIncoming, m.latestHopperServerOutgoing
}

func (m *MinihyperProxy) startProxyServer(serverName string) (proxyPort string, err *HttpError) {

	if serverName == "" {
		return "", EmptyFieldError
	}

	if _, ok := m.Servers[serverName]; ok {
		return "", ServerAlreadyExistsError
	}

	m.latestProxyServer = m.getFreeServerAndIncrement("PROXY_SERVER", "7053")
	tempServer := Server(NewProxyServer(serverName, m.latestProxyServer))

	m.Servers[serverName] = &tempServer
	(*m.Servers[serverName]).Serve()
	return m.latestProxyServer, nil
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
		(*s).Stop()
	}
}

func (m *MinihyperProxy) GetServersInfo() []ServerInfo {
	var ret []ServerInfo
	for _, s := range m.Servers {
		newServerInfo := (*s).Info()
		ret = append(ret, newServerInfo)
	}
	return ret
}

func (m *MinihyperProxy) GetServerInfo(Name string) ServerInfo {
	var ret ServerInfo
	if s, ok := m.Servers[Name]; ok {
		ret = (*s).Info()
	}
	return ret
}

func (m *MinihyperProxy) importConfig() {}
