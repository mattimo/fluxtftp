package main

import (
	"../client"
	"../server"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var daemon bool

const (
	SEARCHPATH  string = "/home/iniuser/.fluxftpd"
	TFTPDADDR   string = ":6969"
	CONTROLSOCK string = "/run/user/%d/fluxtftp/control"
)

var ControlAddr string

func init() {
	flag.BoolVar(&daemon, "daemon", false, "Start in Daemon mode")
	uid := os.Getuid()
	if uid == 0 {
		ControlAddr = "/var/run/fluxtftp/control"
		err := os.MkdirAll(filepath.Dir(ControlAddr), 0760)
		fmt.Println("Created control dir:", filepath.Dir(ControlAddr))
		if err != nil {
			panic("Could not create control sock " + filepath.Dir(ControlAddr))
		}
	} else {
		ControlAddr = fmt.Sprintf(CONTROLSOCK, uid)
		err := os.MkdirAll(filepath.Dir(ControlAddr), 0700)
		fmt.Println("Created control dir:", filepath.Dir(ControlAddr))
		if err != nil {
			panic("Could not create control sock " + filepath.Dir(ControlAddr))
		}
	}
}

func main() {
	flag.Parse()
	if daemon {
		fmt.Println("Starting fluxtftp daemon")
		fluxftpd, _ := server.NewFluxServer(
			&server.Config{
				SearchPath:  SEARCHPATH,
				TftpAddress: TFTPDADDR,
				ControlAddr: ControlAddr,
			})
		go func() {
			err := fluxftpd.StartControl()
			if err != nil {
				fmt.Println("Control Server Failed:", err)
			}
			return
		}()
		err := fluxftpd.ListenAndServe()
		if err != nil {
			fmt.Println("Server Error:", err)
		}
		return
	}

	fmt.Println("fluxtftp")
	client.Start()
}
