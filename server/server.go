package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
)

type basicAuth struct {
	Realm    string
	Username string
	Password string
}

type serverConfig struct {
	Domain []string
	Target *url.URL
	Auth   *basicAuth
}

var version = "0.0.1"

var serverConfigs = []*serverConfig{}
var allVHosts = []string{}
var httpServer *http.Server
var httpsServer *http.Server

// StartServer Start the server
func StartServer() {
	readConfig()
	m := &autocert.Manager{
		Cache:      autocert.DirCache(viper.GetString("cache")),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(allVHosts...),
	}

	httpsServer = &http.Server{Addr: ":443", TLSConfig: m.TLSConfig()}

	for _, vhost := range serverConfigs {
		prox := httputil.NewSingleHostReverseProxy(vhost.Target)
		for _, domain := range vhost.Domain {
			http.HandleFunc(domain+"/", func(w http.ResponseWriter, req *http.Request) {
				if vhost.Auth != nil {
					username, password, ok := req.BasicAuth()
					if !ok || (username != vhost.Auth.Username && password != vhost.Auth.Password) {
						w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, vhost.Auth.Realm))
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
				}
				prox.ServeHTTP(w, req)
			})
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defaultHTTPResponse(w, req)
	})
	go checkSignal()
	go httpRedirector()
	log.Println("Start GOXY", version)
	log.Println("List of virtual hosts :", allVHosts)
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
			defaultHTTPResponse(w, req)
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
		if allVHosts[i] == host {
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

		//auth
		var basic *basicAuth
		authInterface := vh["auth"]
		if authInterface != nil {
			auth := authInterface.(map[string]any)
			username := auth["username"].(string)
			password := auth["password"].(string)
			realm := auth["realm"].(string)
			basic = &basicAuth{
				Username: username,
				Password: password,
				Realm:    realm,
			}
		}
		serverConfigs = append(serverConfigs, &serverConfig{Domain: doms, Target: target, Auth: basic})
	}
}

func checkSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
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
