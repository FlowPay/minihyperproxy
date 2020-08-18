package minihyperproxy

//Usare http/url
import (
	"log"
	"net/url"
	"os"
	"strconv"
)

type HopperServer struct {
	errorLog              *log.Logger
	warnLog               *log.Logger
	infoLog               *log.Logger
	latestPort            int
	OutgoingHopsReference map[string]string
	OutgoingHops          map[string]*ProxyServer
	IncomingHopProxy      *ProxyServer
}

func NewHopperServer(serverName string, incomingHopPort string, outgoingHopFirstPort string) *HopperServer {
	lastestPort, _ := strconv.Atoi(outgoingHopFirstPort)

	s := &HopperServer{OutgoingHops: make(map[string]*ProxyServer),
		latestPort:            lastestPort,
		infoLog:               log.New(os.Stdout, serverName+"-INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLog:               log.New(os.Stdout, serverName+"-WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog:              log.New(os.Stdout, serverName+"-ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		OutgoingHopsReference: make(map[string]string)}

	s.init(incomingHopPort)
	return s
}

func reduceTargetHop(target *url.URL, hop *url.URL) (newTarget *url.URL, newFullRoute *url.URL) {

	newTarget = &url.URL{Host: target.Host, Scheme: target.Scheme}
	newFullRoute = &url.URL{Host: hop.Host, Scheme: hop.Scheme}
	return
}

func (h *HopperServer) init(incomingHopPort string) {
	h.IncomingHopProxy = NewProxyServer("IncomingHopProxy:"+incomingHopPort, incomingHopPort)
	h.IncomingHopProxy.StartIncomingHopProxy(h.OutgoingHops)
}

func (h *HopperServer) Serve() {
	for _, s := range h.OutgoingHops {
		if s.Status == "Down" {
			s.Serve()
		}
	}
	if h.IncomingHopProxy.Status == "Down" {
		h.IncomingHopProxy.Serve()
	}
}

func (h *HopperServer) Stop() {
	for _, s := range h.OutgoingHops {
		if s.Status == "Down" {
			s.Stop()
		}
	}
	if h.IncomingHopProxy.Status == "Down" {
		h.Stop()
	}
}

func (h *HopperServer) putOutgoingHop(target *url.URL, hop *url.URL) *url.URL {

	portString := strconv.Itoa(h.latestPort)

	h.OutgoingHops[target.EscapedPath()] = NewProxyServer("OutgoingHopProxy:"+portString, portString)
	h.OutgoingHops[target.EscapedPath()].NewHopperSenderProxy(hop, target)
	h.OutgoingHopsReference[target.Host] = hop.Host
	h.latestPort++
	return target
}

func (h *HopperServer) deleteOutgoingHop(target *url.URL) *url.URL {
	delete(h.OutgoingHops, target.EscapedPath())
	delete(h.OutgoingHopsReference, target.EscapedPath())
	return target
}

func (h *HopperServer) putIncomingHop(target *url.URL, newIncomingRoute *url.URL) *url.URL {

	h.infoLog.Printf("Creating incoming hop from %v to %v", target, newIncomingRoute)
	h.IncomingHopProxy.NewHopperReceiverProxy(newIncomingRoute, target)
	return target
}

func (h *HopperServer) deleteIncomingHop(target *url.URL) *url.URL {
	h.IncomingHopProxy.DeleteProxy(target)
	return target
}

func (h *HopperServer) serveHop(target *url.URL) {
	h.OutgoingHops[target.EscapedPath()].Serve()
}

func (h *HopperServer) BuildNewOutgoingHop(target *url.URL, hop *url.URL) {
	target, hop = reduceTargetHop(target, hop)
	h.putOutgoingHop(target, hop)
	h.serveHop(target)
}

func (h *HopperServer) BuildNewIncomingHop(target *url.URL, hop *url.URL) {
	target, newIncomingRoute := reduceTargetHop(target, hop)
	h.putIncomingHop(target, newIncomingRoute)
}

func (h *HopperServer) getIncomingHops() map[string]string {
	return h.IncomingHopProxy.ProxyReference
}

func (h *HopperServer) getOutgoingHops() map[string]string {
	return h.OutgoingHopsReference
}

func (h *HopperServer) Type() string {
	return "Hopper Server"
}
