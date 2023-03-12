package serializer

import (
	"encoding/json"
	"fmt"
)

type Serializer interface {
	Serialize(data any) ([]byte, error)
	Deserialize(bytes []byte, data any) error
}

var Json = &jsonSerializer{}

type jsonSerializer struct {
}

func (s *jsonSerializer) Serialize(data any) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("jsonSerializer::Serialize error: %v", err)
	}

	return bytes, nil
}

func (s *jsonSerializer) Deserialize(bytes []byte, data any) error {
	err := json.Unmarshal(bytes, data)
	if err != nil {
		return fmt.Errorf("jsonSerializer::Deserialize error: %w", err)
	}

	return nil
}
