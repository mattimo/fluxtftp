package server

import (
	"github.com/mattimo/fluxtftp"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type controlServer struct {
	listen *net.UnixListener
	flux   *FluxServer
}

func newControlServer(addr *net.UnixAddr) (*controlServer, error) {
	server := &controlServer{}
	l, err := net.ListenUnix("unix", addr)
	server.listen = l
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (c *controlServer) sendResponse(conn net.Conn, errCode int64, message string) error {
	resp := &fluxtftp.ControlResponse{Error: errCode, Message: message}
	wireResp, err := json.Marshal(resp)
	if err != nil {
		panic("Could nor marshal response")
	}
	_, err = conn.Write(wireResp)
	if err != nil {
		log.Printf("Could not write Response (val=%d): %s\n", errCode, err)
		return err
	}
	return nil
}

func (c *controlServer) handleControl(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	req := &fluxtftp.ControlRequest{}
	err := decoder.Decode(req)
	if err != nil || req.Verb == "" || len(req.Data) == 0 {
		if err != nil {
			log.Println("Control: Received:", req)
		}
		c.sendResponse(conn, fluxtftp.ControlErrMalformed, "Something went wrong")
		return
	}

	err = c.flux.PutDefault(req.Data)
	if err != nil {
		c.sendResponse(conn, fluxtftp.ControlErrUnknown, err.Error())
		return
	}
	err = c.sendResponse(conn, fluxtftp.ControlErrOk, "")
	if err != nil {
		return
	}
	log.Println("Control: Handled message from ", conn.RemoteAddr())

	return

}

func (c *controlServer) acceptControl() (net.Conn, error) {
	conn, err := c.listen.Accept()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func closeSocket(l *net.UnixListener) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		l.Close()
		os.Exit(0)
	}(sigc)
}

func (flux *FluxServer) StartControl() error {

	addr, err := net.ResolveUnixAddr("unix", flux.Conf.ControlAddr)
	if err != nil {
		return err
	}

	server, err := newControlServer(addr)
	if err != nil {
		return err
	}
	closeSocket(server.listen)
	server.flux = flux

	for {
		conn, err := server.acceptControl()
		if err != nil {
			log.Println("failed Control Request:", err)
			continue
		}

		go server.handleControl(conn)
	}
}
