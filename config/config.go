package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	RendezvousURL string `json:"rendezvous_url"`
	CodeLen       int    `json:"code_len"`
	confFilename  string
}

func (c *Config) Save() error {
	f, err := os.Create(c.confFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(c)
}

func Load(dataDir string) *Config {
	filename := filepath.Join(dataDir, "wormhole-william.json")
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return &Config{
			confFilename: filename,
		}
	}

	var conf Config
	json.Unmarshal(content, &conf)
	conf.confFilename = filename
	return &conf
}
