/*
   Package goirc implements a basic IRC client. goirc is implemented
   around the Irc structure.

   An IRC configuration
   can be read from a JSON file; see the NewIrc function. For example,
   irc, err := goirc.NewIrc("config.json")

   goirc implements a number of commands that act on the 
*/
package goirc

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	IRC_READ_CHUNK = 4096
	IRC_READ_ALL   = 0
	IRC_TIMEOUT    = "3s"
)

var (
	pingRegex       = regexp.MustCompile("^PING :([\\w.]+)$")
	chanNames       = regexp.MustCompile("^:[\\w.]+ \\d+ \\w+ #?\\w+ :.+$")
	serverMsg       = regexp.MustCompile("^\\w+\\.\\w+\\.\\w+\\. .+$")
	pings     int64 = 0
)

type Irc struct {
	server    string   `json="server"`
	port      int      `json="port"`
	nick      string   `json="nick"`
	real      string   `json="real"`
	host      string   `json="host"`
	sys       string   `json="sys"`
	user      string   `json="user"`
	channels  []string `json="channels"`
	password  string   `json="password"`
	reconnect bool     `json="reconnect"`
	conn      *net.TCPConn
}

/*
   NewIrc creates a new Irc struct from a JSON configuration file. Required fields are:
        * server: specifies the IRC server to connect to
        * user: username on the server
        * nick: nickname to use on the server
        * real: real name
        * sys: system name
        * channels: a list of strings specifying channels to connect to
*/
func NewIrc(filename string) (cfg *Irc, err error) {
	var jsonByte []byte
	if jsonByte, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	err = json.Unmarshal(jsonByte, &cfg)
	if err != nil {
		fmt.Println("[!] error unmarshalling JSON: ", err.Error())
		return
	} else {
		if cfg.port == 0 {
			cfg.port = 6667
		}

		if cfg.real == "" {
			cfg.real = "GoKyle IRC client"
		}

		if cfg.server == "" || cfg.user == "" || cfg.nick == "" ||
			cfg.real == "" || cfg.sys == "" || len(cfg.channels) == 0 {
			err = fmt.Errorf("invalid configuration file")
		}
	}
	return cfg, err
}

// Connect carries out the IRC connection, including identification and
// channel joining.
func (irc *Irc) Connect() (connected bool, err error) {
	var ircServer *net.TCPAddr
	connected = false

	/* TODO: provide ipv6 support */
	if ircServer, err = net.ResolveTCPAddr("tcp4", irc.ConnStr()); err != nil {
		fmt.Printf("[!] couldn't connect to %s: %s\n", irc.ConnStr(),
			err.Error())
		return
	} else if irc.conn, err = net.DialTCP("tcp", nil, ircServer); err != nil {
		fmt.Printf("[!] couldn't dial out: %s\n", err.Error())
		return
	}

	if _, err = irc.Recv(IRC_READ_CHUNK, false); err != nil {
		fmt.Printf("[!] no response from server: %s\n", err.Error())
		return
	}

	fmt.Println("[+] sending nick")
	if err = irc.Send("NICK " + irc.nick); err != nil {
		fmt.Printf("[!] error setting nick: %s\n", err.Error())
		return
	}

	fmt.Println("[+] sending user")
	if err = irc.Send(irc.userline()); err != nil {
		fmt.Printf("[!] error setting user: %s\n", err.Error())
		return
	}

	if _, err = irc.Recv(IRC_READ_CHUNK, true); err != nil {
		fmt.Printf("[!] read error: %s\n", err.Error())
		return
	}

	fmt.Println("[+] identifying")
	if err = irc.identify(); err != nil {
		fmt.Printf("[!] error identifying: %s\n", err.Error())
		return
	}

	fmt.Println("[+] join channels:")
	for _, ch := range irc.channels {
		fmt.Printf("\t[*] %s\n", ch)
		if err = irc.Send("JOIN " + ch); err != nil {
			fmt.Printf("[!] error join channel %s: %s\n",
				ch, err.Error())
			return
		} else {
			irc.Msg(ch, "hello")
		}

	}
	connected = true
	return
}

// Send is a wrapper around the TCPConn.Write that adds proper line endings to strings.
func (irc *Irc) Send(msg string) (err error) {
	_, err = irc.conn.Write(ircbytes(msg))
	return
}

// Recv listens for incoming messages. Two constants have been provided for 
// use: IRC_READ_CHUNK should be used in almost all cases, as it listens for 
// a fixed size message; IRC_READ_ALL will read until the connection closes. 
// The block parameter, if false, will set a timeout on the socket (this 
// timeout can be changed by modifying the constant IRC_TIMEOUT), resetting 
// the socket to blocking mode after the receive.
func (irc *Irc) Recv(size int, block bool) (reply string, err error) {
	buf := make([]byte, size)

	if !block {
		dur, _ := time.ParseDuration(IRC_TIMEOUT)
		timeo := time.Now().Add(dur)
		irc.conn.SetDeadline(timeo)
	}

	if size == 0 {
		buf, err = ioutil.ReadAll(irc.conn)
	} else {
		_, err = irc.conn.Read(buf)
	}

	if err != nil {
		if err == io.EOF {
			fmt.Println("[+] connection closed.")
			irc.Quit(0)
		} else if err != nil && timeout(err) {
			fmt.Println("[-] timeout: resetting err")
			err = nil
		}
	}
	reply = strings.TrimFunc(string(buf), TrimReply)
	irc.conn.SetDeadline(time.Unix(0, 0))
	return
}

// ConnStr returns a host:port string from the Irc fields.
func (irc *Irc) ConnStr() string {
	return fmt.Sprintf("%s:%d", irc.server, irc.port)
}

// Disconnect sends a disconnect command to the IRC server.
func (irc *Irc) Disconnect() error {
	err := irc.Send("QUIT")
	if err != nil {
		fmt.Println("[!] disconnect error: ", err.Error())
		return err
	}
	return err
}

// Pong sends a ping reply to the server.
func (irc *Irc) Pong(daemon string) {
	pings++
	fmt.Printf("PONG #%d -> %s \n", pings, daemon)
	irc.Send("PONG " + daemon)
}

// Reply is a helper function to send a reply.
func (irc *Irc) Reply(sender string, message string) (err error) {
	err = irc.Msg(sender, message)
	return
}

// Message sends a PRIVMSG.
func (irc *Irc) Msg(to string, message string) (err error) {
	pm := fmt.Sprintf("PRIVMSG %s :%s", to, message)
	return irc.Send(pm)
}

// Quit disconnects and exits the IRC interface.
func (irc *Irc) Quit(ret int) {
	irc.Disconnect()
	os.Exit(ret)
}

// ircbytes converts the text to a byte slice with the standard IRC line
// endings.
func ircbytes(text string) (msg []byte) {
	msg = []byte(fmt.Sprintf("%s\r\n", text))
	return
}

// userline returns an appropriate USER line for connecting to an IRC
// server.
func (irc *Irc) userline() string {
	return fmt.Sprintf("USER %s %s %s %s", irc.user, irc.host,
		irc.sys, irc.real)
}

// timeout determines whether an error is a timeout
func timeout(err error) bool {
	timeoRe := regexp.MustCompile("i/o timeout")
	if timeoRe.MatchString(err.Error()) {
		return true
	}
	return false
}

// identify returns an identification string.
func (irc *Irc) identify() error {
	if irc.password == "" {
		return nil
	}
	ident := fmt.Sprintf("IDENTIFY %s %s", irc.user, irc.password)
	err := irc.Msg("NickServ", ident)
	return err
}

// TrimReply is a trim function for use with the reply returned from
// the socket read.
func TrimReply(r rune) bool {
	switch r {
	case ' ':
		return true
	case '\t':
		return true
	case '\n':
		return true
	case '\r':
		return true
	case '\x00':
		return true
	default:
		return false
	}
	return false
}
