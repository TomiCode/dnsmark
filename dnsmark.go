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

  result, ok := router.Command("wan show connection info pppoe_0_35_1_d", []string{"externalIPAddress", "connectionStatus"})
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

  log.Println("Router external IP:", router.RemoteAddr)
  return true
}

func main() {
  log.Println("Starting dnsmark service")
  config := LoadServiceConfig()
  router := &RouterInfo{NetClient: NewNetClient(config.RouterConfig.Addr)}

  cf_client := NewCloudflareClient(config.CloudflareConfig.Email, config.CloudflareConfig.Key)
  dns_list, ok := cf_client.ListDNSRecords(config.CloudflareConfig.ZoneId)
  if !ok {
    log.Fatalln("No dns record list :(")
  }

  ticker := time.NewTicker(time.Duration(config.UpdateSleep) * time.Minute)
  for {
    log.Println("Starting update task..")
    router.Update(config)
    
    for _, dns := range config.CloudflareConfig.Dns {
      entry, ok := dns_list.Record(dns)
      if !ok {
        log.Println("No dns entry for:", dns)
        continue
      }

      if dns_list[entry].Content != router.RemoteAddr {
        log.Println("Updating", dns_list[entry].Name, "from", dns_list[entry].Content, "to", router.RemoteAddr)
        dns_list[entry].Content = router.RemoteAddr
        if !cf_client.UpdateDNSRecord(config.CloudflareConfig.ZoneId, &dns_list[entry]) {
          log.Println("Update", dns_list[entry].Name, "failed!")
        }
      }
    }

    log.Println("Next update occurs after", config.UpdateSleep, "minutes")
    <-ticker.C
  }
}