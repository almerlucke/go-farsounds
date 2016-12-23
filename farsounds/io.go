package farsounds

import (
	"math"
	"os"

	"github.com/mkb218/gosndfile/sndfile"
)

/*
   SoundWriter
*/

// SoundWriter writes samples to soundfile and normalizes them
type SoundWriter struct {
	// Temp output file
	*sndfile.File

	// Output format info
	Channels   int32
	Samplerate int32

	// Set to true if we want a normalized output
	// in that case first write to raw output file
	// and in the end normalize to final output file
	normalize bool

	// The temporary output file path
	tempOutputFilePath string

	// The final output file path
	finalOutputFilePath string

	// The peak value used for normalization
	peak float64
}

// OpenSoundWriter creates a new opened sound writer
func OpenSoundWriter(outputFilePath string, channels int32, samplerate int32, normalize bool) (*SoundWriter, error) {
	info := sndfile.Info{}
	info.Channels = channels
	info.Format = sndfile.SF_FORMAT_RAW | sndfile.SF_FORMAT_DOUBLE
	info.Samplerate = samplerate

	tempOutputFilePath := outputFilePath + ".raw"
	finalOutputFilePath := outputFilePath + ".aiff"

	os.Remove(tempOutputFilePath)
	os.Remove(finalOutputFilePath)

	tempFile, err := sndfile.Open(tempOutputFilePath, sndfile.ReadWrite, &info)
	if err != nil {
		return nil, err
	}

	w := SoundWriter{}
	w.Channels = channels
	w.Samplerate = samplerate
	w.normalize = normalize
	w.tempOutputFilePath = tempOutputFilePath
	w.finalOutputFilePath = finalOutputFilePath
	w.File = tempFile

	return &w, nil
}

// WriteSamples write raw samples to temp output
// keep track of peak value for normalization if needed
func (w *SoundWriter) WriteSamples(in []float64) error {
	_, err := w.WriteItems(in)

	if err != nil {
		return err
	}

	if w.normalize {
		for _, sample := range in {
			abs := math.Abs(sample)
			if abs > w.peak {
				w.peak = abs
			}
		}
	}

	return nil
}

// Close the sound writer
func (w *SoundWriter) Close() error {
	err := w.normalizeAndExport()

	if err != nil {
		w.File.Close()
		os.Remove(w.tempOutputFilePath)
		return err
	}

	err = w.File.Close()
	if err != nil {
		os.Remove(w.tempOutputFilePath)
		return err
	}

	err = os.Remove(w.tempOutputFilePath)

	return err
}

func (w *SoundWriter) normalizeAndExport() error {
	_, err := w.Seek(0, sndfile.Set)

	if err != nil {
		return err
	}

	outputInfo := sndfile.Info{}
	outputInfo.Channels = w.Channels
	outputInfo.Format = sndfile.SF_FORMAT_AIFF | sndfile.SF_FORMAT_DOUBLE
	outputInfo.Samplerate = w.Samplerate

	outputFile, err := sndfile.Open(w.finalOutputFilePath, sndfile.Write, &outputInfo)
	if err != nil {
		return err
	}

	normalizeValue := 1.0

	if w.normalize {
		normalizeValue = 1.0 / w.peak
	}

	sampleBlockSize := int64(2048)
	samples := make([]float64, sampleBlockSize)

	for {
		samplesToNormalize, err := w.ReadItems(samples[:])

		if err != nil {
			outputFile.Close()
			os.Remove(w.finalOutputFilePath)
			return err
		}

		// If we have no more samples to normalize then stop
		if samplesToNormalize == 0 {
			break
		}

		samplesSlice := samples[:samplesToNormalize]

		// Normalize samples
		for i := int64(0); i < samplesToNormalize; i++ {
			samplesSlice[i] *= normalizeValue
		}

		// Write samples to output
		outputFile.WriteItems(samplesSlice)
	}

	// Close output
	return outputFile.Close()
}

/*
   Sound buffer
*/

// SoundFileBuffer contains sound file deinterleaved samples
type SoundFileBuffer struct {
	// Deinterleaved channels
	Channels [][]float64

	// Sample rate
	SampleRate float64

	// Number of frames
	NumFrames int64

	// Duration in seconds
	Duration float64
}

// NewSoundFileBuffer load sound file from disk deinterleaved
func NewSoundFileBuffer(filePath string) (*SoundFileBuffer, error) {
	info := sndfile.Info{}

	file, err := sndfile.Open(filePath, sndfile.Read, &info)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	// Create one big buffer to hold all samples
	fileBuffer := make([]float64, int64(info.Channels)*info.Frames)

	// Create separate channels by splitting buffer into info.Channels parts
	channels := make([][]float64, info.Channels)
	for i := int32(0); i < info.Channels; i++ {
		channels[i] = fileBuffer[int64(i)*info.Frames : int64(i+1)*info.Frames]
	}

	// Deinterleave in blocks
	sampleBlockSize := int64(2048) * int64(info.Channels)
	samples := make([]float64, sampleBlockSize)
	frameIndex := int64(0)

	for {
		framesRead, err := file.ReadFrames(samples)
		if err != nil {
			return nil, err
		}

		if framesRead == 0 {
			break
		}

		for i := int64(0); i < framesRead; i++ {
			for j := int64(0); j < int64(info.Channels); j++ {
				channels[j][frameIndex+i] = samples[i*int64(info.Channels)+j]
			}
		}

		frameIndex += framesRead
	}

	buffer := SoundFileBuffer{}
	buffer.Duration = float64(info.Frames) / float64(info.Samplerate)
	buffer.NumFrames = info.Frames
	buffer.Channels = channels
	buffer.SampleRate = float64(info.Samplerate)

	return &buffer, nil
}
