package frame

import "math"

// Requantization exponents are exact multiples of 1/4. Build with the same
// math.Pow operation as upstream, not Exp2 or a differently rounded identity.
// The table is immutable after initialization and never holds decoder state.
var quarterGains = func() (table [577]float64) {
	for i := range table {
		table[i] = math.Pow(2, float64(i-512)/4)
	}
	return
}()

func requantizeSample(samples *[576]float32, at int, gain float64) {
	v := samples[at]
	if v < 0 {
		samples[at] = float32(gain * -powtab34[int(-v)])
	} else {
		samples[at] = float32(gain * powtab34[int(v)])
	}
}

func (f *Frame) requantizeShortWindow(gr, ch, start, length, band, window int) {
	gain := f.requantizeGainShort(gr, ch, band, window)
	for i := start; i < start+length; i++ {
		requantizeSample(&f.mainData.Is[gr][ch], i, gain)
	}
}

func requantizeGain(exponent float64) float64 {
	q := exponent * 4
	i := int(q)
	if q == float64(i) && i >= -512 && i <= 64 {
		return quarterGains[i+512]
	}
	return math.Pow(2, exponent)
}
