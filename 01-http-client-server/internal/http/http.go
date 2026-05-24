package http

type Request struct {
	Method  Method
	Headers Headers
	Body    string
	HasBody bool
}

type Method string

type Headers map[string]string

const (
	get    Method = "GET"
	delete Method = "DELETE"
	patch  Method = "PATCH"
	post   Method = "POST"
	put    Method = "PUT"
)

// constant for legal HTTP methods
var ValidMethods = []Method{
	get,
	delete,
	patch,
	post,
	put,
}
