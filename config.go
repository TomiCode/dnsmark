package main

import "os"
import "log"
import "encoding/json"

type RouterConfig struct {
  Addr string `json:"addr"`
  User string `json:"username"`
  Pass string `json:"password"`
}

type CloudflareConfig struct {
  Email string `json:"email"`
  Key string `json:"apikey"`

  ZoneId string `json:"zoneid"`
  Dns []string `json:"dns"`
}

type ServiceConfig struct {
  RouterConfig `json:"router"`
  CloudflareConfig `json:"cloudflare"`
}

func LoadServiceConfig() *ServiceConfig {
  var config *ServiceConfig = &ServiceConfig{}

  log.Println("Loading service config")
  sysfile, err := os.Open("config.json")
  if err != nil {
    log.Fatalln("Error opening config.json:", err)
  }

  defer sysfile.Close()
  if err = json.NewDecoder(sysfile).Decode(config); err != nil {
    log.Fatalln("Error decoding config.json:", err)
  }
  return config
}