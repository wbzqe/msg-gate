package http

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/toolkits/web/param"
	"github.com/wbzqe/msg-gate/config"
)

type Smtp struct {
	Address  string
	Username string
	Password string
}

func New(address, username, password string) *Smtp {
	return &Smtp{
		Address:  address,
		Username: username,
		Password: password,
	}
}

func HdMail(w http.ResponseWriter, r *http.Request) {
	cfg := config.Config()
	token := param.String(r, "token", "")
	if cfg.Http.Token != token {
		http.Error(w, "no privilege", http.StatusForbidden)
		return
	}

	tos := param.MustString(r, "tos")
	subject := param.MustString(r, "subject")
	content := param.MustString(r, "content")
	tos = strings.Replace(tos, ",", ";", -1)

	s := New(cfg.Smtp.Addr, cfg.Smtp.Username, cfg.Smtp.Password)
	if hp := strings.Split(cfg.Smtp.Addr, ":"); hp[1] == "25" {
		err := s.SendMail(cfg.Smtp.From, tos, subject, content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "success", http.StatusOK)
		}
	} else {
		err := s.SendMailTLS(cfg.Smtp.From, tos, subject, content)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, "success", http.StatusOK)
		}
	}
}

func (this *Smtp) SendMail(from, tos, subject, body string, contentType ...string) error {
	if this.Address == "" {
		return fmt.Errorf("address is necessary")
	}

	hp := strings.Split(this.Address, ":")
	if len(hp) != 2 {
		return fmt.Errorf("address format error")
	}

	arr := strings.Split(tos, ";")
	count := len(arr)
	safeArr := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if arr[i] == "" {
			continue
		}
		safeArr = append(safeArr, arr[i])
	}

	if len(safeArr) == 0 {
		return fmt.Errorf("tos invalid")
	}

	tos = strings.Join(safeArr, ";")

	b64 := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")

	header := make(map[string]string)
	header["From"] = from
	header["To"] = tos
	header["Subject"] = fmt.Sprintf("=?UTF-8?B?%s?=", b64.EncodeToString([]byte(subject)))
	header["MIME-Version"] = "1.0"

	ct := "text/plain; charset=UTF-8"
	if len(contentType) > 0 && contentType[0] == "html" {
		ct = "text/html; charset=UTF-8"
	}

	header["Content-Type"] = ct
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + b64.EncodeToString([]byte(body))

	auth := smtp.PlainAuth("", this.Username, this.Password, hp[0])

	return smtp.SendMail(this.Address, auth, from, strings.Split(tos, ";"), []byte(message))
}

func (this *Smtp) SendMailTLS(from, tos, subject, body string, contentType ...string) error {
	if this.Address == "" {
		return fmt.Errorf("address is necessary")
	}

	hp := strings.Split(this.Address, ":")
	if len(hp) != 2 {
		return fmt.Errorf("address format error")
	}

	arr := strings.Split(tos, ";")
	count := len(arr)
	safeArr := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if arr[i] == "" {
			continue
		}
		safeArr = append(safeArr, arr[i])
	}

	if len(safeArr) == 0 {
		return fmt.Errorf("tos invalid")
	}

	tos = strings.Join(safeArr, ";")

	b64 := base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")

	header := make(map[string]string)
	header["From"] = from
	header["To"] = tos
	header["Subject"] = fmt.Sprintf("=?UTF-8?B?%s?=", b64.EncodeToString([]byte(subject)))
	header["MIME-Version"] = "1.0"

	ct := "text/plain; charset=UTF-8"
	if len(contentType) > 0 && contentType[0] == "html" {
		ct = "text/html; charset=UTF-8"
	}

	header["Content-Type"] = ct
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + b64.EncodeToString([]byte(body))

	auth := smtp.PlainAuth("", this.Username, this.Password, hp[0])

	host, _, _ := net.SplitHostPort(this.Address)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", this.Address, tlsConfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range strings.Split(tos, ";") {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}
