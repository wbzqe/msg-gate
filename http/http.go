package http

import (
	"github.com/patrickmn/go-cache"
	"github.com/wbzqe/msg-gate/config"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var TokenCache *cache.Cache

func Start() {

	TokenCache = cache.New(6000*time.Second, 60*time.Second)

	configProcRoutes()

	addr := config.Config().Http.Listen
	if addr == "" {
		return
	}
	s := &http.Server{
		Addr:           addr,
		MaxHeaderBytes: 1 << 30,
	}
	log.Println("http listening", addr)
	log.Fatalln(s.ListenAndServe())
}
