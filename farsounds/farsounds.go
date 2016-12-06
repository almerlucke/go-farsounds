package farsounds

import (
	"errors"

	"github.com/almerlucke/go-farsounds/farsounds/io"
	"github.com/almerlucke/go-farsounds/farsounds/utils/filex"
	"github.com/almerlucke/go-farsounds/farsounds/utils/jsonx"
)

// ScriptMainDescriptor describes main script content
type ScriptMainDescriptor struct {
	SampleRate    float64                `json:"sampleRate"`
	BufferLength  int32                  `json:"bufferLength"`
	PatchSettings map[string]interface{} `json:"patch"`
}

// LoadMainScript containing samplerate, bufferlength and main patch
func LoadMainScript(filePath string) (*Patch, error) {
	_patch, err := filex.EvalInFileDirectory(filePath, func(basePath string) (interface{}, error) {
		mainDescriptor := ScriptMainDescriptor{}

		err := jsonx.UnmarshalFromFile(basePath, &mainDescriptor)
		if err != nil {
			return nil, err
		}

		return Registry.NewModule(
			"patch",
			"main",
			mainDescriptor.PatchSettings,
			mainDescriptor.BufferLength,
			mainDescriptor.SampleRate,
		)
	})

	if err != nil {
		return nil, err
	}

	return _patch.(*Patch), nil
}

// SoundFileFromPatch generate soundfile from patch
func SoundFileFromPatch(patch *Patch, soundFilePath string, numSeconds float64) error {
	// Store in local vars
	sr := patch.GetSampleRate()
	buflen := patch.GetBufferLength()
	numChannels := int32(len(patch.GetOutlets()))

	// Sanity check on numChannels
	if numChannels < 1 || numChannels > 2 {
		return errors.New("Main patch must have one or two outputs")
	}

	// Open sound writer
	writer, err := io.OpenSoundWriter(soundFilePath, numChannels, int32(sr), true)
	if err != nil {
		return err
	}

	// Always clean up writer
	defer writer.Close()

	// Prepare DSP loop
	timestamp := int64(0)
	numCycles := int64(((numSeconds * sr) / float64(buflen)) + 0.5)
	sampleBuffer := make([]float64, numChannels*buflen)

	// Generate samples for N cycles
	for i := int64(0); i < numCycles; i++ {
		patch.PrepareDSP()
		patch.DSP(timestamp)

		// Interleave channels
		for j := int32(0); j < buflen; j++ {
			for c := int32(0); c < numChannels; c++ {
				sampleBuffer[j*numChannels+c] = patch.Outlets[c].Buffer[j]
			}
		}

		// Write samples
		err = writer.WriteSamples(sampleBuffer)
		if err != nil {
			return err
		}

		// Increase timestamp
		timestamp += int64(buflen)
	}

	// No errors
	return nil
}

// SoundFileFromScript load script and generate soundfile
func SoundFileFromScript(scriptPath string, soundFilePath string, numSeconds float64) error {
	// Load main script with sr, buflen and main patch
	patch, err := LoadMainScript(scriptPath)
	if err != nil {
		return err
	}

	// Always clean up patch
	defer patch.Cleanup()

	// Generate sound file from patch output
	return SoundFileFromPatch(patch, soundFilePath, numSeconds)
}
