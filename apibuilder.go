package minihyperproxy

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

func buildRoute(m *MinihyperProxy, target func(http.ResponseWriter, *http.Request, *MinihyperProxy)) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) { target(resp, req, m) }
}

func getServers(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	servers := m.GetServersInfo()
	json.NewEncoder(resp).Encode(servers)
}

func getServer(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	vars := mux.Vars(req)
	serverName := vars["Name"]

	server := m.GetServerInfo(serverName)
	json.NewEncoder(resp).Encode(server)
}

// Da filtrare i proxy
func getProxies(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	servers := m.GetServersInfo()
	json.NewEncoder(resp).Encode(servers)
}

func getProxy(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	vars := mux.Vars(req)
	serverName := vars["Name"]

	server := m.GetServerInfo(serverName)
	json.NewEncoder(resp).Encode(server)
}

func getProxyMap(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	vars := mux.Vars(req)
	serverName := vars["Name"]

	server := m.GetProxyMap(serverName)
	json.NewEncoder(resp).Encode(server)
}

func createProxy(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	vars := mux.Vars(req)

	serverName := vars["Name"]
	_ = vars["Port"]

	serverPort := m.startProxyServer(serverName)
	json.NewEncoder(resp).Encode(serverPort)
}

func createRoute(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	vars := mux.Vars(req)

	serverName := vars["Name"]
	route := vars["Route"]
	target := vars["Target"]

	routeURL, _ := url.Parse(route)
	targetURL, _ := url.Parse(target)

	m.addProxyRedirect(serverName, routeURL, targetURL)

	resp.Write([]byte{})
}

// Da filtrare gli hopper
func getHoppers(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	servers := m.GetServersInfo()
	json.NewEncoder(resp).Encode(servers)
}

func getHopper(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	vars := mux.Vars(req)
	serverName := vars["Name"]

	server := m.GetServerInfo(serverName)
	json.NewEncoder(resp).Encode(server)
}

func createHopper(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	vars := mux.Vars(req)

	serverName := vars["Name"]
	_ = vars["Port"]

	serverPort := m.startProxyServer(serverName)
	json.NewEncoder(resp).Encode(serverPort)
}

func getHops(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {

}

func getIncomingHops(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	vars := mux.Vars(req)

	serverName := vars["Name"]

	serverPort := m.GetIncomingHops(serverName)
	json.NewEncoder(resp).Encode(serverPort)
}

func getOutgoingHops(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	vars := mux.Vars(req)

	serverName := vars["Name"]

	serverPort := m.GetOutgoingHops(serverName)
	json.NewEncoder(resp).Encode(serverPort)
}

func createIncomingHop(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	vars := mux.Vars(req)

	serverName := vars["Name"]
	route := vars["Route"]
	target := vars["Target"]

	routeURL, _ := url.Parse(route)
	targetURL, _ := url.Parse(target)

	m.ReceiveHop(serverName, routeURL, targetURL)

	resp.Write([]byte{})
}

func createOutgoingHop(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	vars := mux.Vars(req)

	serverName := vars["Name"]
	route := vars["Route"]
	target := vars["Target"]

	routeURL, _ := url.Parse(route)
	targetURL, _ := url.Parse(target)

	m.AddHop(serverName, routeURL, targetURL)

	resp.Write([]byte{})
}

func buildAPI(m *MinihyperProxy) *mux.Router {

	httpMux := mux.NewRouter().StrictSlash(true)
	httpMux.HandleFunc("/servers/get", buildRoute(m, getServers)).Methods("GET")
	httpMux.HandleFunc("/servers/get/{name}", buildRoute(m, getServers)).Methods("GET")

	httpMux.HandleFunc("/proxy", buildRoute(m, getProxies)).Methods("GET")
	httpMux.HandleFunc("/proxy", buildRoute(m, createProxy)).Methods("POST")
	httpMux.HandleFunc("/proxy/{name}", buildRoute(m, getProxy)).Methods("GET")

	httpMux.HandleFunc("/proxy/{name}/route", buildRoute(m, getProxyMap)).Methods("GET")
	httpMux.HandleFunc("/proxy/{name}/route", buildRoute(m, createRoute)).Methods("POST")

	httpMux.HandleFunc("/hopper", buildRoute(m, getHoppers)).Methods("GET")
	httpMux.HandleFunc("/hopper", buildRoute(m, createHopper)).Methods("POST")
	httpMux.HandleFunc("/hopper/{name}", buildRoute(m, getHopper)).Methods("GET")

	httpMux.HandleFunc("/hopper/{name}/hop/out", buildRoute(m, getOutgoingHops)).Methods("GET")
	httpMux.HandleFunc("/hopper/{name}/hop/in", buildRoute(m, getIncomingHops)).Methods("GET")
	httpMux.HandleFunc("/hopper/{name}/hop/out", buildRoute(m, createOutgoingHop)).Methods("POST")
	httpMux.HandleFunc("/hopper/{name}/hop/in", buildRoute(m, createIncomingHop)).Methods("POST")

	return httpMux
}

func serveAPI(port string, httpMux *mux.Router) {
	http.ListenAndServe(":"+port, httpMux)
}
