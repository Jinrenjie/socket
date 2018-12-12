package config

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Redis struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Pass string `yaml:"pass"`
}

type Kafka struct {
	Broker string `yaml:"broker"`
	Topic  string `yaml:"topic"`
}

type Web struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Name string `yaml:"name"`
	Pass string `yaml:"pass"`
}

type Api struct {
	Prefix string `yaml:"prefix"`
}

type Socket struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Env struct {
	Web    Web    `yaml:"web"`
	Api    Api    `yaml:"api"`
	Redis  Redis  `yaml:"redis"`
	Kafka  Kafka  `yaml:"kafka"`
	Socket Socket `yaml:"socket"`
	Pid    string `yaml:"pid"`
}

func (conf *Env) Get(key string) interface{} {
	return conf.Redis
}

// Load service config
func Init(path string) *Env {
	env := &Env{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(file, env)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return env
}
