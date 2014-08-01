package fluxtftp

const (
	ControlErrOk        int64 = iota // Everything ok
	ControlErrUnknown                // error that we don't know
	ControlErrMalformed              // Maklformed request
	ControlErrSize                   // To large
)

type ControlRequest struct {
	Verb string
	Data []byte
}

type ControlResponse struct {
	Error   int64
	Message string
}
