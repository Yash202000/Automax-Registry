package database

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestRedactRemovesSecretHashes(t *testing.T) {
	bcryptHash := "$2a$10$" + strings.Repeat("a", 53)
	in := "INSERT api_secret_hash='" + bcryptHash + "' AND api_secret_hash=" + bcryptHash
	out := Redact(in)
	if strings.Contains(out, bcryptHash) {
		t.Fatalf("redact left the hash in %s", out)
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Fatalf("redact output = %s", out)
	}
}

func TestLogWriterRedactsSecretHash(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	hash := "$2b$12$" + strings.Repeat("b", 53)
	(redactingLogWriter{}).Printf("slow query api_secret_hash='%s'", hash)
	if strings.Contains(buf.String(), hash) {
		t.Fatalf("log leaked the hash: %s", buf.String())
	}
}

func TestSplitSQLKeepsQuotedSemicolons(t *testing.T) {
	parts := splitSQL("SELECT 'a;b';\nSELECT 1;")
	if len(parts) != 2 {
		t.Fatalf("parts = %#v", parts)
	}
	if parts[0] != "SELECT 'a;b'" || parts[1] != "SELECT 1" {
		t.Fatalf("parts = %#v", parts)
	}
}
