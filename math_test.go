package things

import (
	"math"
	"testing"

	"github.com/chewxy/math32"
)

func TestMod(t *testing.T) {
	//format: x, divisor, modulo (correct result)
	testNumbers := []int{
		4, 5, 4,
		1, 1, 0,
		0, 1, 0,
		4, 3, 1,
		4, 4, 0,
		-1, 4, 3,
		-1, -4, 3,
		-2, 5, 3,
	}

	for i := 0; i < len(testNumbers); i += 3 {
		result := Mod(testNumbers[i], testNumbers[i+1])
		if result != testNumbers[i+2] {
			t.Errorf("Mod(%v, %v) = %v; wanted %v", testNumbers[i], testNumbers[i+1], result, testNumbers[i+2])
		}
	}
}

func TestModf(t *testing.T) {
	//format: x, divisor, modulo (correct result)
	testNumbers := []float64{
		4, 5, 4,
		1, 1, 0,
		0, 1, 0,
		4, 3, 1,
		4, 4, 0,
		-1, 4, 3,
		-1, -4, 3,
		-2, 5, 3,
		math.NaN(), 0, math.NaN(),
		0, math.NaN(), math.NaN(),
		math.MaxFloat64, math.Inf(1), math.MaxFloat64,
		-math.MaxFloat64, math.Inf(-1), math.MaxFloat64,
	}

	for i := 0; i < len(testNumbers); i += 3 {
		result := Modf(testNumbers[i], testNumbers[i+1])
		if !checkPassed(testNumbers[i+2], result) {
			t.Errorf("Modf(%v, %v) = %v; wanted %v", testNumbers[i], testNumbers[i+1], result, testNumbers[i+2])
		}
	}
}

func checkPassed(a, b float64) bool {
	if a == b {
		return true
	}
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return false
}

func TestModf32(t *testing.T) {
	//format: x, divisor, modulo (correct result)
	testNumbers := []float32{
		4, 5, 4,
		1, 1, 0,
		0, 1, 0,
		4, 3, 1,
		4, 4, 0,
		-1, 4, 3,
		-1, -4, 3,
		-2, 5, 3,
		math32.NaN(), 0, math32.NaN(),
		math32.NaN(), 0, math32.NaN(),
		0, math32.NaN(), math32.NaN(),
		math32.MaxFloat32, math32.Inf(1), math32.MaxFloat32,
		-math32.MaxFloat32, math32.Inf(-1), math32.MaxFloat32,
	}

	for i := 0; i < len(testNumbers); i += 3 {
		result := Modf32(testNumbers[i], testNumbers[i+1])
		if !checkPassed32(testNumbers[i+2], result) {
			t.Errorf("Modf(%v, %v) = %v; wanted %v", testNumbers[i], testNumbers[i+1], result, testNumbers[i+2])
		}
	}
}

func checkPassed32(a, b float32) bool {
	if a == b {
		return true
	}
	if math32.IsNaN(a) && math32.IsNaN(b) {
		return true
	}
	return false
}
