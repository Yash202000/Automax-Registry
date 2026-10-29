package utils

import "regexp"

var (
	bcryptHash   = regexp.MustCompile(`\$2[aby]\$\d{2}\$[./A-Za-z0-9]{53}`)
	secretAssign = regexp.MustCompile(`(?i)(api_secret_hash\s*[=:]\s*)(?:'[^']*'|"[^"]*"|\S+)`)
)

// Redact strips credential hashes from text before it is written to a log.
func Redact(s string) string {
	s = secretAssign.ReplaceAllString(s, "${1}'[REDACTED]'")
	return bcryptHash.ReplaceAllString(s, "[REDACTED]")
}
