package server

import (
	"github.com/mattimo/fluxtftp/tftp"
	"bytes"
	"io"
	"log"
	"net"
	"sync"
)

type FluxServer struct {
	sync.RWMutex
	Conf *Config
	mem  **bytes.Buffer
}

func NewFluxServer(config *Config) (*FluxServer, error) {
	flux := &FluxServer{}
	if config == nil {
		flux.Conf = &Config{}
	} else {
		flux.Conf = config
	}
	return flux, nil
}

func (flux *FluxServer) PutDefault(data []byte) error {
	buf := bytes.NewBuffer(data)
	flux.Lock()
	defer flux.Unlock()
	flux.mem = &buf
	return nil
}

func (flux *FluxServer) handleRequest(res *tftp.RRQresponse) {
	log.Println("Got request for", res.Request.Path)

	err := res.WriteOACK()
	if err != nil {
		log.Println("Failed to Write OACK:", err)
		return
	}

	reader, err := flux.GetFile(res.Request.Path)
	if err != nil {
		log.Println("Failed to open file:", err)
		if err == NoFileRegisteredErr {
			res.WriteError(tftp.NOT_FOUND, err.Error())
		} else {
			res.WriteError(tftp.UNKNOWN_ERROR, "Server Error")
		}
		return
	}

	b := make([]byte, res.Request.Blocksize)
	totalBytes := 0

	for {
		bytesRead, err := reader.Read(b)
		totalBytes += bytesRead

		if err == io.EOF {
			_, err := res.Write(b[:bytesRead])
			if err != nil {
				log.Println("Failed to write last bytes of the reader:", err)
			}
			res.End()
			break
		} else if err != nil {
			log.Println("Error while reading:", err)
			res.WriteError(tftp.UNKNOWN_ERROR, "Unknown Error")
			return
		}

		_, err = res.Write(b[:bytesRead])
		if err != nil {
			log.Println("Failed to write bytes:", err)
			return
		}
	}
}

func (flux *FluxServer) ListenAndServe() error {
	// TODO: eliminate this
	err := flux.fakeRegister("/home/iniuser/toPrint.ps")
	if err != nil {
		panic("that didn't work:" + err.Error())
	}

	log.Println("Starting server")

	// Check if Address is set, otherwise use default
	confAddr := flux.Conf.TftpAddress
	if confAddr == "" {
		confAddr = ":69"
	}
	addr, err := net.ResolveUDPAddr("udp", confAddr)
	if err != nil {
		return err
	}

	server, err := tftp.NewTFTPServer(addr)
	if err != nil {
		return err
	}

	for {
		res, err := server.Accept()
		if err != nil {
			log.Println("Bad Request:", err)
			continue
		}

		go flux.handleRequest(res)
	}

}
