package common

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	CommonComfig `yaml:"common"`
	MasterConfig `yaml:"master"`
	WorkerConfig `yaml:"worker"`
}
type CommonComfig struct {
	EtcdConfig  `yaml:"etcd"`
	MongoConfig `yaml:"mongo"`
}
type EtcdConfig struct {
	Endpoints   []string `yaml:"endpoints"`
	DialTimeout int64    `yaml:"timeout"`
}
type MongoConfig struct {
	Uri            string `yaml:"uri"`
	ConnectTimeout int    `yaml:"timeout"`
	Database       string `yaml:"database"`
	Collection     string `yaml:"collection"`
}
type MasterConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type WorkerConfig struct {
	MaxLog int `yaml:"maxlog"`
}

var conf *Config

func InitConfig(cfgfile string) error {
	var (
		bytes []byte
		err   error
	)
	conf = new(Config)
	if bytes, err = ioutil.ReadFile(cfgfile); err != nil {
		return err
	}
	if err = yaml.Unmarshal(bytes, &conf); err != nil {
		return err
	}
	return err
}

func GetConfig() *Config {
	if conf == nil {
		InitConfig("./settings/default.yml")
		return conf
	}
	return conf
}
