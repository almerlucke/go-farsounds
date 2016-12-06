package farsounds

import "aliensareamongus.com/careibu/utils/jsonx"

// ScriptMainDescriptor describes main script content
type ScriptMainDescriptor struct {
	SampleRate    float64                `json:"sampleRate"`
	BufferLength  int32                  `json:"bufferLength"`
	PatchSettings map[string]interface{} `json:"patch"`
}

// LoadMainScript containing samplerate, bufferlength and main patch
func LoadMainScript(filePath string) (*Patch, error) {
	mainDescriptor := ScriptMainDescriptor{}

	err := jsonx.UnmarshalFromFile(filePath, &mainDescriptor)
	if err != nil {
		return nil, err
	}

	_patch, err := Registry.NewModule(
		"patch",
		"main",
		mainDescriptor.PatchSettings,
		mainDescriptor.BufferLength,
		mainDescriptor.SampleRate,
	)

	if err != nil {
		return nil, err
	}

	patch := _patch.(*Patch)

	return patch, nil
}
