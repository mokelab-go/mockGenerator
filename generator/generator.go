package generator

// Generator is an interface
type Generator interface {
	Generate(src string) (string, error)
}
