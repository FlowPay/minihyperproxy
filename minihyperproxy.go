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
	ServersNameReference       map[string]bool
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

func (m *MinihyperProxy) getFreeServerAndIncrement(referenceEnv string, referenceDefault string, increment bool) string {
	var tempServer int
	if m.latestServer != "" {
		tempServer, _ = strconv.Atoi(m.latestServer)
		tempServer++
	} else {
		tempServer, _ = strconv.Atoi(getEnv(referenceEnv, referenceDefault))
	}
	serverPort := strconv.Itoa(tempServer)

	if increment {
		m.latestServer = serverPort
	}
	return serverPort
}

func (m *MinihyperProxy) AddHop(serverName string, target *url.URL, hop *url.URL) (httpErr *HttpError) {
	if s, ok := m.Servers[serverName]; ok {
		if hopperServer, ok := (*s).(*HopperServer); ok {
			hopperServer.BuildNewOutgoingHop(target, hop)
		} else {
			httpErr = WrongServerTypeError
		}
	} else {
		httpErr = NoServerFoundError
	}
	return
}

func (m *MinihyperProxy) ReceiveHop(serverName string, target *url.URL, hop *url.URL) (httpErr *HttpError) {
	if s, ok := m.Servers[serverName]; ok {
		if hopperServer, ok := (*s).(*HopperServer); ok {
			hopperServer.BuildNewIncomingHop(target, hop)
		} else {
			httpErr = WrongServerTypeError
		}
	} else {
		httpErr = NoServerFoundError
	}
	return
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

func (m *MinihyperProxy) startHopperServer(serverName string, hostname string) (incomingPort, outgoingPort, finalHostname string, httpErr *HttpError) {

	if serverName == "" {
		httpErr = EmptyFieldError
	}

	if _, ok := m.Servers[serverName]; ok {
		httpErr = ServerNameAlreadyExistsError
	}

	if hostname == "" {
		hostname = "localhost"
	}

	if httpErr == nil {
		incomingPort = m.getFreeServerAndIncrement("HOPPER_SERVER_INCOMING", "7053", false)
		outgoingPort = m.getFreeServerAndIncrement("HOPPER_SERVER_OUTGOING", "7054", false)

		fullIncomingServerName := hostname + ":" + incomingPort
		fullOutgoingServerName := hostname + ":" + outgoingPort

		if m.ServersNameReference[fullIncomingServerName] || m.ServersNameReference[fullOutgoingServerName] {
			httpErr = ServerHostnamePortTakenError
		} else {
			finalHostname = hostname
			m.getFreeServerAndIncrement("HOPPER_SERVER_INCOMING", "7053", true)
			m.getFreeServerAndIncrement("HOPPER_SERVER_OUTGOING", "7054", true)
			m.latestHopperServerIncoming = incomingPort
			m.latestHopperServerOutgoing = outgoingPort
			tempServer := Server(NewHopperServer(serverName, hostname, m.latestHopperServerIncoming, m.latestHopperServerOutgoing))
			m.Servers[serverName] = &tempServer
			(*m.Servers[serverName]).Serve()

		}
	}
	return
}

func (m *MinihyperProxy) startProxyServer(serverName string, hostname string) (proxyPort, finalHostname string, httpErr *HttpError) {

	if serverName == "" {
		httpErr = EmptyFieldError
	}

	if _, ok := m.Servers[serverName]; ok {
		httpErr = ServerNameAlreadyExistsError
	}

	if hostname == "" {
		hostname = "localhost"
	}

	if httpErr == nil {
		proxyPort = m.getFreeServerAndIncrement("PROXY_SERVER", "7053", false)
		fullServerName := hostname + ":" + proxyPort

		if m.ServersNameReference[fullServerName] {
			httpErr = ServerHostnamePortTakenError
		} else {
			finalHostname = hostname
			m.getFreeServerAndIncrement("PROXY_SERVER", "7053", true)
			m.latestProxyServer = proxyPort
			tempServer := Server(NewProxyServer(serverName, hostname, m.latestProxyServer))

			m.Servers[serverName] = &tempServer
			(*m.Servers[serverName]).Serve()
		}
	}
	return
}

func (m *MinihyperProxy) addProxyRedirect(serverName string, path *url.URL, target *url.URL) (httpErr *HttpError) {
	if s, ok := m.Servers[serverName]; ok {
		if proxyServer, ok := (*s).(*ProxyServer); ok {
			proxyServer.NewProxy(&url.URL{Path: path.EscapedPath()}, target)
		} else {
			httpErr = WrongServerTypeError
		}
	} else {
		httpErr = NoServerFoundError
	}
	return
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
