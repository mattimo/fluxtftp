package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mattimo/fluxtftp"
	"io"
	"net"
)

type FluxClient struct {
	Conn net.Conn
}

func NewFluxClient(dial string) (*FluxClient, error) {
	f := &FluxClient{}

	addr, err := net.ResolveUnixAddr("unix", dial)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return nil, err
	}
	f.Conn = conn

	return f, nil
}

func (f *FluxClient) Add(reader io.Reader) error {

	buf := &bytes.Buffer{}
	buf.ReadFrom(reader)

	message, err := json.Marshal(&fluxtftp.ControlRequest{
		Verb: "Upload",
		Data: buf.Bytes(),
	})
	if err != nil {
		return err
	}

	_, err = f.Conn.Write(message)
	if err != nil {
		return err
	}

	answer := &fluxtftp.ControlResponse{}
	dec := json.NewDecoder(f.Conn)
	err = dec.Decode(answer)
	if err != nil {
		return err
	}

	if answer.Error != 0 {
		return fmt.Errorf("Server Error: %s", answer.Message)
	}

	return nil
}

func (f *FluxClient) Close() error {
	return f.Conn.Close()
}
