package minihyperproxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
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
	Port           string
	Status         string
	httpServer     *http.Server
	httpMux        *http.ServeMux
	infoLog        *log.Logger
	warnLog        *log.Logger
	errorLog       *log.Logger
	ProxyReference map[string]string
	ProxyMap       map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewProxyServer(serverName string, port string) *ProxyServer {
	s := &ProxyServer{ServerName: serverName,
		Port:           port,
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
	s.httpMux = http.NewServeMux()
	s.httpServer = &http.Server{Addr: ":" + s.Port,
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
	s.infoLog.Printf("Listening on: " + s.Port)
	s.Status = "Up"
}

func (s *ProxyServer) Stop() {
	if s.Status == "Down" {
		s.warnLog.Printf("Trying to stop server: %s which is already stopped", s.ServerName)
	} else if err := s.httpServer.Shutdown(context.Background()); err != nil {
		s.errorLog.Printf(err.Error())
	}
}
func (s *ProxyServer) StartIncomingHopProxy(nextHopMap) {

	director := func(req *http.Request) {
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}

		fmt.Printf("%+v", req)
	}
	rProxy := &httputil.ReverseProxy{Director: director}

	s.ProxyMap["/"] = func(w http.ResponseWriter, r *http.Request) {

		s.infoLog.Printf("Hopping request to %v")

		rProxy.ServeHTTP(w, r)

	}
	s.ProxyReference["/"] = "hop_route"
	s.httpMux.HandleFunc("/", s.ProxyMap["/"])
}
func (s *ProxyServer) NewHopperSenderProxy(route *url.URL, target *url.URL) {

	s.infoLog.Printf("Creating new hopper proxy from %v to %v", route.EscapedPath(), target)

	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = route.Scheme
		req.URL.Host = route.Host
		req.URL.Path = ""
		req.URL.RawQuery = ""
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		req.Header.Add("X-MHP-TargetScheme", target.Scheme)
		req.Header.Add("X-MHP-TargetHost", target.Host)
		req.Header.Add("X-MHP-TargetRoute", target.EscapedPath())
		req.Header.Add("X-MHP-TargetQuery", targetQuery)

		fmt.Printf("%+v", req)
	}
	rProxy := &httputil.ReverseProxy{Director: director}

	s.ProxyMap[route.EscapedPath()] = func(w http.ResponseWriter, r *http.Request) {

		s.infoLog.Printf("Hopping request to %v", target.Host+target.EscapedPath())

		rProxy.ServeHTTP(w, r)

	}
	s.ProxyReference[route.EscapedPath()] = target.Host + target.EscapedPath()
	s.httpMux.HandleFunc(route.EscapedPath(), s.ProxyMap[route.EscapedPath()])
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
	return "ProxyServer"
}

func (s *ProxyServer) getProxyMap() map[string]string {
	return s.ProxyReference
}
