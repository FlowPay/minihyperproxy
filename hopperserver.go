package minihyperproxy

//Usare http/url
import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type HopperServer struct {
	errorLog    *log.Logger
	warnLog     *log.Logger
	infoLog     *log.Logger
	proxyServer *ProxyServer
	OutgoingHop map[*url.URL]*url.URL
	IncomingHop map[*url.URL]func(w http.ResponseWriter, r *http.Request)
}

func NewHopperServer(serverName string, port string, infoLog, warnLog, errorLog *log.Logger) *HopperServer {
	s := &HopperServer{proxyServer: NewProxyServer(serverName, port, infoLog, warnLog, errorLog)}
	s.init()
	return s
}

func (s *HopperServer) init() {
	s.proxyServer.init()
}

func reduceTargetHop(target *url.URL, hop *url.URL) (newTarget *url.URL, newFullRoute *url.URL) {

	host, port := func(x []string) (string, string) { return x[0], x[1] }(strings.SplitAfterN(target.Host, ":", 2))

	path := target.Path
	newTarget = &url.URL{Host: target.Host}
	newFullRoute = &url.URL{Host: hop.Host, Path: host + "-" + port + "/" + path}
	return
}

func (s *HopperServer) Serve() {
	s.proxyServer.Serve()
}

func (s *HopperServer) Stop() {
	s.proxyServer.Stop()
}

func (h *HopperServer) putOutgoingHop(target *url.URL, hop *url.URL) *url.URL {
	h.OutgoingHop[target] = hop
	return target
}

func (h *HopperServer) deleteOutgoingHop(target *url.URL) *url.URL {
	delete(h.OutgoingHop, target)
	return target
}

func (h *HopperServer) putIncomingHop(target *url.URL, newIncomingRoute *url.URL) *url.URL {

	h.proxyServer.infoLog.Printf("Creating incoming hop from %v to %v", target, newIncomingRoute)

	if _, ok := h.IncomingHop[target]; !ok {
		defer h.proxyServer.httpMux.HandleFunc(strings.SplitAfterN(newIncomingRoute.RawPath, "/", 2)[1], h.IncomingHop[target])
	}

	h.IncomingHop[target] = func(w http.ResponseWriter, r *http.Request) {
		if val, ok := h.OutgoingHop[target]; ok {
			h.proxyServer.infoLog.Printf("Hopping request to  %v", target)
			httputil.NewSingleHostReverseProxy(val).ServeHTTP(w, r)
		} else {
			h.proxyServer.infoLog.Printf("Proxying request to %v", target)
			httputil.NewSingleHostReverseProxy(target).ServeHTTP(w, r)
		}
	}

	return target
}

func (h *HopperServer) deleteIncomingHop(target *url.URL) *url.URL {
	delete(h.IncomingHop, target)
	return target
}

func (h *HopperServer) serveHop(target *url.URL, hop *url.URL) {
	h.proxyServer.infoLog.Printf("Creating outgoing hop from %v to %v", target, hop)
	h.proxyServer.NewProxy(target, hop)
}

func (h *HopperServer) BuildNewOutgoingHop(target *url.URL, hop *url.URL) {
	target, newFullRoute := reduceTargetHop(target, hop)
	h.putOutgoingHop(target, newFullRoute)
	h.serveHop(target, newFullRoute)
}

func (h *HopperServer) BuildNewIncomingHop(target *url.URL, hop *url.URL) {
	target, newIncomingRoute := reduceTargetHop(target, hop)
	h.putIncomingHop(target, newIncomingRoute)
}
