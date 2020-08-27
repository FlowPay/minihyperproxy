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

func throwError(resp http.ResponseWriter, m *MinihyperProxy, httpErr *HttpError) {
	if httpErr == nil {
		return
	}

	m.ErrorLog.Printf(httpErr.Error())
	resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resp.WriteHeader(httpErr.code)
	if err := json.NewEncoder(resp).Encode(httpErr); err != nil {
		panic(err)
	}
}

func unmarshalBody(resp http.ResponseWriter, req *http.Request, target interface{}, m *MinihyperProxy) (httpErr *HttpError) {
	bodyBytes, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		httpErr = RequestUnmarshallError
	}
	if err = json.Unmarshal(bodyBytes, target); err != nil {
		httpErr = BodyUnmarshallError
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
	json.NewEncoder(resp).Encode(ListServersResponse{Info: servers})
}

func getServer(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	var getServerRequest GetServerRequest
	var httpErr *HttpError
	var serverInfo ServerInfo

	defer throwError(resp, m, httpErr)

	httpErr = unmarshalBody(resp, req, &getServerRequest, m)

	if httpErr == nil {
		serverInfo, httpErr = m.GetServerInfo(getServerRequest.Name)
		if httpErr == nil {
			json.NewEncoder(resp).Encode(ListServersResponse{Info: []ServerInfo{serverInfo}})
		}
	}
}

func getProxies(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	serversInfo := m.GetProxiesInfo()
	json.NewEncoder(resp).Encode(ListServersResponse{Info: serversInfo})
}

func getProxy(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	var getServerRequest GetServerRequest
	var httpErr *HttpError
	var serverInfo ServerInfo

	defer throwError(resp, m, httpErr)

	httpErr = unmarshalBody(resp, req, &getServerRequest, m)

	if httpErr == nil {
		serverInfo, httpErr = m.GetProxyInfo(getServerRequest.Name)
		if httpErr == nil {
			json.NewEncoder(resp).Encode(ListServersResponse{Info: []ServerInfo{serverInfo}})
		}
	}
}

func getProxyMap(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	var getServerRequest GetServerRequest
	var httpErr *HttpError
	var proxyMap map[string]string

	defer throwError(resp, m, httpErr)

	httpErr = unmarshalBody(resp, req, &getServerRequest, m)

	if httpErr == nil {
		proxyMap, httpErr = m.GetProxyMap(getServerRequest.Name)
		if httpErr == nil {
			json.NewEncoder(resp).Encode(ProxyMapResponse{ProxyMap: proxyMap})
		}
	}

}

func createProxy(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	var createProxyRequest CreateProxyRequest
	var httpErr *HttpError
	var name string

	defer throwError(resp, m, httpErr)

	httpErr = unmarshalBody(resp, req, &createProxyRequest, m)

	if httpErr == nil {
		name, httpErr = m.startProxyServer(createProxyRequest.Name)
		if httpErr == nil {
			response := CreateProxyResponse{Name: createProxyRequest.Name, Port: name}
			json.NewEncoder(resp).Encode(response)
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

func getHoppers(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	serversInfo := m.GetHoppersInfo()
	json.NewEncoder(resp).Encode(ListServersResponse{Info: serversInfo})
}

func getHopper(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	var getServerRequest GetServerRequest
	var httpErr *HttpError
	var serverInfo ServerInfo

	defer throwError(resp, m, httpErr)

	httpErr = unmarshalBody(resp, req, &getServerRequest, m)

	if httpErr == nil {
		serverInfo, httpErr = m.GetHopperInfo(getServerRequest.Name)
		if httpErr == nil {
			json.NewEncoder(resp).Encode(ListServersResponse{Info: []ServerInfo{serverInfo}})
		}
	}
}

func createHopper(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	var createHopperRequest CreateHopperRequest
	var httpErr *HttpError
	var incomingPort, outgoingPort string

	defer throwError(resp, m, httpErr)

	httpErr = unmarshalBody(resp, req, &createHopperRequest, m)

	if httpErr == nil {
		incomingPort, outgoingPort, httpErr = m.startHopperServer(createHopperRequest.Name)
		if httpErr == nil {
			json.NewEncoder(resp).Encode(CreateHopperResponse{Name: createHopperRequest.Name, IncomingPort: incomingPort, OutgoingPort: outgoingPort})
		}
	}
}

func getHops(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	var getHopsRequest GetHopsRequest
	var httpErr *HttpError
	var incomingHops, outgoingHops map[string]*url.URL

	defer throwError(resp, m, httpErr)

	httpErr = unmarshalBody(resp, req, &getHopsRequest, m)

	if httpErr == nil {
		incomingHops, httpErr = m.GetIncomingHops(getHopsRequest.Name)
		outgoingHops, httpErr = m.GetOutgoingHops(getHopsRequest.Name)
		if httpErr == nil {
			json.NewEncoder(resp).Encode(GetHopsResponse{IncomingHops: incomingHops, OutgoingHops: outgoingHops})
		}
	}
}

func getIncomingHops(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	var getHopsRequest GetHopsRequest
	var httpErr *HttpError
	var hops map[string]*url.URL

	defer throwError(resp, m, httpErr)

	httpErr = unmarshalBody(resp, req, &getHopsRequest, m)

	if httpErr == nil {
		hops, httpErr = m.GetIncomingHops(getHopsRequest.Name)
		if httpErr == nil {
			json.NewEncoder(resp).Encode(GetIncomingHopsResponse{IncomingHops: hops})
		}
	}
}

func getOutgoingHops(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy) {
	var getHopsRequest GetHopsRequest
	var httpErr *HttpError
	var hops map[string]*url.URL

	defer throwError(resp, m, httpErr)

	httpErr = unmarshalBody(resp, req, &getHopsRequest, m)

	if httpErr == nil {
		hops, httpErr = m.GetOutgoingHops(getHopsRequest.Name)
		if httpErr == nil {
			response := GetOutgoingHopsResponse{OutgoingHops: hops}
			json.NewEncoder(resp).Encode(response)
		}
	}
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
	httpMux.HandleFunc("/servers", buildRoute(m, getServers)).Methods("GET")
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
