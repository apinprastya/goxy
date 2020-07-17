package server

import (
	"net/http/httputil"
	"net/url"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDb() {
	var err error
	db, err = gorm.Open(sqlite.Open("domain.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Domain{})
}

//Domain struct
type Domain struct {
	gorm.Model
	Domain string `json:"domain"`
	Target string `json:"target"`
}

func readDomainFromDb() {
	all := []Domain{}
	db.Find(&all)
	for _, d := range all {
		if t, err := url.Parse(d.Target); err == nil {
			allVHosts = append(allVHosts, d.Domain)
			serverConfigMaps[d.Domain] = httputil.NewSingleHostReverseProxy(t)
		}
	}
}

func insertNewDomain(domain, target string) {
	if t, err := url.Parse(target); err == nil {
		allVHosts = append(allVHosts, domain)
		serverConfigMaps[domain] = httputil.NewSingleHostReverseProxy(t)
		db.Create(&Domain{Domain: domain, Target: target})
	}
}
