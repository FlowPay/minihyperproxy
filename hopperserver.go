package minihyperproxy

//Usare http/url
import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type HopperServer struct {
	ServerName            string
	Hostname              string
	errorLog              *log.Logger
	warnLog               *log.Logger
	infoLog               *log.Logger
	outgoingHopPort       int
	incomingHopPort       int
	IncomingHopsReference map[string]*url.URL
	OutgoingHopsReference map[string]*url.URL
	IncomingHopProxy      *ProxyServer
	OutgoingHopProxy      *ProxyServer
	Status                string
}

func NewHopperServer(serverName string, hostname string, incomingHopPort string, outgoingHopPort string) *HopperServer {
	outgoingHopPortInt, _ := strconv.Atoi(outgoingHopPort)
	incomingHopPortInt, _ := strconv.Atoi(incomingHopPort)

	s := &HopperServer{
		ServerName:            serverName,
		Hostname:              hostname,
		infoLog:               log.New(os.Stdout, serverName+"-INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLog:               log.New(os.Stdout, serverName+"-WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog:              log.New(os.Stdout, serverName+"-ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		outgoingHopPort:       outgoingHopPortInt,
		incomingHopPort:       incomingHopPortInt,
		OutgoingHopsReference: make(map[string]*url.URL),
		IncomingHopsReference: make(map[string]*url.URL),
		Status:                "Down"}

	s.init(hostname, incomingHopPort, outgoingHopPort)
	return s
}

func (h *HopperServer) Name() string {
	return h.ServerName
}

func (h *HopperServer) outgoingHopperDirector(req *http.Request) {
	tempString := strings.SplitAfterN(req.URL.EscapedPath(), "/", 3)
	targetHost := strings.Trim(tempString[1], "/")
	targetPath := ""
	if len(tempString) == 3 {
		targetPath = tempString[2]
	}
	if newURL, ok := h.OutgoingHopsReference[targetHost]; ok {
		if _, ok := req.Header["X-MHP-Target-Host"]; !ok {
			req.Header.Set("X-MHP-Target-Host", targetHost)
			if _, ok := req.Header["X-MHP-Target-Scheme"]; !ok {
				req.Header.Set("X-MHP-Target-Scheme", req.URL.Scheme)
			}
			if _, ok := req.Header["X-MHP-Target-Path"]; !ok {
				req.Header.Set("X-MHP-Target-Path", targetPath)
			}
			if _, ok := req.Header["X-MHP-Target-Query"]; !ok {
				req.Header.Set("X-MHP-Target-Query", req.URL.RawQuery)
			}
			if _, ok := req.Header["X-MHP-Forwarded-Host"]; !ok {
				req.Header.Set("X-MHP-Forwarded-Host", req.Header.Get("Host"))
			}
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		req.Header.Set("X-Forwarded-Host", req.Header.Get("X-MHP-Forwarded-Host"))
		req.URL = newURL
		req.Host = newURL.Host
	} else {
		_, cancel := context.WithCancel(req.Context())
		cancel()
	}
}

func (h *HopperServer) incomingHopperDirector(req *http.Request) {
	if req.Header.Get("X-MHP-Target-Host") == "" {
		_, cancel := context.WithCancel(req.Context())
		cancel()
	}
	targetHost := req.Header.Get("X-MHP-Target-Host")
	targetPath := req.Header.Get("X-MHP-Target-Path")
	targetQuery := req.Header.Get("X-MHP-Target-Query")
	targetScheme := req.Header.Get("X-MHP-Target-Scheme")
	if newURL, ok := h.IncomingHopsReference[targetHost]; ok {
		if newURL.Host == targetHost {
			req.Header.Set("X-Forwarded-Host", req.Header.Get("X-MHP-Forwarded-Host"))
		}
		req.URL = newURL
		req.URL.Path = targetPath
		req.URL.RawQuery = targetQuery
		if targetScheme == "" {
			targetScheme = "http"
		}
		req.URL.Scheme = targetScheme
		req.Host = newURL.Host

		req.Header.Del("X-MHP-Target-Host")
		req.Header.Del("X-MHP-Target-Path")
		req.Header.Del("X-MHP-Target-Query")
		req.Header.Del("X-MHP-Target-Scheme")
	}
}

func (h *HopperServer) serveOutgoingRequest(rProxy *httputil.ReverseProxy, resp http.ResponseWriter, req *http.Request) {
	targetHost := strings.Trim(strings.SplitAfter(req.URL.EscapedPath(), "/")[1], "/")
	if _, ok := h.OutgoingHopsReference[targetHost]; !ok {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte("500 - Hop not registered for " + targetHost))
	} else {
		rProxy.ServeHTTP(resp, req)
	}
}

func (h *HopperServer) serveIncomingRequest(rProxy *httputil.ReverseProxy, resp http.ResponseWriter, req *http.Request) {
	if _, ok := h.IncomingHopsReference[req.Header.Get("X-MHP-Target-Host")]; !ok {
		resp.WriteHeader(http.StatusInternalServerError)
		resp.Write([]byte("500 - Hop not registered for " + req.Header.Get("X-MHP-Target-Host")))
	} else {
		rProxy.ServeHTTP(resp, req)
	}
}

func reduceTargetHop(target *url.URL, hop *url.URL) (newTarget *url.URL, newFullRoute *url.URL) {

	newTarget = &url.URL{Host: target.Host, Scheme: target.Scheme}
	newFullRoute = &url.URL{Host: hop.Host, Scheme: hop.Scheme}
	return
}

func (h *HopperServer) init(hostname string, incomingHopPort string, outgoingHopPort string) {

	h.IncomingHopProxy = NewProxyServer("IncomingHopProxy: "+hostname+":"+incomingHopPort, hostname, incomingHopPort)
	h.OutgoingHopProxy = NewProxyServer("OutgoingHopProxy: "+hostname+":"+outgoingHopPort, hostname, outgoingHopPort)
	h.IncomingHopProxy.StartIncomingHopProxy(h.incomingHopperDirector, h.serveIncomingRequest)
	h.OutgoingHopProxy.StartOutgoingHopProxy(h.outgoingHopperDirector, h.serveOutgoingRequest)
}

func (h *HopperServer) Serve() {
	h.Status = "Up"
	h.OutgoingHopProxy.Serve()
	h.IncomingHopProxy.Serve()
}

func (h *HopperServer) Stop() {
	h.Status = "Down"
	h.OutgoingHopProxy.Stop()
	h.IncomingHopProxy.Stop()
}

func (h *HopperServer) putOutgoingHop(target *url.URL, hop *url.URL) *url.URL {
	h.infoLog.Printf("Creating outgoing hop for %v, hop %v", target, hop)
	hostname := target.Hostname()
	h.OutgoingHopsReference[hostname] = hop
	if val, ok := h.IncomingHopsReference[hostname]; ok && val.Host == hostname {
		h.IncomingHopsReference[hostname] = hop
	}
	return target
}

func (h *HopperServer) deleteOutgoingHop(target *url.URL) *url.URL {
	h.infoLog.Printf("Deleting hop to %v", target)
	delete(h.OutgoingHopsReference, target.Hostname())
	return target
}

func (h *HopperServer) putIncomingHop(target *url.URL) *url.URL {
	hostname := target.Hostname()
	h.infoLog.Printf("Creating incoming hop for %v", target)
	if _, ok := h.OutgoingHopsReference[hostname]; ok {
		h.IncomingHopsReference[hostname] = &url.URL{Host: "localhost:" + h.OutgoingHopProxy.ServerPort}
	} else {
		h.IncomingHopsReference[hostname] = target
	}
	return target
}

func (h *HopperServer) deleteIncomingHop(target *url.URL) *url.URL {
	h.IncomingHopProxy.DeleteProxy(target)
	return target
}
func (h *HopperServer) BuildNewOutgoingHop(target *url.URL, hop *url.URL) {
	target, hop = reduceTargetHop(target, hop)
	h.putOutgoingHop(target, hop)
}

func (h *HopperServer) BuildNewIncomingHop(target *url.URL, hop *url.URL) {
	target, _ = reduceTargetHop(target, hop)
	h.putIncomingHop(target)
}

func (h *HopperServer) getIncomingHops() map[string]*url.URL {
	return h.IncomingHopsReference
}

func (h *HopperServer) getOutgoingHops() map[string]*url.URL {
	return h.OutgoingHopsReference
}

func (h *HopperServer) Type() string {
	return "Hopper"
}

func (s *HopperServer) Info() *map[string]interface{} {
	ret := make(map[string]interface{})
	ret["Name"] = s.ServerName
	ret["IncomingPort"] = s.incomingHopPort
	ret["OutgoingPort"] = s.outgoingHopPort
	ret["Type"] = s.Type()
	ret["Status"] = s.Status
	return &ret
}
