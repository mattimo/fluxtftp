package server

import (
	"bytes"
	"github.com/mattimo/fluxtftp/tftp"
	"io"
	"log"
	"net"
	"sync"
)

type TftpReader interface {
	io.Reader
	io.ByteReader
}

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

type asciiReader struct {
	from TftpReader
	cr   bool
	null bool
}

func newAsciiReader(from TftpReader) *asciiReader {
	r := &asciiReader{
		from: from,
		cr:   false,
		null: false,
	}
	return r
}

func (r *asciiReader) ReadByte() (byte, error) {
	if r.cr {
		r.cr = false
		return 0x0a, nil
	}
	if r.null {
		r.null = false
		return 0x00, nil
	}
	b, err := r.from.ReadByte()
	if err != nil {
		return b, err
	}
	if b == 0x0d {
		r.null = true
		return b, nil
	}
	if b == 0x0a {
		r.cr = true
		b = 0x0d
	}
	return b, nil
}

func (r *asciiReader) Read(b []byte) (int, error) {
	inlen := cap(b)
	n := 0
	for n < inlen {
		cb, err := r.ReadByte()
		if err != nil {
			return n, err
		}
		b[n] = cb
		n++
	}
	return n, nil
}

func (flux *FluxServer) handleRequest(res *tftp.RRQresponse) {
	addr := *res.Request.Addr
	log.Printf("Got request for %s from %s/%s", res.Request.Path,
			addr.Network(),
			addr.String())

	err := res.WriteOACK()
	if err != nil {
		log.Println("Failed to Write OACK:", err)
		return
	}
	var reader TftpReader
	reader, err = flux.GetFile(res.Request.Path)
	if err != nil {
		log.Println("Failed to open file:", err)
		if err == NoFileRegisteredErr {
			res.WriteError(tftp.NOT_FOUND, err.Error())
		} else {
			res.WriteError(tftp.UNKNOWN_ERROR, "Server Error")
		}
		return
	}
	if res.Request.Mode == tftp.NETASCII {
		reader = newAsciiReader(reader)
	}

	b := make([]byte, res.Request.Blocksize)
	totalBytes := 0

	for {
		bytesRead, err := reader.Read(b)
		totalBytes = totalBytes + bytesRead

		if err == io.EOF {
			_, err = res.Write(b[:bytesRead])
			if err != nil {
				log.Println("Failed to write last bytes of the reader:", err)
				break
			}
			res.End()
			break
		} else if err != nil {
			log.Println("Error while reading:", err)
			res.WriteError(tftp.UNKNOWN_ERROR, "Unknown Error")
			break
		}

		_, err = res.Write(b[:bytesRead])
		if err != nil {
			log.Println("Failed to write bytes:", err)
			break
		}
	}
	return
}

func (flux *FluxServer) ListenAndServe() error {
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
