package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Entry    string `json:"entry"`
	BuildDir string `json:"buildDir"`
}

func Load(path string) Config {
	c := Config{Entry: "main.aa", BuildDir: "dist"}
	b, err := os.ReadFile(path)
	if err != nil {
		return c
	}
	_ = json.Unmarshal(b, &c)
	return c
}
