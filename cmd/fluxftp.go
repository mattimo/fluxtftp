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
		ControlAddr = fmt.Sprintf(CONTROLSOCK, uid)
		err := os.MkdirAll(filepath.Dir(ControlAddr), 0700)
		fmt.Println("Created control dir:", filepath.Dir(ControlAddr))
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
	err = client.Add(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
