package audio

import (
	"encoding/binary"
	"math"
	"math/rand"
)

const sampleRate = 44100

func makeWAV(samples []int16) []byte {
	numChannels := 1
	bitsPerSample := 16
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8
	dataSize := len(samples) * 2

	buf := make([]byte, 44+dataSize)
	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], uint32(36+dataSize))
	copy(buf[8:12], "WAVE")
	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16)
	binary.LittleEndian.PutUint16(buf[20:22], 1) // PCM
	binary.LittleEndian.PutUint16(buf[22:24], uint16(numChannels))
	binary.LittleEndian.PutUint32(buf[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(buf[28:32], uint32(byteRate))
	binary.LittleEndian.PutUint16(buf[32:34], uint16(blockAlign))
	binary.LittleEndian.PutUint16(buf[34:36], uint16(bitsPerSample))
	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataSize))
	for i, s := range samples {
		binary.LittleEndian.PutUint16(buf[44+i*2:46+i*2], uint16(s))
	}
	return buf
}

func shatterSamples() []int16 {
	n := sampleRate * 3 / 10
	out := make([]int16, n)
	r := rand.New(rand.NewSource(1))
	for i := range out {
		t := float64(i) / float64(sampleRate)
		env := math.Exp(-t * 10)
		noise := r.Float64()*2 - 1
		val := noise * env * 0.7
		out[i] = int16(val * 32767)
	}
	return out
}

func serveSamples() []int16 {
	n := sampleRate / 8
	out := make([]int16, n)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		freq := 300.0
		env := math.Exp(-t * 15)
		v := math.Sin(2 * math.Pi * freq * t)
		out[i] = int16(v * env * 0.25 * 32767)
	}
	return out
}

func catchSamples() []int16 {
	n := sampleRate / 6
	out := make([]int16, n)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		freq := 660.0
		env := math.Exp(-t * 8)
		v := math.Sin(2 * math.Pi * freq * t)
		out[i] = int16(v * env * 0.3 * 32767)
	}
	return out
}

func powerUpSamples() []int16 {
	n := sampleRate / 4
	out := make([]int16, n)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		freq := 800 + 400*math.Sin(t*10)
		env := math.Exp(-t * 6)
		v := math.Sin(2 * math.Pi * freq * t)
		out[i] = int16(v * env * 0.3 * 32767)
	}
	return out
}

func excitedSamples() []int16 {
	n := sampleRate / 5
	out := make([]int16, n)
	for i := range out {
		t := float64(i) / float64(sampleRate)
		freq := 500 + 1500*t
		val := math.Sin(2*math.Pi*freq*t) * (1 - t*4) * 0.5
		out[i] = int16(val * 32767)
	}
	return out
}

func successSamples() []int16 {
	n := sampleRate * 6 / 5
	out := make([]int16, n)
	r := rand.New(rand.NewSource(2))
	// firework pops
	for burst := 0; burst < 3; burst++ {
		start := burst * sampleRate / 4
		for i := 0; i < sampleRate/8; i++ {
			idx := start + i
			if idx >= len(out) {
				break
			}
			t := float64(i) / float64(sampleRate)
			env := math.Exp(-t * 25)
			noise := r.Float64()*2 - 1
			out[idx] += int16(noise * env * 0.5 * 32767)
		}
	}
	// fanfare notes
	notes := []struct {
		freq  float64
		start float64
		dur   float64
	}{
		{523.25, 0.05, 0.12},
		{659.25, 0.18, 0.12},
		{783.99, 0.31, 0.12},
		{1046.50, 0.44, 0.30},
	}
	for _, nt := range notes {
		for i := range out {
			t := float64(i) / float64(sampleRate)
			if t < nt.start || t > nt.start+nt.dur {
				continue
			}
			lt := t - nt.start
			env := math.Exp(-lt * 4)
			v := math.Sin(2 * math.Pi * nt.freq * t)
			out[i] += int16(v * env * 0.4 * 32767)
		}
	}
	// sparkle tail
	for i := range out {
		t := float64(i) / float64(sampleRate)
		if t < 0.44 {
			continue
		}
		env := math.Exp(-(t - 0.44) * 3)
		noise := r.Float64()*2 - 1
		out[i] += int16(noise * env * 0.15 * 32767)
	}
	return out
}
