package main

import (
	"../client"
	"../server"
	"flag"
	"fmt"
)

var daemon bool

func init() {
	flag.BoolVar(&daemon, "daemon", false, "Start in Daemon mode")
}

func main() {
	flag.Parse()
	if daemon {
		fmt.Println("Starting fluxtftp daemon")
		fluxftpd, _ := server.NewFluxServer(
			&server.Config{
				SearchPath:  "/home/iniuser/.fluxftpd",
				TftpAddress: ":6969",
				ControlAddr: "/run/user/1000/fluxtftpd.sock",
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
