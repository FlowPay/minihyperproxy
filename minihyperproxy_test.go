package minihyperproxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestGenerale(t *testing.T) {
	os.Setenv("PROXY_SERVER", "7052")

	m := NewMinihyperProxy()
	m.startHopperServer("prova")

	m.AddHop("prova", &url.URL{Host: "google.com", Scheme: "http"}, &url.URL{Host: "localhost:7052", Scheme: "http"})
	m.ReceiveHop("prova", &url.URL{Host: "google.com", Scheme: "http"}, &url.URL{Host: "localhost:7052", Scheme: "http"})

	fmt.Printf("Incoming hops: %+v\n", m.GetIncomingHops("prova"))
	fmt.Printf("Outgoing hops: %+v\n", m.GetOutgoingHops("prova"))

	time.Sleep(10 * time.Second)

	resp, err := http.Get("http://localhost:7053")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(body))
	m.stopServer("prova")
}
