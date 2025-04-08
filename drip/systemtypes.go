package drip

type ServerSystem interface {
	Run(Server) error
}
