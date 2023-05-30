package main

import (
	"C"
	"context"
	"encoding/json"
	"fmt"
	"github.com/things-go/go-socks5"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os"
	"time"
)

type TunnelResponse struct {
	IsSuccess bool
	Message   string
}

var sshConn *ssh.Client

func main() {
}

//export OpenTunnel
func OpenTunnel(sshAddress *C.char, socks5Address *C.char, user *C.char, password *C.char) *C.char {
	var isSuccess = true
	var message string

	sshConf := &ssh.ClientConfig{
		User:            C.GoString(user),
		Auth:            []ssh.AuthMethod{ssh.Password(C.GoString(password))},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         60 * time.Second,
	}

	var err error
	sshConn, err = ssh.Dial("tcp", C.GoString(sshAddress), sshConf)

	if err != nil {
		fmt.Println("error tunnel to server: ", err)
		message = err.Error()
		isSuccess = false
	}

	fmt.Println("connected to ssh server")

	go func() {
		fmt.Println("start server: " + C.GoString(socks5Address))

		server := socks5.NewServer(
			socks5.WithDial(func(ctx context.Context, network, addr string) (net.Conn, error) {
				return sshConn.Dial(network, addr)
			}),
			socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags))),
		)

		// Create SOCKS5 proxy on localhost port 7201
		if err := server.ListenAndServe("tcp", C.GoString(socks5Address)); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println("finish")
	response := TunnelResponse{
		IsSuccess: isSuccess, Message: message,
	}

	jsonBytes, _ := json.Marshal(response)
	cString := C.CString(string(jsonBytes))
	return cString
}

//export CloseTunnel
func CloseTunnel() {
	if sshConn != nil {
		sshConn.Close()
	}
}
