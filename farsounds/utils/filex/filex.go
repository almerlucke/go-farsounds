package filex

import (
	"os"
	"path/filepath"
)

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
