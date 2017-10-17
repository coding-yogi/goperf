package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Settings ...
type Settings struct {
	RabbitMQ RabbitMQ
}

// RabbitMQ ...
type RabbitMQ struct {
	Host     string
	Port     int32
	Username string
	Password string
}

// GetConfig ...
func GetConfig() (Settings, error) {
	var settings Settings

	pwd, _ := os.Getwd()
	file, err := ioutil.ReadFile(pwd + "/config.json")
	if err != nil {
		return settings, err
	}

	settings = Settings{}
	json.Unmarshal(file, &settings)

	return settings, nil
}
