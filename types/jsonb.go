package types

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type Jsonb struct {
	json.RawMessage
}

func (j Jsonb) Valid() bool {
	_, err := j.RawMessage.MarshalJSON()
	return err == nil
}

func (j Jsonb) Value() (driver.Value, error) {
	if len(j.RawMessage) == 0 {
		return nil, nil
	}
	return j.MarshalJSON()
}

func (j *Jsonb) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, j)
}
