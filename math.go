package things

import (
	"math"
	"math/bits"
	"unsafe"

	"github.com/chewxy/math32"
)

// Integer constants
const (
	UintBits = 32 << (^uint(0) >> 32 & 1) // 32 or 64 for
	IntBits  = UintBits

	MinUint8  = 0
	MaxUint8  = 1<<8 - 1
	MinUint16 = 0
	MaxUint16 = 1<<16 - 1
	MinUint32 = 0
	MaxUint32 = 1<<32 - 1
	MinUint64 = 0
	MaxUint64 = 1<<64 - 1

	MinInt8  = -1 << 7
	MaxInt8  = 1<<7 - 1
	MinInt16 = -1 << 15
	MaxInt16 = 1<<15 - 1
	MinInt32 = -1 << 31
	MaxInt32 = 1<<31 - 1
	MinInt64 = -1 << 63
	MaxInt64 = 1<<63 - 1

	MinUint = 0
	MaxUint = 1<<UintBits - 1
	MinInt  = -1 << (IntBits - 1)
	MaxInt  = 1<<(IntBits-1) - 1
)

// Float32 constants
const (
	F32FractionBits = 23
	F32ExponentBits = 8

	F32ExponentShift = F32FractionBits
	F32SignShift     = F32FractionBits + F32ExponentBits

	F32FractionMask = 1<<F32FractionBits - 1
	F32ExponentMask = (1<<F32ExponentBits - 1) << F32FractionBits
	F32SignMask     = 1 << 31

	F32ImpliedFractionBit = 1 << F32FractionBits

	F32MaxExponent = (1 << F32ExponentBits) - 1
	F32MinExponent = 0
	F32Bias        = (1 << (F32ExponentBits - 1)) - 1
)

// Float64 constants
const (
	F64FractionBits = 52
	F64ExponentBits = 11

	F64ExponentShift = F64FractionBits
	F64SignShift     = F64FractionBits + F64ExponentBits

	F64FractionMask = 1<<F64FractionBits - 1
	F64ExponentMask = (1<<F64ExponentBits - 1) << F64FractionBits
	F64SignMask     = 1 << 63

	F64ImpliedFractionBit = 1 << F64FractionBits

	F64MaxExponent = (1 << F64ExponentBits) - 1
	F64MinExponent = 0
	F64Bias        = (1 << (F64ExponentBits - 1)) - 1
)

// F64Bits converts a float64 to bits
func F64Bits(f float64) uint64 { return *(*uint64)(unsafe.Pointer(&f)) }

// F32Bits converts a float32 to bits
func F32Bits(f float32) uint32 { return *(*uint32)(unsafe.Pointer(&f)) }

// BitsF64 converts bits to a float64
func BitsF64(b uint64) float64 { return *(*float64)(unsafe.Pointer(&b)) }

// BitsF32 converts bits to a float32
func BitsF32(b uint32) float32 { return *(*float32)(unsafe.Pointer(&b)) }

// FrExp32 returns the sign, exponent and fractional component of a floating point number.
// It adds the implicit fractional bit if appropriate.
func FrExp32(f float32) (uint32, uint32, uint32) {
	fbits := F32Bits(f)
	exp := (fbits & F32ExponentMask) >> F32ExponentShift
	if exp == 0 {
		return (fbits & F32SignMask) >> F32SignShift,
			0,
			fbits & F32FractionMask
	}
	return (fbits & F32SignMask) >> F32SignShift,
		exp,
		(fbits & F32FractionMask) | F32ImpliedFractionBit
}

// LdExp32 assembles a float from sign, fraction and exponent.
// It expects the implicit fraction bit to be set.
func LdExp32(sign, exponent, fraction uint32) float32 {
	if fraction == 0 {
		return BitsF32(sign << F32SignShift)
	}

	shift := bits.LeadingZeros32(fraction) - F32ExponentBits // Add one for sign bit, but take one for implied fraction bit; so we just use ExponentBits
	if shift <= 0 {
		if exponent == 0 {
			exponent = 1
		}
		exponent += uint32(-shift)
	} else if shift >= int(exponent) {
		shift = int(exponent) - 1
		exponent = 0
	} else {
		exponent -= uint32(shift)
	}

	if exponent >= F32MaxExponent {
		exponent = F32MaxExponent
		fraction = 0
	} else if shift > 0 {
		fraction = fraction << uint(shift)
	} else if shift < 0 {
		fraction = fraction >> uint(-shift)
	}

	return BitsF32(sign<<F32SignShift | exponent<<F32ExponentShift | fraction&F32FractionMask)
}

// FrExp64 returns the sign, exponent and fractional component of a floating point number.
// It adds the implicit fractional bit if appropriate.
func FrExp64(f float64) (uint64, uint64, uint64) {
	fbits := F64Bits(f)
	exp := (fbits & F64ExponentMask) >> F64ExponentShift
	if exp == 0 {
		return (fbits & F64SignMask) >> F64SignShift,
			0,
			fbits & F64FractionMask
	}
	return (fbits & F64SignMask) >> F64SignShift,
		exp,
		(fbits & F64FractionMask) | F64ImpliedFractionBit
}

// LdExp64 assembles a float from sign, fraction and exponent.
// It expects the implicit fraction bit to be set.
func LdExp64(sign, exponent, fraction uint64) float64 {
	if fraction == 0 {
		return BitsF64(sign << F64SignShift)
	}

	shift := bits.LeadingZeros64(fraction) - F64ExponentBits // Add one for sign bit, but take one for implied fraction bit; so we just use ExponentBits
	if shift <= 0 {
		if exponent == 0 {
			exponent = 1
		}
		exponent += uint64(-shift)
	} else if shift >= int(exponent) {
		shift = int(exponent) - 1
		exponent = 0
	} else {
		exponent -= uint64(shift)
	}

	if exponent >= F64MaxExponent {
		exponent = F64MaxExponent
		fraction = 0
	} else if shift > 0 {
		fraction = fraction << uint(shift)
	} else if shift < 0 {
		fraction = fraction >> uint(-shift)
	}

	return BitsF64(sign<<F64SignShift | exponent<<F64ExponentShift | fraction&F64FractionMask)
}

// Mod implements a mathematic modulo.
// The return is always positive, and it functions for negative values of x and divisor
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
// The return is always positive, and it functions for negative values of x and divisor
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
// The return is always positive, and it functions for negative values of x and divisor
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
