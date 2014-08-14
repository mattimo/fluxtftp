package main

import (
	"github.com/mattimo/fluxtftp/client"
	"github.com/mattimo/fluxtftp/server"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var daemon bool

const (
	SEARCHPATH      string = "/var/lib/tftpd/"
	TFTPDADDR       string = ":69"
	USERCONTROLSOCK string = "/var/run/user/%d/fluxtftp/control"
	CONTROLSOCK     string = "/var/run/fluxtftp/control"
)

var controlAddr string
var daemon bool = false
var tftpListen string = TFTPDADDR
var searchPath string = SEARCHPATH

func init() {
	flag.Usage = usage
	const (
		daemonUsage = "Start in Daemon mode"
	)
	flag.BoolVar(&daemon, "daemon", false, daemonUsage)
	flag.BoolVar(&daemon, "d", false, daemonUsage+"(shorthand)")
}

func serverInit() {
	uid := os.Getuid()
	if uid == 0 {
		ControlAddr = "/var/run/fluxtftp/control"
		err := os.MkdirAll(filepath.Dir(ControlAddr), 0760)
		fmt.Println("Created control dir:", filepath.Dir(ControlAddr))
		if err != nil {
			panic("Could not create control sock " + filepath.Dir(ControlAddr))
		}
	} else {
		control = fmt.Sprintf(USERCONTROLSOCK, uid)
		_, err := os.Stat(filepath.Dir(control))
		if err != nil {
			panic("Could not create control sock " + filepath.Dir(ControlAddr))
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s [filename]:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, " filename: name of the file to be served\n")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if daemon {
		fmt.Println("Starting fluxtftp daemon")
		serverInit()
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

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	file, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	// create client
	c, err := client.NewFluxClient(controlAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer c.Close()
	err = c.Add(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
