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

	//m.AddHop("prova", &url.URL{Host: "www.google.com", Scheme: "http"}, &url.URL{Host: "localhost:7052", Scheme: "http"})
	//m.ReceiveHop("prova", &url.URL{Host: "www.google.com", Scheme: "http"}, &url.URL{Host: "localhost:7052", Scheme: "http"})
	target, err := url.Parse("https://google.com/")
	if err != nil {
		t.Fatal(err)
	}
	m.startProxyServer("prova2")
	m.addProxyRedirect("prova2", &url.URL{Path: "/google", Scheme: "http"}, target)

	time.Sleep(5 * time.Second)

	resp, err := http.Get("http://localhost:7052/google")
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
