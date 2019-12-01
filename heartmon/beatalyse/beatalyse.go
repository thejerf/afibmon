package beatalyse

/*

I hope nobody ever tries to come after this legally. However, should a
patent troll sniff about: The algorithms in this code are heavily
influenced by "Signal Processing Methods for Heart Rate Variability"
by Gari D. Clifford at St. Cross College, a PhD thesis from 2002.

Any patent trolls sniffing around this file are well advised to
consider both A: the date of the thesis in question and the fact
that any patent based on it is by necessity either expired or
nearly so (not to mention even in the thesis these techniques are referred
to as past work, not something developed in this thesis) and B: that as
my signal processing experience is basically "I took an undergrad class in
the topic 20 years ago", that I have a fairly compelling case that I
can't possibly come up with an algorithm that is not virtually by
definition "obvious to one skilled in the art".

So, you know, think twice before advising your client to start getting
uppity about patents at me.

*/

import (
	"fmt"
	"io"

	"gonum.org/v1/gonum/fourier"
)

// SampleRate is how many times per second we get a sample from the EKG
// device.
const SampleRate = float64(30.0)

type BeatAnalyzer struct {
	size int
	fft  *fourier.FFT
}

type FFT struct {
	SampleRate   float64
	Coefficients []float64
	Frequencies  []float64
}

// DumpText dumps out the content of this FFT analysis in a format suitable
// for use by gnuplot.
func (fft FFT) DumpText(w io.Writer) error {
	var err error
	for idx, value := range fft.Coefficients {
		_, err = fmt.Fprintf(w, "%5.3f %5.3f\n",
			fft.Frequencies[idx],
			value,
		)
	}
	return err
}

func (fft FFT) Buckets(bucketCount int) Buckets {
	// Page 53 of the thesis suggests that the key data is between 5 and 15
	// Hz.

	buckets := make([]float64, bucketCount)
	bucketInterval := float64(10) / float64(bucketCount)

	for idx, freq := range fft.Frequencies {
		if freq < 5 || freq > 15 {
			continue
		}

		bucket := int((freq - 5) / bucketInterval)
		if bucket >= 0 && bucket < bucketCount {
			buckets[bucket] += fft.Coefficients[idx]
		}
	}

	return Buckets{
		Interval: bucketInterval,
		BottomHz: 5,
		Buckets:  buckets,
	}
}

type Buckets struct {
	Interval float64
	BottomHz float64
	Buckets  []float64
}

// Normalized returns all buckets divided by the first. The first bucket is
// consequently always one.
func (b Buckets) Normalized() []float64 {
	ret := make([]float64, len(b.Buckets))

	total := float64(0)
	for _, bucket := range b.Buckets {
		total += bucket
	}

	for idx, bucket := range b.Buckets {
		ret[idx] = (bucket / total) * float64(len(b.Buckets))
	}

	return ret
}

// New returns a new heartbeat analyzer. It should probably be called only
// with powers of two, though the fourier function doesn't say anything
// about that.
func New(i int) *BeatAnalyzer {
	return &BeatAnalyzer{
		i,
		fourier.NewFFT(i),
	}
}

// Analyze takes EKG data in the form of uint16s and does whatever it ends
// up doing once it ends up doing it. Panics if the length is not the same
// as it was created to be.
func (ba *BeatAnalyzer) FFT(ekg []uint16) FFT {
	if len(ekg) != ba.size {
		panic("Wrong length")
	}

	sequence := make([]float64, len(ekg))
	for idx, sample := range ekg {
		sequence[idx] = float64(sample)
	}

	coeffs := ba.fft.Coefficients(nil, sequence)
	frequencies := make([]float64, len(coeffs))

	// We don't care about phase information, so throw it away
	reals := make([]float64, len(coeffs))
	for idx, coeff := range coeffs {
		if real(coeff) < 0 {
			reals[idx] = -real(coeff)
		} else {
			reals[idx] = real(coeff)
		}
		frequencies[idx] = ba.fft.Freq(idx) * SampleRate
	}

	return FFT{SampleRate, reals, frequencies}
}
