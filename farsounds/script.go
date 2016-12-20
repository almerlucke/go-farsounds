package farsounds

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ScriptMainDescriptor describes main script content
type ScriptMainDescriptor struct {
	SampleRate    float64                `json:"sampleRate"`
	BufferLength  int32                  `json:"bufferLength"`
	PatchSettings map[string]interface{} `json:"patch"`
}

// UnmarshalFromFile unmarshal a JSON object from file
func UnmarshalFromFile(filePath string, obj interface{}) error {
	file, err := os.Open(filePath)

	if err != nil {
		return err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(obj)

	return err
}

// EvalInFileDirectory evaluate function in directory of file, change working directory
// if needed, and afterwards change it back to previous working directory
func EvalInFileDirectory(filePath string, eval func(basePath string) (interface{}, error)) (interface{}, error) {
	// Store current working directory
	oldWorkingDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Get absolute path of file
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	// Get directory file is located in
	newWorkingDirectory := filepath.Dir(absFilePath)

	// Strip file path to base
	basePath := filepath.Base(filePath)

	// If new directory is not the same as old directory
	// change working directory
	if newWorkingDirectory != oldWorkingDirectory {
		err = os.Chdir(newWorkingDirectory)
		if err != nil {
			return nil, err
		}

		// Restore old working directory
		defer os.Chdir(oldWorkingDirectory)
	}

	// Evaluate function with base path, old working directory is restored
	// after evaluation
	return eval(basePath)
}

// EvalScript loads json from script and calls eval function with unmarshalled json
func EvalScript(filePath string, eval func(obj interface{}) (interface{}, error)) (interface{}, error) {
	obj, err := EvalInFileDirectory(filePath, func(basePath string) (interface{}, error) {
		var script interface{}

		err := UnmarshalFromFile(basePath, &script)
		if err != nil {
			return nil, err
		}

		return script, nil
	})

	if err != nil {
		return nil, err
	}

	return eval(obj)
}

// LoadMainScript containing samplerate, bufferlength and main patch
func LoadMainScript(filePath string) (*Patch, error) {
	_patch, err := EvalInFileDirectory(filePath, func(basePath string) (interface{}, error) {
		mainDescriptor := ScriptMainDescriptor{}

		err := UnmarshalFromFile(basePath, &mainDescriptor)
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

// RenderScript load script and generate soundfile
func RenderScript(scriptPath string, soundFilePath string, numSeconds float64) error {
	// Load main script with sr, buflen and main patch
	patch, err := LoadMainScript(scriptPath)
	if err != nil {
		return err
	}

	// Always clean up patch
	defer patch.Cleanup()

	// Generate sound file from patch output
	return patch.Render(soundFilePath, numSeconds)
}
