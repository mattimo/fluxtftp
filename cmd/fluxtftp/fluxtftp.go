package main

import (
	"flag"
	"fmt"
	"github.com/coreos/go-systemd/activation"
	"github.com/mattimo/fluxtftp/client"
	"github.com/mattimo/fluxtftp/server"
	"os"
	"path/filepath"
)

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

	// Evaluate environment
	daemon_env := os.Getenv("FLUXTFTP_DAEMON")
	if daemon_env == "true" || daemon_env == "1" {
		daemon = true
	} else {
		daemon = false
	}
	tftpListen = os.Getenv("FLUXTFTP_TFTP_LISTEN")
	searchPath = os.Getenv("FLUXTFTP_SEARCH_PATH")

	// Evalueate Flags
	flag.Usage = usage
	const (
		daemonUsage = "Start Daemon"
		listenUsage = "tftp Port to listen on"
		pathUsage   = "file search path"
		hostUsage   = "fluxtftp host address/socket"
	)
	flag.BoolVar(&daemon, "daemon", daemon, daemonUsage)
	flag.BoolVar(&daemon, "d", daemon, daemonUsage+"(shorthand)")
	flag.StringVar(&tftpListen, "tftp-listen", tftpListen, listenUsage)
	flag.StringVar(&tftpListen, "t", tftpListen, listenUsage+"(shorthand)")
	flag.StringVar(&searchPath, "search-path", searchPath, pathUsage)
	flag.StringVar(&searchPath, "s", searchPath, pathUsage+"(shorthand)")
	flag.StringVar(&controlAddr, "host", controlAddr, hostUsage)
	flag.StringVar(&controlAddr, "H", controlAddr, hostUsage+"(shorthand)")
}

// init the server classically
func serverInitClassic() string {
	var control string
	uid := os.Getuid()
	if uid == 0 {
		control = CONTROLSOCK
		_, err := os.Stat(filepath.Dir(control))
		if err != nil {
			err := os.MkdirAll(filepath.Dir(control), 0760)
			fmt.Println("Created control dir:", filepath.Dir(control))
			if err != nil {
				panic("Could not create control sock " + filepath.Dir(control))
			}
		}
	} else {
		control = fmt.Sprintf(USERCONTROLSOCK, uid)
		_, err := os.Stat(filepath.Dir(control))
		if err != nil {
			err := os.MkdirAll(filepath.Dir(control), 0700)
			fmt.Println("Created control dir:", filepath.Dir(control))
			if err != nil {
				panic("Could not create control sock " + filepath.Dir(control))
			}
		}
	}
	return control
}

// use socketactivation return false if value could not be set
func serverInitSystemd() string {
	var control string
	files := activation.Files(true)
	if len(files) == 0 {
		return ""
	}
	// return only the last file name set
	for _, f := range files {
		info, err := f.Stat()
		if err != nil {
			continue
		}
		control = info.Name()
	}
	return control
}

func serverInit() {
	if controlAddr != "" {
		return
	}
	controlAddr = serverInitSystemd()
	if controlAddr == "" {
		controlAddr = serverInitClassic()
	}

}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s [-d] [filename]:\n", os.Args[0])
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
				SearchPath:  searchPath,
				TftpAddress: tftpListen,
				ControlAddr: controlAddr,
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
