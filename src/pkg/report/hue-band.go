package report

import (
	"fmt"
	"hash/fnv"
)

/*
Categorical palette with wide-separated hue bands (12+ families).
Mapping is deterministic per task name, but ensures bands differ a lot.
*/
type hueBand struct {
	hMin float64
	hMax float64
	s    float64
	l    float64
}

var paletteBands = []hueBand{
	{0, 360, 0.00, 0.60},   // gray
	{210, 230, 0.70, 0.52}, // blue
	{35, 45, 0.85, 0.50},   // amber
	{95, 110, 0.70, 0.48},  // lime
	{270, 290, 0.60, 0.52}, // purple
	{310, 330, 0.65, 0.52}, // magenta
	{50, 60, 0.90, 0.46},   // yellow
	{335, 350, 0.65, 0.52}, // pink
	{170, 185, 0.65, 0.50}, // teal
	{120, 135, 0.65, 0.50}, // green
	{190, 205, 0.70, 0.48}, // cyan
	{235, 255, 0.65, 0.52}, // indigo
	{20, 30, 0.80, 0.50},   // orange
}

func taskColorHex(id int, task string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(task))
	hash := h.Sum32()

	bandIdx := int(uint32(id) % uint32(len(paletteBands)))
	band := paletteBands[bandIdx]

	// vary hue inside band
	inner := float64((hash>>8)%1000) / 1000.0 // 0..1
	hue := band.hMin + inner*(band.hMax-band.hMin)

	// vary lightness slightly across 3 steps
	lightSteps := []float64{-0.07, 0.0, +0.06}
	li := int((hash>>18)%uint32(len(lightSteps)))
	light := clamp01(band.l + lightSteps[li])

	r, g, b := hslToRGB(hue/360.0, band.s, light)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}


// HSL -> RGB helpers
func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		v := uint8(l * 255.0)
		return v, v, v
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	r := hue2rgb(p, q, h+1.0/3.0)
	g := hue2rgb(p, q, h)
	b := hue2rgb(p, q, h-1.0/3.0)
	return uint8(r*255.0 + 0.5), uint8(g*255.0 + 0.5), uint8(b*255.0 + 0.5)
}
func hue2rgb(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}
