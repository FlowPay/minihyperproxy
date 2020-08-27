package minihyperproxy

type ServerInfo *map[string]interface{}

type ProxyInfo struct {
}

type HopInfo struct {
}

type ListServersResponse struct {
	Info []ServerInfo `json:"Info"`
}

type CreateProxyRequest struct {
	Name string `json:"Name"`
}

type CreateProxyResponse struct {
	Name string `json:"Name"`
	Port string `json:"Port"`
}
