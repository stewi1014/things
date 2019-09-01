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

func TestLdExp32(t *testing.T) {
	type args struct {
		sign     uint32
		exponent uint32
		fraction uint32
	}
	tests := []struct {
		name string
		args args
		want float32
	}{
		{
			name: "1",
			args: args{
				sign:     0,
				exponent: F32Bias,
				fraction: 1 << F32FractionBits,
			},
			want: 1,
		},
		{
			name: "0",
			args: args{
				sign:     0,
				exponent: F32Bias,
				fraction: 0,
			},
			want: 0,
		},
		{
			name: "negative",
			args: args{
				sign:     1,
				exponent: F32Bias,
				fraction: 1 << F32FractionBits,
			},
			want: -1,
		},
		{
			name: "become normal",
			args: args{
				sign:     0,
				exponent: 0,
				fraction: 5 << (F32FractionBits - 2),
			},
			want: BitsF32(0x00a00000),
		},
		{
			name: "become denormal",
			args: args{
				sign:     0,
				exponent: 1,
				fraction: 1 << (F32FractionBits - 2),
			},
			want: BitsF32(0x00200000),
		},
		{
			name: "shift left",
			args: args{
				sign:     0,
				exponent: F32Bias,
				fraction: 1 << (F32FractionBits - 2),
			},
			want: 0.25,
		},
		{
			name: "shift right",
			args: args{
				sign:     0,
				exponent: F32Bias,
				fraction: 1 << (F32FractionBits + 2),
			},
			want: 4,
		},
		{
			name: "infinity",
			args: args{
				sign:     0,
				exponent: F32MaxExponent - 1,
				fraction: 1 << (F32FractionBits + 2),
			},
			want: math32.Inf(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LdExp32(tt.args.sign, tt.args.exponent, tt.args.fraction); got != tt.want {
				t.Errorf("LdExp32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLdExp64(t *testing.T) {
	type args struct {
		sign     uint64
		exponent uint64
		fraction uint64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "1",
			args: args{
				sign:     0,
				exponent: F64Bias,
				fraction: 1 << F64FractionBits,
			},
			want: 1,
		},
		{
			name: "0",
			args: args{
				sign:     0,
				exponent: F64Bias,
				fraction: 0,
			},
			want: 0,
		},
		{
			name: "negative",
			args: args{
				sign:     1,
				exponent: F64Bias,
				fraction: 1 << F64FractionBits,
			},
			want: -1,
		},
		{
			name: "become normal",
			args: args{
				sign:     0,
				exponent: 0,
				fraction: 5 << (F64FractionBits - 2),
			},
			want: BitsF64(0x0014000000000000),
		},
		{
			name: "become denormal",
			args: args{
				sign:     0,
				exponent: 1,
				fraction: 1 << (F64FractionBits - 2),
			},
			want: BitsF64(0x0004000000000000),
		},
		{
			name: "shift left",
			args: args{
				sign:     0,
				exponent: F64Bias,
				fraction: 1 << (F64FractionBits - 2),
			},
			want: 0.25,
		},
		{
			name: "shift right",
			args: args{
				sign:     0,
				exponent: F64Bias,
				fraction: 1 << (F64FractionBits + 2),
			},
			want: 4,
		},
		{
			name: "infinity",
			args: args{
				sign:     0,
				exponent: F64MaxExponent - 1,
				fraction: 1 << (F64FractionBits + 2),
			},
			want: math.Inf(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LdExp64(tt.args.sign, tt.args.exponent, tt.args.fraction); got != tt.want {
				t.Errorf("LdExp32() = %v, want %v", got, tt.want)
			}
		})
	}
}
