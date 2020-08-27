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

func (m *MinihyperProxy) GetOutgoingHops(serverName string) (hops map[string]*url.URL, httpErr *HttpError) {
	if s, ok := m.Servers[serverName]; ok {
		if hopperServer, ok := (*s).(*HopperServer); ok {
			hops = hopperServer.getOutgoingHops()
		} else {
			httpErr = WrongServerTypeError
		}
	} else {
		httpErr = NoServerFoundError
	}
	return
}

func (m *MinihyperProxy) GetIncomingHops(serverName string) (hops map[string]*url.URL, httpErr *HttpError) {
	if s, ok := m.Servers[serverName]; ok {
		if hopperServer, ok := (*s).(*HopperServer); ok {
			hops = hopperServer.getIncomingHops()
		} else {
			httpErr = WrongServerTypeError
		}
	} else {
		httpErr = NoServerFoundError
	}
	return
}

func (m *MinihyperProxy) GetProxyMap(serverName string) (proxyMap map[string]string, httpErr *HttpError) {
	if s, ok := m.Servers[serverName]; ok {
		if proxyServer, ok := (*s).(*ProxyServer); ok {
			proxyMap = proxyServer.getProxyMap()
		} else {
			httpErr = WrongServerTypeError
		}
	} else {
		httpErr = NoServerFoundError
	}
	return
}

func (m *MinihyperProxy) startHopperServer(serverName string) (incomingPort, outgoingPort string, httpErr *HttpError) {
	if serverName == "" {
		httpErr = EmptyFieldError
	}

	if _, ok := m.Servers[serverName]; ok {
		httpErr = ServerAlreadyExistsError
	}

	if httpErr != nil {
		m.latestHopperServerIncoming = m.getFreeServerAndIncrement("HOPPER_SERVER_INCOMING", "7053")
		m.latestHopperServerOutgoing = m.getFreeServerAndIncrement("HOPPER_SERVER_OUTGOING", "7054")
		tempServer := Server(NewHopperServer(serverName, m.latestHopperServerIncoming, m.latestHopperServerOutgoing))
		m.Servers[serverName] = &tempServer
		(*m.Servers[serverName]).Serve()

		incomingPort = m.latestHopperServerIncoming
		outgoingPort = m.latestHopperServerOutgoing
	}
	return
}

func (m *MinihyperProxy) startProxyServer(serverName string) (proxyPort string, httpErr *HttpError) {

	if serverName == "" {
		httpErr = EmptyFieldError
	}

	if _, ok := m.Servers[serverName]; ok {
		httpErr = ServerAlreadyExistsError
	}

	if httpErr != nil {
		m.latestProxyServer = m.getFreeServerAndIncrement("PROXY_SERVER", "7053")
		tempServer := Server(NewProxyServer(serverName, m.latestProxyServer))

		m.Servers[serverName] = &tempServer
		(*m.Servers[serverName]).Serve()
		proxyPort = m.latestProxyServer
	}
	return
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

func (m *MinihyperProxy) GetServersInfo() (serversInfo []ServerInfo) {
	for _, s := range m.Servers {
		newServerInfo := (*s).Info()
		serversInfo = append(serversInfo, newServerInfo)
	}
	return
}

func (m *MinihyperProxy) GetProxiesInfo() (serversInfo []ServerInfo) {
	for _, s := range m.Servers {
		newServerInfo := (*s).Info()
		if (*newServerInfo)["Type"] != "Proxy" {
			continue
		}
		serversInfo = append(serversInfo, newServerInfo)
	}
	return
}

func (m *MinihyperProxy) GetProxyInfo(Name string) (serverInfo ServerInfo, httpErr *HttpError) {

	if s, ok := m.Servers[Name]; ok {
		newServerInfo := (*s).Info()
		if (*newServerInfo)["Type"] == "Proxy" {
			serverInfo = (*s).Info()
		} else {
			httpErr = WrongServerTypeError
		}
	} else {
		httpErr = NoServerFoundError
	}
	return
}

func (m *MinihyperProxy) GetHoppersInfo() (serversInfo []ServerInfo) {
	for _, s := range m.Servers {
		newServerInfo := (*s).Info()
		if (*newServerInfo)["Type"] != "Hopper" {
			continue
		}
		serversInfo = append(serversInfo, newServerInfo)
	}
	return
}

func (m *MinihyperProxy) GetHopperInfo(Name string) (serverInfo ServerInfo, httpErr *HttpError) {
	if s, ok := m.Servers[Name]; ok {
		newServerInfo := (*s).Info()
		if (*newServerInfo)["Type"] == "Hopper" {
			serverInfo = (*s).Info()
		} else {
			httpErr = WrongServerTypeError
		}
	} else {
		httpErr = NoServerFoundError
	}
	return
}

func (m *MinihyperProxy) GetServerInfo(Name string) (serverInfo ServerInfo, httpErr *HttpError) {
	if s, ok := m.Servers[Name]; ok {
		serverInfo = (*s).Info()
	} else {
		httpErr = NoServerFoundError
	}
	return
}

func (m *MinihyperProxy) importConfig() {}
