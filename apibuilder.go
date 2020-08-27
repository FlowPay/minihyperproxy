package minihyperproxy

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/gorilla/mux"
)

func throwError(resp http.ResponseWriter, m *MinihyperProxy, err error, code int, message string) {
	m.ErrorLog.Printf(message)
	resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resp.WriteHeader(code)
	if err := json.NewEncoder(resp).Encode(message); err != nil {
		panic(err)
	}
}

func unmarshalBody(resp http.ResponseWriter, req *http.Request, target interface{}, m *MinihyperProxy) (err error) {
	bodyBytes, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		throwError(resp, m, err, 422, "Error unmarshalling request")
	}
	if err = json.Unmarshal(bodyBytes, target); err != nil {
		throwError(resp, m, err, 422, "Error unmarshalling body")
	}
	return
}

func validateBody(body interface{}, reference interface{}) error {
	referenceFields := reflect.TypeOf(reference)
	bodyFields := reflect.TypeOf(body)

	if referenceFields.NumField() > bodyFields.NumField() {
		return errors.New("Request body invalid")
	}

	for i := 0; i < referenceFields.NumField(); i++ {
		field := referenceFields.Field(i)
		if _, found := bodyFields.FieldByName(field.Name); found {
			return errors.New("Request body invalid")
		}
	}
	return nil
}

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
	var createProxyRequest CreateProxyRequest
	if err := unmarshalBody(resp, req, &createProxyRequest, m); err == nil {
		if createProxyRequest.Name == "" {
			throwError(resp, m, err, 422, "Field Name cannot be empty")
		} else if name, err := m.startProxyServer(createProxyRequest.Name); err == nil {
			response := CreateProxyResponse{Name: createProxyRequest.Name, Port: name}
			json.NewEncoder(resp).Encode(response)
		} else {
			throwError(resp, m, err, 500, err.Error())
		}
	}
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

	serverPort, _ := m.startProxyServer(serverName)
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

func BuildAPI(m *MinihyperProxy) *mux.Router {

	m.InfoLog.Printf("Initializing API")

	httpMux := mux.NewRouter().StrictSlash(true)
	httpMux.HandleFunc("/servers/", buildRoute(m, getServers)).Methods("GET")
	httpMux.HandleFunc("/servers/{name}", buildRoute(m, getServers)).Methods("GET")

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
