package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

const usageStr string = "%s [options] domain_0 alias_0 ... alias_i, ... domain_i alias_0 ... alias_i\n"

func newProxyMux(doms Domains) *proxyMux {
	us := doms.Urls()
	p := new(proxyMux)
	p.rProxys = make(map[url.URL]*httputil.ReverseProxy, len(us))
	for _, u := range us {
		p.rProxys[*u] = httputil.NewSingleHostReverseProxy(u)
	}
	return p
}

type proxyMux struct {
	rProxys map[url.URL]*httputil.ReverseProxy
}

func (pm *proxyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// pm.rProxys[*r.URL].ServeHTTP(w, r)
	if f, ok := pm.rProxys[*r.URL]; ok {
		f.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

func commaSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Return nothing if at end of file and no data passed
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := strings.Index(string(data), ","); i >= 0 {
		return i + 1, data[0:i], nil
	}
	// If at end of file with data return the data
	if atEOF {
		return len(data), data, nil
	}
	return
}

func getDomains(arguments []string) (Domains, error) {
	s := bufio.NewScanner(strings.NewReader(strings.Join(flag.Args(), " ")))
	s.Split(commaSplit)
	var doms Domains
	for s.Scan() {
		args := strings.Fields(s.Text())
		doms = append(doms, &Domain{args[0], args[1:], nil})
	}
	if err := doms.check(); err != nil {
		return nil, err
	}
	return doms, nil
}

type Domains []*Domain

func (doms Domains) check() (err error) {
	for _, d := range doms {
		d.dUrl, err = url.Parse(d.name)
		if err != nil {
			return err
		}
		for _, a := range d.alias {
			d.dUrl, err = url.Parse(a)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (doms Domains) Urls() (out []*url.URL) {
	for _, d := range doms {
		out = append(out, d.dUrl)
	}
	return out
}

// returns each domain
func (doms Domains) Doms() []string {
	var out []string
	for _, d := range doms {
		out = append(out, d.name)
	}
	return out
}

func (doms Domains) All() []string {
	var out []string
	for _, d := range doms {
		out = append(out, append(d.alias, d.name)...)
	}
	return out
}

type Domain struct {
	name  string
	alias []string
	dUrl  *url.URL
}

func main() {

	rTimeoutRaw := flag.String("rt", "5s", "set read timeout, default 5 seconds")
	wTimeoutRaw := flag.String("wt", "10s", "set write timeout, default 10 seconds")
	iTimeoutRaw := flag.String("it", "120s", "set idle timeout, default 120 seconds")
	flag.Usage = func() {
		fmt.Printf(usageStr, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	doms, err := getDomains(flag.Args())
	if err != nil {
		log.Fatal(err, "\n")
		flag.Usage()
	}

	rTimeout, err := time.ParseDuration(*rTimeoutRaw)
	if err != nil {
		log.Fatal(err)
	}
	wTimeout, err := time.ParseDuration(*wTimeoutRaw)
	if err != nil {
		log.Fatal(err)
	}
	iTimeout, err := time.ParseDuration(*iTimeoutRaw)
	if err != nil {
		log.Fatal(err)
	}
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(doms.All()...),
		Cache:      autocert.DirCache("/srv/certs"), //Folder for storing certificates
	}
	server := &http.Server{
		ReadTimeout:  rTimeout,
		WriteTimeout: wTimeout,
		IdleTimeout:  iTimeout,
		Addr:         ":https",
		Handler:      newProxyMux(doms),
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}
	defer server.Close()

	go http.ListenAndServe(":80", certManager.HTTPHandler(nil))
	fmt.Println("proxy up, using domains: ", doms.Doms())
	log.Fatal(server.ListenAndServeTLS("", ""))
}
