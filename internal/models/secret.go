package models

import (
	"database/sql/driver"
	"fmt"
	"io"
)

// SecretHash is a credential secret stored at rest. It implements driver.Valuer
// so the real hash is written to Postgres, and fmt.Formatter plus MarshalJSON
// so logs and API encoding cannot print it. REG-06 is the only place that
// returns the plaintext secret, and only once, at creation.
type SecretHash string

func (s SecretHash) String() string { return "[REDACTED]" }

func (s SecretHash) Format(f fmt.State, _ rune) {
	_, _ = io.WriteString(f, "[REDACTED]")
}

func (s SecretHash) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

func (s SecretHash) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	return string(s), nil
}

func (s *SecretHash) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*s = ""
		return nil
	case string:
		*s = SecretHash(v)
		return nil
	case []byte:
		*s = SecretHash(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into SecretHash", src)
	}
}
