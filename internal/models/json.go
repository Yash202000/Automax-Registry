package models

import (
	"database/sql/driver"
	"fmt"
)

// JSON is a jsonb value. It stores and returns raw JSON, not a base64 string.
type JSON []byte

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte(`{}`), nil
	}
	return []byte(j), nil
}

func (j *JSON) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*j = JSON(`{}`)
		return nil
	case []byte:
		*j = append(JSON(nil), v...)
		return nil
	case string:
		*j = JSON(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into JSON", src)
	}
}

func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte(`{}`), nil
	}
	return []byte(j), nil
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("nil JSON")
	}
	*j = append(JSON(nil), data...)
	return nil
}
