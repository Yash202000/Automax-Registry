package services

import (
	"bytes"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/automax/registry/internal/models"
)

func validateName(fields map[string]string, field, value string, max int) string {
	value = strings.TrimSpace(value)
	switch {
	case value == "":
		fields[field] = field + " is required"
	case utf8.RuneCountInString(value) > max:
		fields[field] = field + " must be at most " + itoa(max) + " characters"
	}
	return value
}

func validateMetadata(fields map[string]string, raw models.JSON) models.JSON {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return models.JSON(`{}`)
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) > 64*1024 {
		fields["metadata"] = "metadata must be at most 65536 bytes"
		return raw
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &obj); err != nil || obj == nil {
		fields["metadata"] = "metadata must be a JSON object"
		return raw
	}
	return models.JSON(trimmed)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
