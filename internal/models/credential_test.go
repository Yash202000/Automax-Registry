package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestDeviceCredentialNeverRevealsSecretHash(t *testing.T) {
	const secret = "plain-secret-value"
	cred := DeviceCredential{
		ID:            uuid.New(),
		ThingID:       uuid.New(),
		APIKey:        "device-key",
		APISecretHash: SecretHash(secret),
		Status:        "active",
	}

	body, err := json.Marshal(cred)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), secret) || strings.Contains(string(body), "api_secret_hash") {
		t.Fatalf("json exposed the hash: %s", body)
	}

	raw, err := SecretHash(secret).MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "null" || strings.Contains(string(raw), secret) {
		t.Fatalf("MarshalJSON = %s", raw)
	}

	rendered := []string{
		fmt.Sprintf("%v", cred),
		fmt.Sprintf("%+v", cred),
		fmt.Sprintf("%#v", cred),
		fmt.Sprintf("%s", cred.APISecretHash),
		fmt.Sprintf("%v", cred.APISecretHash),
		fmt.Sprintf("%+v", cred.APISecretHash),
		fmt.Sprintf("%#v", cred.APISecretHash),
		fmt.Sprintf("%q", cred.APISecretHash),
		cred.String(),
		cred.GoString(),
	}
	for _, line := range rendered {
		if strings.Contains(line, secret) {
			t.Fatalf("format leaked the hash: %s", line)
		}
	}
}
