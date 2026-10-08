package database

import (
	"fmt"
	"log"

	"github.com/automax/registry/pkg/utils"
)

// Redact strips credential hashes from text that might be written to a log.
func Redact(s string) string {
	return utils.Redact(s)
}

type redactingLogWriter struct{}

func (redactingLogWriter) Printf(format string, args ...any) {
	log.Print(Redact(fmt.Sprintf(format, args...)))
}
