package minimizer

import "log"

type Vector interface {
	AddScaled(float64, Vector)

	Negate()
	Scale(float64)

	Copy() Vector
	DotProduct(Vector) float64
}

type DifferentiableFunction interface {
	InitialWeights() Vector
	Value(weights Vector) (value float64)
	Gradient(weights Vector) (value float64, gradient Vector)
}

type MinimizerOptions struct {
	MinIterations, MaxIterations int
	Epsilon, Tolerance float64
}

var Standard = MinimizerOptions{MinIterations: 0, MaxIterations: 25, Epsilon: 1e-10, Tolerance: 1e-4}

type minimizer struct {
	opt MinimizerOptions
	fn DifferentiableFunction
	l *log.Logger

	iteration int
	point Vector

	pointHistory, gradientHistory []Vector

	lastValue, value float64
	gradient Vector
}

func start(opt MinimizerOptions, fn DifferentiableFunction, l *log.Logger) *minimizer {
	m := &minimizer{opt: opt, fn: fn, l: l, point: fn.InitialWeights()}
	m.value, m.gradient = fn.Gradient(m.point)

	return m
}

func (m *minimizer) finished() bool {
	done := m.iteration > m.opt.MinIterations

	if len(m.pointHistory) > 0 {
		change := (m.value - m.lastValue) / ((m.value + m.lastValue + m.opt.Epsilon) / 2.0)

		done = done && change < m.opt.Tolerance
	}

	return done || m.iteration >= m.opt.MaxIterations
}

func (m *minimizer) hessianScale() float64 {
	// FIXME: update this
	hessianScale := 1.0

	m.l.Printf("Found hessian scaling: %f", hessianScale)
	return hessianScale
}

func (m *minimizer) direction() Vector {
	direction := implicitMultiply(m.hessianScale(), m.gradient, m.gradientHistory, m.pointHistory)
	direction.Negate()

	m.l.Printf("Found direction")
	return direction
}

func (m *minimizer) stepSizeMultiplier() float64 {
	stepSize := 0.5
	// Breaking out from the first value requires smaller steps in most cases,
	// so don't overdo the step size
	if m.iteration == 0 {
		stepSize = 0.01
	}

	return stepSize
}

func (m *minimizer) iterate() {
	direction := m.direction()
	stepSizeMultiplier := m.stepSizeMultiplier()

	point := m.lineMinimize(direction, stepSizeMultiplier)
	m.l.Printf("Line minimization done")

	val, grad := m.fn.Gradient(point)

	m.pointHistory = append(m.pointHistory, m.point)
	m.gradientHistory = append(m.gradientHistory, m.gradient)

	m.value = val
	m.gradient = grad
	m.point = point
	m.iteration += 1
}

func GradientDescent(opt MinimizerOptions, fn DifferentiableFunction, l *log.Logger) Vector {
	l.Println("Starting gradient descent")

	var m *minimizer
	for m = start(opt, fn, l); !m.finished(); m.iterate() {
		l.Printf("Iteration %d", m.iteration)
	}

	return m.point
}

// FIXME: docs
func implicitMultiply(hessianScale float64, gradient Vector, gradientHistory, pointHistory []Vector) Vector {
	rho := []float64{}
	alpha := []float64{}
	right := gradient.Copy()

	for i := len(gradientHistory); i >= 0; i-- {
		pointDelta := pointHistory[i]
		derivativeDelta := gradientHistory[i]

		nextRho := pointDelta.DotProduct(derivativeDelta)
		if nextRho == 0.0 {
			panic("Curvature problem")
		}

		nextAlpha := pointDelta.DotProduct(right) / nextRho

		rho = append(rho, nextRho)
		alpha = append(alpha, nextAlpha)

		right.AddScaled(-nextAlpha, derivativeDelta)
	}

	left := right.Copy()
	left.Scale(hessianScale)

	for i := 0; i < len(gradientHistory); i++ {
		a := alpha[-i]
		r := rho[-i]
		pointDelta := pointHistory[i]
		derivativeDelta := gradientHistory[i]

		scale := a - derivativeDelta.DotProduct(left) / r
		left.AddScaled(scale, pointDelta)
	}

	return left
}

// Search along a line, defined by the direction from the current
// point, using stepSizeMultipler to control how far along the line we
// search. Returns either a new point, if a point is found with a
// lower value along the line, or the current point.
func (m *minimizer) lineMinimize(direction Vector, stepSizeMultiplier float64) Vector {
	stepSize := 1.0
	derivative := direction.DotProduct(m.gradient)

	m.l.Printf("Starting with step size %f", stepSize)
	for {
		guess := m.point.Copy()
		guess.AddScaled(stepSize, direction)

		guessValue := m.fn.Value(guess)
		sufficientDecreaseValue := m.value + m.opt.Tolerance * derivative * stepSize

		if guessValue <= sufficientDecreaseValue {
			m.l.Println("Line searcher found match")
			return guess
		}
		
		stepSize *= stepSizeMultiplier
		if stepSize < m.opt.Epsilon {
			m.l.Println("Line searcher underflow")
			return m.point
		}

		m.l.Printf("Retrying with step size %f", stepSize)
	}

	return m.point
}
