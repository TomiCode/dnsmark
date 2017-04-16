package main

import "fmt"
import "log"
import "bytes"
import "net/http"
import "encoding/json"

const cloudflare_endpoint = "https://api.cloudflare.com/client/v4"

type CloudflareClient struct {
  client *http.Client

  email string
  key string
}

type DNSRecord struct {
  Id      string  `json:"id"`
  Type    string  `json:"type"`
  Name    string  `json:"name"`
  Content string  `json:"content"`
  Ttl     int     `json:"ttl"`
  Locked  bool    `json:"locked"`
}

type DNSRecordList []DNSRecord

func (list DNSRecordList) Record(name string) (*DNSRecord, bool) {
  for _, record := range list {
    if record.Name == name {
      return &record, true
    }
  }
  return nil, false
}

func NewCloudflareClient(email, key string) CloudflareClient {
  return CloudflareClient{client: &http.Client{}, email: email, key: key}
}

func (client *CloudflareClient) ListDNSRecords(zone string) (DNSRecordList, bool) {
  req, err := http.NewRequest("GET", fmt.Sprintf("%s/zones/%s/dns_records", cloudflare_endpoint, zone), nil)
  if err != nil {
    log.Println("Error on NewRequest:", err)
    return nil, false
  }
  req.Header.Set("X-Auth-Email", client.email)
  req.Header.Set("X-Auth-Key", client.key)
  
  resp, err := client.client.Do(req)
  if err != nil {
    log.Println("Error occured:", err)
    return nil, false
  }
  defer resp.Body.Close()

  var json_resp struct {
    Success bool `json:"success"`
    Result DNSRecordList `json:"result"`
  }

  if err = json.NewDecoder(resp.Body).Decode(&json_resp); err != nil {
    log.Println("Error decoding response:", err)
    return nil, false
  }
  return json_resp.Result, json_resp.Success
}

func (client *CloudflareClient) UpdateDNSRecord(zone string, record *DNSRecord) bool {
  var json_req = struct {
    Type string `json:"type"`
    Name string `json:"name"`
    Content string `json:"content"`
  }{
    Type: record.Type,
    Name: record.Name,
    Content: record.Content,
  }

  var req_content bytes.Buffer
  if err := json.NewEncoder(&req_content).Encode(&json_req); err != nil {
    log.Println("Error encoding request:", err)
    return false
  }

  req, err := http.NewRequest("PUT", fmt.Sprintf("%s/zones/%s/dns_records/%s", cloudflare_endpoint, zone, record.Id), &req_content)
  if err != nil {
    log.Println("Error newrequest:", err)
    return false
  }
  req.Header.Set("X-Auth-Email", client.email)
  req.Header.Set("X-Auth-Key", client.key)

  resp, err := client.client.Do(req)
  if err != nil {
    log.Println("Error occured:", err)
    return false
  }
  defer resp.Body.Close()

  var json_resp struct {
    Success bool `json:"success"`
  }
  if err = json.NewDecoder(resp.Body).Decode(&json_resp); err != nil {
    log.Println("Error decoding response:", err)
    return false
  }
  return json_resp.Success
}