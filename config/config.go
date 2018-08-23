package config

import (
	"encoding/json"
	"fmt"
	"github.com/toolkits/file"
	"log"
	"sync"
)

type HttpConfig struct {
	Listen string `json:"listen"`
	Token  string `json:"token"`
}

type SmtpConfig struct {
	Addr     string `json:"addr"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
}

type QywxConfig struct {
	CorpID  string `json:"CorpID"`
	AgentId string `json:"AgentId"`
	Secret  string `json:"Secret"`
}

type GlobalConfig struct {
	Debug bool        `json:"debug"`
	Http  *HttpConfig `json:"http"`
	Smtp  *SmtpConfig `json:"smtp"`
	Qywx  *QywxConfig `json:"qywx"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	configLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func Parse(cfg string) error {
	if cfg == "" {
		return fmt.Errorf("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		return fmt.Errorf("configuration file %s is nonexistent", cfg)
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		return fmt.Errorf("read configuration file %s fail %s", cfg, err.Error())
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		return fmt.Errorf("parse configuration file %s fail %s", cfg, err.Error())
	}

	configLock.Lock()
	defer configLock.Unlock()
	config = &c

	log.Println("load configuration file", cfg, "successfully")
	return nil
}
