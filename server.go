package minihyperproxy

type Server interface {
	Serve()
	Stop()
	Info() *map[string]interface{}
}
