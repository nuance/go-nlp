package gnlp

type Counter interface {
	Get(string) float64
	Set(string, float64)
	Incr(string)

	Log()
	Exp()
	Normalize()
	LogNormalize()

	Apply(op func(*string, float64) float64)
}
