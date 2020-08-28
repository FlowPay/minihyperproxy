package minihyperproxy

import "net/url"

type ServerInfo *map[string]interface{}

type ProxyInfo struct {
}

type HopInfo struct {
}

type ListServersResponse struct {
	Info []ServerInfo `json:"Info"`
}

type ProxyMapResponse struct {
	ProxyMap map[string]string `json:"ProxyMap"`
}
type GetServerRequest struct {
	Name string `json:"Name"`
}
type CreateProxyRequest struct {
	Name     string `json:"Name"`
	Hostname string `json:"Hostname"`
}

type CreateProxyResponse struct {
	Name     string `json:"Name"`
	Hostname string `json:"Hostname"`
	Port     string `json:"Port"`
}

type CreateRouteRequest struct {
	Name   string `json:"Name"`
	Route  string `json:"Route"`
	Target string `json:"Target"`
}

type CreateRouteResponse CreateRouteRequest

type CreateHopperRequest struct {
	Name     string `json:"Name"`
	Hostname string `json:"Hostname"`
}

type CreateHopperResponse struct {
	Name         string `json:"Name"`
	Hostname     string `json:"Hostname"`
	IncomingPort string `json:"IncomingPort"`
	OutgoingPort string `json:"OutgoingPort"`
}

type CreateIncomingHopRequest struct {
	Name   string `json:"Name"`
	Route  string `json:"Route"`
	Target string `json:"Target"`
}

type CreateOutgoingHopRequest struct {
	Name   string `json:"Name"`
	Route  string `json:"Route"`
	Target string `json:"Target"`
}

type GetHopsRequest struct {
	Name string `json:"Name"`
}

type GetIncomingHopsResponse struct {
	IncomingHops map[string]*url.URL `json:"IncomingHops"`
}

type GetOutgoingHopsResponse struct {
	OutgoingHops map[string]*url.URL `json:"OutgoingHops"`
}

type GetHopsResponse struct {
	IncomingHops map[string]*url.URL `json:"IncomingHops"`
	OutgoingHops map[string]*url.URL `json:"OutgoingHops"`
}
