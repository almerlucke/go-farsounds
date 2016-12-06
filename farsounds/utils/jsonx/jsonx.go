package jsonx

import (
	"encoding/json"
	"os"
)

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
