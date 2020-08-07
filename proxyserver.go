package minihyperproxy

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ProxyServer struct {
	ServerName string
	Port       string
	Status     string
	httpServer *http.Server
	httpMux    *http.ServeMux
	infoLog    *log.Logger
	warnLog    *log.Logger
	errorLog   *log.Logger
	ProxyMap   map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewProxyServer(serverName string, port string, infoLog, warnLog, errorLog *log.Logger) *ProxyServer {
	s := &ProxyServer{ServerName: serverName,
		Port:    port,
		infoLog: infoLog, warnLog: warnLog, errorLog: errorLog,
		Status:   "Down",
		ProxyMap: make(map[string]func(w http.ResponseWriter, r *http.Request))}
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

func (s *ProxyServer) NewProxy(route *url.URL, target *url.URL) {

	s.infoLog.Printf("Creating new proxy from %v to %v", route, target)
	s.ProxyMap[route.RawPath] = func(w http.ResponseWriter, r *http.Request) {
		s.infoLog.Printf("Proxying request to %v", target)
		httputil.NewSingleHostReverseProxy(target).ServeHTTP(w, r)
	}
	s.httpMux.HandleFunc(route.RawPath, s.ProxyMap[route.RawPath])
}

func (s *ProxyServer) DeleteProxy(route *url.URL) {
	s.infoLog.Printf("Deleting proxy for: %v", route)
	delete(s.ProxyMap, route.RawPath)
}

func (s *ProxyServer) Type() string {
	return "ProxyServer"
}
