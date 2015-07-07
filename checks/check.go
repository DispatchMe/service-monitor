package checks

type Check interface {
	Run(serviceName string) error
	GetName() string
}
