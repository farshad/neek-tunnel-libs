package main

import (
	"C"
	"context"
	"encoding/json"
	"fmt"
	"github.com/armon/go-socks5"
	"golang.org/x/crypto/ssh"
	"net"
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
		Timeout:         30,
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
		fmt.Println("start server")
		socksConf := &socks5.Config{
			Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return sshConn.Dial(network, addr)
			},
		}

		serverSocks, err := socks5.New(socksConf)
		if err != nil {
			fmt.Println(err)
			message = err.Error()
			isSuccess = false
		}

		if err := serverSocks.ListenAndServe("tcp", C.GoString(socks5Address)); err != nil {
			fmt.Println("failed to create socks5 server", err)
			message = err.Error()
			isSuccess = false
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
