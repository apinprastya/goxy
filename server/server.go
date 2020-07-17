package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
)

var version = "0.0.1"

var serverConfigMaps map[string]*httputil.ReverseProxy
var allVHosts = []string{}
var httpServer *http.Server
var httpsServer *http.Server

//StartServer Start the server
func StartServer() {
	initDb()
	readConfig()

	m := &autocert.Manager{
		Cache:      autocert.DirCache(viper.GetString("cache")),
		Prompt:     autocert.AcceptTOS,
		Email:      viper.GetString("email"),
		HostPolicy: whiteList,
	}

	handler := newHTTPSHandler()
	httpsServer = &http.Server{Addr: ":443", TLSConfig: m.TLSConfig(), Handler: handler}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defaultHTTPResponse(w, req)
	})

	go checkSignal()
	go httpRedirector()

	log.Println("Start GOXY", version)
	log.Println("All vhosts :", allVHosts)
	if err := httpsServer.ListenAndServeTLS("", ""); err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}

func httpRedirector() {
	httpServer = &http.Server{Addr: ":80"}
	httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !containVHost(req.Host) {
			if req.URL.Path == "/_addnewdomain" && req.Method == "PUT" {
				jdata := struct {
					Domain string `json:"domain"`
					Target string `json:"target"`
				}{}
				if ba, err := ioutil.ReadAll(req.Body); err == nil {
					if err := json.Unmarshal(ba, &jdata); err == nil {
						insertNewDomain(jdata.Domain, jdata.Target)
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(200)
						w.Write([]byte("{\"ok\":true}"))
					} else {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(400)
						w.Write([]byte("{\"ok\":false}"))
					}
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(400)
					w.Write([]byte("{\"ok\":false}"))
				}
			} else {
				defaultHTTPResponse(w, req)
			}
		} else {
			http.Redirect(w, req, "https://"+req.Host+req.URL.String(), http.StatusMovedPermanently)
		}
	})
	if err := httpServer.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}

func containVHost(host string) bool {
	for i := range allVHosts {
		if strings.HasPrefix(allVHosts[i], "*.") {
		} else if allVHosts[i] == host {
			return true
		}
	}
	return false
}

func readConfig() {
	appPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	viper.SetConfigName(".goxy")
	viper.AddConfigPath(appPath)
	viper.SetDefault("cache", "/root/cert")
	viper.SetDefault("email", "no-reply@gmail.com")
	serverConfigMaps = make(map[string]*httputil.ReverseProxy)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	vhosts := viper.Get("vhosts").([]interface{})
	for _, v := range vhosts {
		vh := v.(map[string]interface{})
		domInterface := vh["domain"].([]interface{})
		doms := []string{}
		for _, di := range domInterface {
			doms = append(doms, di.(string))
		}
		allVHosts = append(allVHosts, doms...)
		target, _ := url.Parse(vh["target"].(string))
		for _, d := range doms {
			serverConfigMaps[d] = httputil.NewSingleHostReverseProxy(target)
		}
	}
	readDomainFromDb()
}

func checkSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-sig
	log.Println("Shutting down")
	if httpServer != nil {
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutdown %v", err)
		}
	}
	if httpsServer != nil {
		if err := httpsServer.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutdown %v", err)
		}
	}
	log.Println("Shutdown success!!!")
}

func whiteList(ctx context.Context, host string) error {
	if !containVHost(host) {
		return fmt.Errorf("%s no found", host)
	}
	return nil
}
