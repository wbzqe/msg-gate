package http

import (
	"github.com/wbzqe/msg-gate/config"
	"net/http"
)

func configProcRoutes() {

	http.HandleFunc("/sender/mail", HdMail)
	http.HandleFunc("/sender/qywx", HdQywx)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(config.VERSION))
	})

}
