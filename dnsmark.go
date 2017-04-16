package main

import "log"
import "time"

type RouterInfo struct {
  NetClient
  RemoteAddr string
}

func (router *RouterInfo) Update(config *ServiceConfig) bool {
  if !router.Connect() {
    log.Fatalln("Cannot connect to remote host!")
  }
  defer router.Close()
  
  if !router.Auth(config.RouterConfig.User, config.RouterConfig.Pass) {
    log.Fatalln("Cannot auth to remote host!")
  }

  result, ok := router.Command("", []string{"externalIPAddress", "connectionStatus"})
  if !ok {
    log.Println("Telnet update command failed")
    return false
  }

  if connStatus, ok := result["connectionStatus"]; !ok || connStatus != "Connected" {
    log.Println("Connection status not valid")
    return false
  }

  if router.RemoteAddr, ok = result["externalIPAddress"]; !ok {
    log.Println("Cannot get externalIPAddress!")
    return false
  }
  return true
}

func main() {
  log.Println("Starting dnsmark service")
  config := LoadServiceConfig()
  router := &RouterInfo{NetClient: NewNetClient(config.RouterConfig.Addr)}

  cf_client := NewCloudflareClient(config.CloudflareConfig.Email, config.CloudflareConfig.Key)
  dns_map, ok := cf_client.ListDNSRecords(config.CloudflareConfig.ZoneId)
  if !ok {
    log.Fatalln("No dns record list :(")
  }

  ticker := time.NewTicker(time.Minute * 5)
  for {
    log.Println("Updating dns records")
    router.Update(config)

    for _, dns := range config.CloudflareConfig.Dns {
      entry, ok := dns_map[dns]
      if !ok {
        log.Println("No dns entry for:", dns)
        continue
      }
      log.Println(entry)
    }

    <-ticker.C
  }
}