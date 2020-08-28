package minihyperproxy

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

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
func unmarshalBody(req *http.Request, target interface{}) (httpErr *HttpError) {
	bodyBytes, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		httpErr = RequestUnmarshallError
	}
	if err = json.Unmarshal(bodyBytes, target); err != nil {
		httpErr = BodyUnmarshallError
	}
	return
}
func buildRoute(m *MinihyperProxy, referenceObject interface{}, routeFunction RouteFunction) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {

		var httpErr *HttpError
		var obj, response interface{}

		defer throwError(resp, m, &httpErr)
		resp.Header().Set("Content-Type", "application/json; charset=UTF-8")

		if httpErr = unmarshalBody(req, &obj); httpErr == nil {
			if err := mapstructure.Decode(obj, &referenceObject); err == nil {
				if response, httpErr = routeFunction(referenceObject, m); httpErr == nil {
					json.NewEncoder(resp).Encode(response)
				}
			} else {
				httpErr = InvalidBodyError
			}
		}
	}
}

func getServers(getServersRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	servers := m.GetServersInfo()
	response = ListServersResponse{Info: servers}
	return
}

func getServer(getServerRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := getServerRequest.(GetServerRequest)
	if serverInfo, httpErr := m.GetServerInfo(obj.Name); httpErr == nil {
		response = ListServersResponse{Info: []ServerInfo{serverInfo}}
	}
	return
}

func getProxies(getProxiesRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	serversInfo := m.GetProxiesInfo()
	response = ListServersResponse{Info: serversInfo}
	return
}

func getProxy(getProxyRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := getProxyRequest.(GetServerRequest)
	if serverInfo, httpErr := m.GetProxyInfo(obj.Name); httpErr == nil {
		response = ListServersResponse{Info: []ServerInfo{serverInfo}}
	}
	return
}
func getProxyMap(getProxyMapRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := getProxyMapRequest.(GetServerRequest)
	if proxyMap, httpErr := m.GetProxyMap(obj.Name); httpErr == nil {
		response = ProxyMapResponse{ProxyMap: proxyMap}
	}
	return
}
func createProxy(createProxyRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := createProxyRequest.(CreateProxyRequest)
	if name, hostname, httpErr := m.startProxyServer(obj.Name, obj.Hostname); httpErr == nil {
		response = CreateProxyResponse{Name: obj.Name, Hostname: hostname, Port: name}
	}
	return
}

func createRoute(createRouteRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := createRouteRequest.(CreateRouteRequest)
	if routeURL, err := url.Parse(obj.Route); err == nil {
		if targetURL, err := url.Parse(obj.Target); err == nil {
			if httpErr = m.addProxyRedirect(obj.Name, routeURL, targetURL); httpErr == nil {
				response = obj
			}
		} else {
			httpErr = URLParsingError
		}
	} else {
		httpErr = URLParsingError
	}
	return
}

func getHoppers(getHoppersRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	serversInfo := m.GetHoppersInfo()
	response = ListServersResponse{Info: serversInfo}
	return
}

func getHopper(getHopperRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := getHopperRequest.(GetServerRequest)
	if serverInfo, httpErr := m.GetHopperInfo(obj.Name); httpErr == nil {
		response = ListServersResponse{Info: []ServerInfo{serverInfo}}
	}
	return
}

func createHopper(createHopperRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := createHopperRequest.(CreateHopperRequest)
	if incomingport, outgoingport, hostname, httpErr := m.startHopperServer(obj.Name, obj.Hostname); httpErr == nil {
		response = CreateHopperResponse{Name: obj.Name, Hostname: hostname, IncomingPort: incomingport, OutgoingPort: outgoingport}
	}

	return
}

func getHops(getHopsRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := getHopsRequest.(GetHopsRequest)
	if incomingHops, httpErr := m.GetIncomingHops(obj.Name); httpErr == nil {
		if outgoingHops, httpErr := m.GetOutgoingHops(obj.Name); httpErr == nil {
			response = GetHopsResponse{IncomingHops: incomingHops, OutgoingHops: outgoingHops}
		}
	}
	return
}

func getIncomingHops(getIncomingHopsRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := getIncomingHopsRequest.(GetHopsRequest)
	if hops, httpErr := m.GetIncomingHops(obj.Name); httpErr == nil {
		response = GetIncomingHopsResponse{IncomingHops: hops}
	}
	return
}

func getOutgoingHops(getOutgoingHopsRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := getOutgoingHopsRequest.(GetHopsRequest)
	if hops, httpErr := m.GetOutgoingHops(obj.Name); httpErr == nil {
		response = GetOutgoingHopsResponse{OutgoingHops: hops}
	}
	return
}

func createIncomingHop(createIncomingHopRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := createIncomingHopRequest.(CreateIncomingHopRequest)
	if routeURL, err := url.Parse(obj.Route); err == nil {
		if targetURL, err := url.Parse(obj.Target); err == nil {
			if httpErr = m.ReceiveHop(obj.Name, routeURL, targetURL); httpErr == nil {
				response = obj
			}
		} else {
			httpErr = URLParsingError
		}
	} else {
		httpErr = URLParsingError
	}
	return
}

func createOutgoingHop(createOutgoingHopRequest interface{}, m *MinihyperProxy) (response interface{}, httpErr *HttpError) {
	obj := createOutgoingHopRequest.(CreateIncomingHopRequest)
	if routeURL, err := url.Parse(obj.Route); err == nil {
		if targetURL, err := url.Parse(obj.Target); err == nil {
			if httpErr = m.ReceiveHop(obj.Name, routeURL, targetURL); httpErr == nil {
				response = obj
			}
		} else {
			httpErr = URLParsingError
		}
	} else {
		httpErr = URLParsingError
	}
	return
}

func BuildAPI(m *MinihyperProxy) *mux.Router {

	m.InfoLog.Printf("Initializing API")

	httpMux := mux.NewRouter().StrictSlash(true)
	httpMux.HandleFunc("/servers", buildRoute(m, EmptyRequest{}, getServers)).Methods("GET")
	httpMux.HandleFunc("/server", buildRoute(m, GetServerRequest{}, getServer)).Methods("GET")

	httpMux.HandleFunc("/proxies", buildRoute(m, EmptyRequest{}, getProxies)).Methods("GET")
	httpMux.HandleFunc("/proxy", buildRoute(m, CreateProxyRequest{}, createProxy)).Methods("POST")
	httpMux.HandleFunc("/proxy", buildRoute(m, GetServerRequest{}, getProxy)).Methods("GET")

	httpMux.HandleFunc("/proxy/route", buildRoute(m, GetServerRequest{}, getProxyMap)).Methods("GET")
	httpMux.HandleFunc("/proxy/route", buildRoute(m, CreateRouteRequest{}, createRoute)).Methods("POST")

	httpMux.HandleFunc("/hoppers", buildRoute(m, EmptyRequest{}, getHoppers)).Methods("GET")
	httpMux.HandleFunc("/hopper", buildRoute(m, CreateHopperRequest{}, createHopper)).Methods("POST")
	httpMux.HandleFunc("/hopper", buildRoute(m, GetServerRequest{}, getHopper)).Methods("GET")

	httpMux.HandleFunc("/hopper/hop", buildRoute(m, GetHopsRequest{}, getHops)).Methods("GET")
	httpMux.HandleFunc("/hopper/hop/out", buildRoute(m, GetHopsRequest{}, getOutgoingHops)).Methods("GET")
	httpMux.HandleFunc("/hopper/hop/in", buildRoute(m, GetHopsRequest{}, getIncomingHops)).Methods("GET")
	httpMux.HandleFunc("/hopper/hop/out", buildRoute(m, CreateOutgoingHopRequest{}, createOutgoingHop)).Methods("POST")
	httpMux.HandleFunc("/hopper/hop/in", buildRoute(m, CreateIncomingHopRequest{}, createIncomingHop)).Methods("POST")

	return httpMux
}

func serveAPI(port string, httpMux *mux.Router) {
	http.ListenAndServe(":"+port, httpMux)
}
