package minihyperproxy

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

type ProxyServer struct {
	ServerName     string
	Hostname       string
	ServerPort     string
	Status         string
	httpServer     *http.Server
	httpMux        *mux.Router
	infoLog        *log.Logger
	warnLog        *log.Logger
	errorLog       *log.Logger
	ProxyReference map[string]string
	ProxyMap       map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewProxyServer(serverName string, hostname string, port string) *ProxyServer {

	s := &ProxyServer{ServerName: serverName,
		Hostname:       hostname,
		ServerPort:     port,
		infoLog:        log.New(os.Stdout, serverName+"-INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLog:        log.New(os.Stdout, serverName+"-WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog:       log.New(os.Stdout, serverName+"-ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		Status:         "Down",
		ProxyMap:       make(map[string]func(w http.ResponseWriter, r *http.Request)),
		ProxyReference: make(map[string]string)}
	s.init()

	return s
}

func (s *ProxyServer) init() {
	s.httpMux = mux.NewRouter().StrictSlash(true)
	s.httpServer = &http.Server{Addr: s.Hostname + ":" + s.ServerPort,
		Handler: s.httpMux}
}

func (s *ProxyServer) Serve() {
	s.httpServer.RegisterOnShutdown(func() {
		s.infoLog.Printf("Server: " + s.ServerName + " stopping")
	})
	s.infoLog.Printf("Server: " + s.ServerName + " starting")
	go func() {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			s.errorLog.Printf(err.Error())
		}
	}()
	s.infoLog.Printf("Listening on: " + s.ServerPort)
	s.Status = "Up"
}

func (s *ProxyServer) Stop() {
	if s.Status == "Down" {
		s.warnLog.Printf("Trying to stop server: %s which is already stopped", s.ServerName)
	} else if err := s.httpServer.Shutdown(context.Background()); err != nil {
		s.errorLog.Printf(err.Error())
	}
}
func (s *ProxyServer) StartIncomingHopProxy(director func(*http.Request), serveFunc func(*httputil.ReverseProxy, http.ResponseWriter, *http.Request)) {
	rProxy := &httputil.ReverseProxy{Director: director}
	s.ProxyMap["/"] = func(w http.ResponseWriter, r *http.Request) {
		serveFunc(rProxy, w, r)
	}
	s.ProxyReference["/"] = "incoming_hop_server"
	s.httpMux.PathPrefix("/").HandlerFunc(s.ProxyMap["/"])
}
func (s *ProxyServer) StartOutgoingHopProxy(director func(*http.Request), serveFunc func(*httputil.ReverseProxy, http.ResponseWriter, *http.Request)) {
	rProxy := &httputil.ReverseProxy{Director: director}
	s.ProxyMap["/"] = func(w http.ResponseWriter, r *http.Request) {
		serveFunc(rProxy, w, r)
	}
	s.ProxyReference["/"] = "incoming_hop_server"
	s.httpMux.PathPrefix("/").HandlerFunc(s.ProxyMap["/"])
}

func (s *ProxyServer) NewProxy(route *url.URL, target *url.URL) {

	s.infoLog.Printf("Creating new proxy from %v to %v", route.EscapedPath(), target)

	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	rProxy := &httputil.ReverseProxy{Director: director}
	s.ProxyMap[route.EscapedPath()] = func(w http.ResponseWriter, r *http.Request) {
		s.infoLog.Printf("Proxying request to %v", target.Host+target.EscapedPath())
		r.URL.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = target.Host
		rProxy.ServeHTTP(w, r)
	}
	s.ProxyReference[route.EscapedPath()] = target.Host + target.EscapedPath()
	s.httpMux.HandleFunc(route.EscapedPath(), s.ProxyMap[route.EscapedPath()])
}

func (s *ProxyServer) DeleteProxy(route *url.URL) {
	s.infoLog.Printf("Deleting proxy for: %v", route)
	delete(s.ProxyReference, route.EscapedPath())
	delete(s.ProxyMap, route.EscapedPath())
}

func (s *ProxyServer) Type() string {
	return "Proxy"
}

func (s *ProxyServer) Info() *map[string]interface{} {
	ret := make(map[string]interface{})
	ret["Name"] = s.ServerName
	ret["Port"] = s.ServerPort
	ret["Type"] = s.Type()
	ret["Status"] = s.Status
	return &ret
}

func (s *ProxyServer) getProxyMap() map[string]string {
	return s.ProxyReference
}
