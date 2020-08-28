package minihyperproxy

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/gorilla/mux"
)

//func throwError(resp http.ResponseWriter, m *MinihyperProxy, httpErr *HttpError) {
//	if httpErr == nil {
//		return
//	}
//
//	m.ErrorLog.Printf(httpErr.Error())
//	resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
//	resp.WriteHeader(httpErr.code)
//	if err := json.NewEncoder(resp).Encode(httpErr); err != nil {
//		panic(err)
//	}
//}

func throwError(resp http.ResponseWriter, m *MinihyperProxy, httpErr **HttpError) {
	if *httpErr == nil {
		return
	}

	m.ErrorLog.Printf((*httpErr).Error())
	resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resp.WriteHeader((*httpErr).code)
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

func validateBody(body interface{}, reference interface{}) (httpErr *HttpError) {
	referenceFields := reflect.TypeOf(reference)
	bodyFields := reflect.TypeOf(body)

	if referenceFields.NumField() > bodyFields.NumField() {
		httpErr = InvalidBodyError
	}

	for i := 0; i < referenceFields.NumField(); i++ {
		field := referenceFields.Field(i)
		if _, found := bodyFields.FieldByName(field.Name); found {
			httpErr = InvalidBodyError
		}
	}
	return
}

//func buildRoute(m *MinihyperProxy, target func(http.ResponseWriter, *http.Request, *MinihyperProxy)) func(http.ResponseWriter, *http.Request) {
//	return func(resp http.ResponseWriter, req *http.Request) {
//		resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
//		target(resp, req, m)
//	}
//}

func buildRoute(m *MinihyperProxy, target func(http.ResponseWriter, *http.Request, *MinihyperProxy, **HttpError) (reponse interface{})) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		var httpErr *HttpError

		defer throwError(resp, m, &httpErr)
		resp.Header().Set("Content-Type", "application/json; charset=UTF-8")

		response := target(resp, req, m, &httpErr)

		if httpErr == nil {
			json.NewEncoder(resp).Encode(response)
		}
	}
}

func buildRoute2(m *MinihyperProxy, referenceObject interface{}, target func(interface{}, *MinihyperProxy) (reponse interface{}, httpErr **HttpError)) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		var httpErr *HttpError
		var object interface{}

		defer throwError(resp, m, &httpErr)
		resp.Header().Set("Content-Type", "application/json; charset=UTF-8")

		if httpErr = unmarshalBody(resp, req, &object, m); httpErr == nil {
			if httpErr = validateBody(object, referenceObject); httpErr == nil {
				if response, httpErr := target(object, m); httpErr == nil {
					json.NewEncoder(resp).Encode(response)
				}
			}
		}
	}
}

func getServers(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {
	servers := m.GetServersInfo()
	response = ListServersResponse{Info: servers}
	return
}
func getServer(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {

	var getServerRequest GetServerRequest
	var serverInfo ServerInfo

	*httpErr = unmarshalBody(resp, req, &getServerRequest, m)

	if *httpErr == nil {
		serverInfo, *httpErr = m.GetServerInfo(getServerRequest.Name)
		if *httpErr == nil {
			response = ListServersResponse{Info: []ServerInfo{serverInfo}}
		}
	}
	return
}

func getServer2(getServerRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr **HttpError) {

	var serverInfo ServerInfo
	if serverInfo, *httpErr = m.GetServerInfo(getServerRequest.Name); *httpErr == nil {
		response = ListServersResponse{Info: []ServerInfo{serverInfo}}
	}
	return
}

func getProxies(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {
	serversInfo := m.GetProxiesInfo()
	json.NewEncoder(resp).Encode(ListServersResponse{Info: serversInfo})
	return
}

func getProxy(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {

	var getServerRequest GetServerRequest
	var serverInfo ServerInfo

	*httpErr = unmarshalBody(resp, req, &getServerRequest, m)

	if *httpErr == nil {
		serverInfo, *httpErr = m.GetProxyInfo(getServerRequest.Name)
		if *httpErr == nil {
			response = ListServersResponse{Info: []ServerInfo{serverInfo}}
		}
	}
	return
}

func getProxyMap(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {
	var getServerRequest GetServerRequest
	var proxyMap map[string]string

	*httpErr = unmarshalBody(resp, req, &getServerRequest, m)

	if *httpErr == nil {
		proxyMap, *httpErr = m.GetProxyMap(getServerRequest.Name)
		if *httpErr == nil {
			response = ProxyMapResponse{ProxyMap: proxyMap}
		}
	}
	return
}

func createProxy(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {
	var createProxyRequest CreateProxyRequest
	var name, hostname string

	*httpErr = unmarshalBody(resp, req, &createProxyRequest, m)

	if httpErr == nil {
		name, hostname, *httpErr = m.startProxyServer(createProxyRequest.Name, createProxyRequest.Hostname)
		if httpErr == nil {
			response = CreateProxyResponse{Name: createProxyRequest.Name, Hostname: hostname, Port: name}
		}
	}
	return
}

func createRoute(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {

	var createRouteRequest CreateRouteRequest
	var routeURL, targetURL *url.URL

	*httpErr = unmarshalBody(resp, req, &createRouteRequest, m)

	routeURL, err := url.Parse(createRouteRequest.Route)
	if err != nil {
		*httpErr = URLParsingError
	}

	targetURL, err = url.Parse(createRouteRequest.Target)
	if err != nil {
		*httpErr = URLParsingError
	}
	if *httpErr == nil {
		*httpErr = m.addProxyRedirect(createRouteRequest.Name, routeURL, targetURL)
		if *httpErr == nil {
			response = createRouteRequest
		}
	}
	return
}

func getHoppers(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {
	serversInfo := m.GetHoppersInfo()
	response = ListServersResponse{Info: serversInfo}
	return
}

func getHopper(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {

	var getServerRequest GetServerRequest
	var serverInfo ServerInfo

	*httpErr = unmarshalBody(resp, req, &getServerRequest, m)

	if *httpErr == nil {
		serverInfo, *httpErr = m.GetHopperInfo(getServerRequest.Name)
		if *httpErr == nil {
			response = ListServersResponse{Info: []ServerInfo{serverInfo}}
		}
	}
	return
}

func createHopper(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {

	var createHopperRequest CreateHopperRequest
	var hostname, incomingPort, outgoingPort string

	*httpErr = unmarshalBody(resp, req, &createHopperRequest, m)

	if httpErr == nil {
		hostname, incomingPort, outgoingPort, *httpErr = m.startHopperServer(createHopperRequest.Name, createHopperRequest.Hostname)
		if httpErr == nil {
			response = CreateHopperResponse{Name: createHopperRequest.Name, Hostname: hostname, IncomingPort: incomingPort, OutgoingPort: outgoingPort}
		}
	}
	return
}

func getHops(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {
	var getHopsRequest GetHopsRequest
	var incomingHops, outgoingHops map[string]*url.URL

	if *httpErr == nil {
		incomingHops, *httpErr = m.GetIncomingHops(getHopsRequest.Name)
		outgoingHops, *httpErr = m.GetOutgoingHops(getHopsRequest.Name)
		if httpErr == nil {
			response = GetHopsResponse{IncomingHops: incomingHops, OutgoingHops: outgoingHops}
		}
	}
	return
}

func getIncomingHops(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {
	var getHopsRequest GetHopsRequest
	var hops map[string]*url.URL

	*httpErr = unmarshalBody(resp, req, &getHopsRequest, m)

	if *httpErr == nil {
		hops, *httpErr = m.GetIncomingHops(getHopsRequest.Name)
		if *httpErr == nil {
			response = GetIncomingHopsResponse{IncomingHops: hops}
		}
	}
	return
}

func getOutgoingHops(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {
	var getHopsRequest GetHopsRequest
	var hops map[string]*url.URL

	*httpErr = unmarshalBody(resp, req, &getHopsRequest, m)

	if httpErr == nil {
		hops, *httpErr = m.GetOutgoingHops(getHopsRequest.Name)
		if httpErr == nil {
			response = GetOutgoingHopsResponse{OutgoingHops: hops}
		}
	}
	return
}

func createIncomingHop(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {

	var createIncomingHopRequest CreateIncomingHopRequest
	var routeURL, targetURL *url.URL

	*httpErr = unmarshalBody(resp, req, &createIncomingHopRequest, m)

	routeURL, err := url.Parse(createIncomingHopRequest.Route)
	if err != nil {
		*httpErr = URLParsingError
	}

	targetURL, err = url.Parse(createIncomingHopRequest.Target)
	if err != nil {
		*httpErr = URLParsingError
	}
	if *httpErr == nil {
		*httpErr = m.ReceiveHop(createIncomingHopRequest.Name, routeURL, targetURL)
		if *httpErr == nil {
			response = createIncomingHopRequest
		}
	}
	return
}

func createOutgoingHop(resp http.ResponseWriter, req *http.Request, m *MinihyperProxy, httpErr **HttpError) (response interface{}) {

	var createIncomingHopRequest CreateIncomingHopRequest
	var routeURL, targetURL *url.URL

	*httpErr = unmarshalBody(resp, req, &createIncomingHopRequest, m)

	routeURL, err := url.Parse(createIncomingHopRequest.Route)
	if err != nil {
		*httpErr = URLParsingError
	}

	targetURL, err = url.Parse(createIncomingHopRequest.Target)
	if err != nil {
		*httpErr = URLParsingError
	}
	if *httpErr == nil {
		*httpErr = m.AddHop(createIncomingHopRequest.Name, routeURL, targetURL)
		if *httpErr == nil {
			response = createIncomingHopRequest
		}
	}
	return
}

func BuildAPI(m *MinihyperProxy) *mux.Router {

	m.InfoLog.Printf("Initializing API")

	httpMux := mux.NewRouter().StrictSlash(true)
	httpMux.HandleFunc("/servers", buildRoute(m, getServers)).Methods("GET")
	httpMux.HandleFunc("/server", buildRoute2(m, GetServerRequest{}, getServer2)).Methods("GET")

	httpMux.HandleFunc("/proxies", buildRoute(m, getProxies)).Methods("GET")
	httpMux.HandleFunc("/proxy", buildRoute(m, createProxy)).Methods("POST")
	httpMux.HandleFunc("/proxy", buildRoute(m, getProxy)).Methods("GET")

	httpMux.HandleFunc("/proxy/route", buildRoute(m, getProxyMap)).Methods("GET")
	httpMux.HandleFunc("/proxy/route", buildRoute(m, createRoute)).Methods("POST")

	httpMux.HandleFunc("/hoppers", buildRoute(m, getHoppers)).Methods("GET")
	httpMux.HandleFunc("/hopper", buildRoute(m, createHopper)).Methods("POST")
	httpMux.HandleFunc("/hopper", buildRoute(m, getHopper)).Methods("GET")

	httpMux.HandleFunc("/hopper/hop/out", buildRoute(m, getOutgoingHops)).Methods("GET")
	httpMux.HandleFunc("/hopper/hop/in", buildRoute(m, getIncomingHops)).Methods("GET")
	httpMux.HandleFunc("/hopper/hop/out", buildRoute(m, createOutgoingHop)).Methods("POST")
	httpMux.HandleFunc("/hopper/hop/in", buildRoute(m, createIncomingHop)).Methods("POST")

	return httpMux
}

func serveAPI(port string, httpMux *mux.Router) {
	http.ListenAndServe(":"+port, httpMux)
}
