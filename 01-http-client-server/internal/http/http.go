package http

type Request struct {
	Method  Method
	Headers Headers
	Body    string
	HasBody bool
}

type Method string

type Headers map[string]string
