package things

import (
	"math"

	"github.com/chewxy/math32"
)

// Mod implements a mathematic modulo.
// The return is always positive, and it functions for negative values of i and mod
func Mod(x, divisor int) int {
	if divisor < 0 {
		divisor = -divisor
	}
	modulo := x - x/divisor*divisor
	if modulo < 0 {
		return modulo + divisor
	}
	return modulo
}

// Modf implements a mathematic modulo.
// The return is always positive, and it functions for negative values of i and mod
func Modf(x, divisor float64) float64 {
	if divisor == 0 || math.IsNaN(x) || math.IsNaN(divisor) {
		return math.NaN()
	}
	if math.IsInf(divisor, 0) {
		return math.Abs(x)
	}
	divisor = math.Abs(divisor)

	yfr, yexp := math.Frexp(divisor)
	r := x
	if x < 0 {
		r = -x
	}

	for r >= divisor {
		rfr, rexp := math.Frexp(r)
		if rfr < yfr {
			rexp = rexp - 1
		}
		r = r - math.Ldexp(divisor, rexp-yexp)
	}
	if x < 0 {
		return divisor - r
	}
	return r
}

// Modf32 implements a mathematic modulo.
// The return is always positive, and it functions for negative values of i and mod
func Modf32(x, divisor float32) float32 {
	if divisor == 0 || math32.IsNaN(x) || math32.IsNaN(divisor) {
		return math32.NaN()
	}
	if math32.IsInf(divisor, 0) {
		return math32.Abs(x)
	}
	divisor = math32.Abs(divisor)

	yfr, yexp := math32.Frexp(divisor)
	r := x
	if x < 0 {
		r = -x
	}

	for r >= divisor {
		rfr, rexp := math32.Frexp(r)
		if rfr < yfr {
			rexp = rexp - 1
		}
		r = r - math32.Ldexp(divisor, rexp-yexp)
	}
	if x < 0 {
		return divisor - r
	}
	return r
}
