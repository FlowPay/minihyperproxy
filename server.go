package minihyperproxy

type Server interface {
	Serve()
	Stop()
	Type() string
}
