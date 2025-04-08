package blueprint

type CoreSystem interface {
	Run(scene Scene, deltaTime float64) error
}
