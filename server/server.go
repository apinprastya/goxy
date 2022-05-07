package server

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
)

var version = "1.0.0"

var config *Config
var httpServer *http.Server
var httpsServer *http.Server

//StartServer Start the server
func StartServer() {
	config = readConfig()
	allVHosts := config.AllVHost()
	m := &autocert.Manager{
		Cache:      autocert.DirCache(config.Cache),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(allVHosts...),
	}

	httpsServer = &http.Server{Addr: ":443", TLSConfig: m.TLSConfig()}

	for _, vhost := range config.VHosts {
		url, err := url.Parse(vhost.Target)
		if err != nil {
			fmt.Printf("WARNING! domain: %s failed parse URL: %s\n", vhost.Target, err)
			continue
		}
		prox := httputil.NewSingleHostReverseProxy(url)
		for _, domain := range vhost.Domain {
			http.HandleFunc(domain+"/", func(w http.ResponseWriter, req *http.Request) {
				if vhost.BasicAuth != nil {
					username, password, ok := req.BasicAuth()
					if ok {
						usernameHash := sha256.Sum256([]byte(username))
						passwordHash := sha256.Sum256([]byte(password))
						expectedUsernameHash := sha256.Sum256([]byte(vhost.BasicAuth.Username))
						expectedPasswordHash := sha256.Sum256([]byte(vhost.BasicAuth.Password))
						usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
						passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)
						if usernameMatch && passwordMatch {
							prox.ServeHTTP(w, req)
						}
						return
					}
					w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				} else {
					prox.ServeHTTP(w, req)
				}
			})
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defaultHTTPResponse(w, req)
	})
	go checkSignal()
	go httpRedirector(allVHosts)
	log.Println("Start GOXY", version)
	log.Println("List of virtual hosts :", allVHosts)
	if err := httpsServer.ListenAndServeTLS("", ""); err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}

func httpRedirector(allVHosts []string) {
	httpServer = &http.Server{Addr: ":80"}
	httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !containVHost(allVHosts, req.Host) {
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

func containVHost(allVHosts []string, host string) bool {
	for i := range allVHosts {
		if allVHosts[i] == host {
			return true
		}
	}
	return false
}

func readConfig() *Config {
	appPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	viper.SetConfigName(".goxy")
	viper.AddConfigPath(appPath)
	viper.SetDefault("cache", "/root/cert")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	configJson := path.Join(appPath, ".goxy.json")
	jsonByte, err := os.ReadFile(configJson)
	if err != nil {
		panic(err)
	}
	config := &Config{}
	err = json.Unmarshal(jsonByte, config)
	if err != nil {
		panic(err)
	}
	return config
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
