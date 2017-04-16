package main

import "log"

func main() {
  log.Println("Starting dnsmark service")
  config := LoadServiceConfig()

  log.Println(config)
}