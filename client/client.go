package client

import (
	"github.com/mattimo/fluxtftp"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

func Add(reader io.Reader) error {
	addr, err := net.ResolveUnixAddr("unix", "/run/user/1000/fluxtftp/control")
	if err != nil {
		return err
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	buf.ReadFrom(reader)

	message, err := json.Marshal(&fluxtftp.ControlRequest{
		Verb: "Upload",
		Data: buf.Bytes(),
	})
	if err != nil {
		return err
	}

	_, err = conn.Write(message)
	if err != nil {
		return err
	}

	answer := &fluxtftp.ControlResponse{}
	dec := json.NewDecoder(conn)
	err = dec.Decode(answer)
	if err != nil {
		return err
	}

	if answer.Error != 0 {
		return fmt.Errorf("Server Error: %s", answer.Message)
	}

	return nil
}
