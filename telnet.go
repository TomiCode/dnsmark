package main

import "net"
import "log"
import "fmt"
import "time"
import "regexp"

var cmdSuccess_exp = regexp.MustCompile("cmd:SUCC")

type NetClient struct {
  buffer []byte
  bytes  int
  conn   net.Conn
  addr   string
}

func NewNetClient(addr string) NetClient {
  return NetClient{buffer: make([]byte, 2048), addr: addr}
}

func (client *NetClient) Connect() bool {
  var err error

  log.Println("Connecting to", client.addr)
  client.conn, err = net.Dial("tcp", client.addr)
  if err != nil {
    log.Println(err)
    return false
  }
  return client.Handshake()
}

func (client *NetClient) Read() bool {
  var err error = nil

  client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
  if client.bytes, err = client.conn.Read(client.buffer); err != nil {
    log.Println(err)
    return false
  }
  client.buffer[client.bytes - 1] = 0x00
  return true
}

func (client *NetClient) Write(buf []byte) bool {
  if _, err := client.conn.Write(buf); err != nil {
    log.Println("Error at NetClient write:", err)
    return false
  }
  return true
}

func (client *NetClient) CommandSuccess() bool {
  log.Println("Checking command status")
  if cmdSuccess_exp.Match(client.buffer) {
    return true
  }

  buffer := make([]byte, 64)
  if _, err := client.conn.Read(buffer); err != nil {
    log.Println(err)
    return false
  }
  return cmdSuccess_exp.Match(buffer)
}

func (client *NetClient) Close() bool {
  if err := client.conn.Close(); err != nil {
    log.Println(err)
    return false
  }
  log.Println("Closed telnet connection")
  return true
}

func (client *NetClient) Handshake() bool {
  log.Println("Sending telnet handshake to remote", client.conn.RemoteAddr())
  client.Read()
  if !client.Write([]byte{0xff, 0xfe, 0x01, 0xff, 0xfb, 0x03, 0xff, 0xfd, 0x03}) {
    return false
  }
  client.Read()
  return true
}

func (client *NetClient) Auth(user, pass string) bool {
  log.Println("Authorizing telnet session as", user)
  if !client.Write([]byte(fmt.Sprintf("%s\r\n%s\r\n", user, pass))) {
    return false
  }

  client.Read()
  if !client.Write([]byte{'\r', '\n'}) {
    return false
  }

  // client.Read()
  if !client.Read() {
    return false
  }

  status, err := regexp.Match("TP-LINK\\(conf\\)#", client.buffer)
  if err != nil {
    log.Println(err)
    return false
  }
  log.Println("Telnet auth status:", status)
  return status
}

func (client *NetClient) Command(cmd string, args []string) (map[string]string, bool) {
  log.Println("Sending command", cmd)
  
  local_params := make(map[string]string)
  if !client.Write([]byte(fmt.Sprintf("%s\r\n", cmd))) {
    log.Println("Cannot send command to remote host")
    return nil, false
  }
  if !client.Read() {
    log.Println("Error occured while reading command output from remote")
    return nil, false
  }

  if !client.CommandSuccess() {
    log.Println("Command", cmd, "end without success on remote host")
    return nil, false
  }

  for _, field := range args {
    field_regexp, err := regexp.Compile(fmt.Sprintf("%s=(.*?)\r\n", field))
    if err != nil {
      log.Println("Error with arg", field, err)
      continue
    }

    match := field_regexp.FindSubmatch(client.buffer)
    if match == nil {
      log.Println("No match for argument", field)
      continue
    }
    local_params[field] = string(match[1])
  }
  return local_params, true
}