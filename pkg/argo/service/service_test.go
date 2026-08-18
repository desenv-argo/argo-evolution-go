package argo_service

import "testing"

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name    string
		input   ApplicationInput
		wantErr bool
	}{
		{name: "valid", input: ApplicationInput{Slug: "argo-erp", Name: "Argo ERP", HealthURL: "https://erp.argo.app.br/health"}},
		{name: "invalid slug", input: ApplicationInput{Slug: "ERP!", Name: "ERP"}, wantErr: true},
		{name: "missing name", input: ApplicationInput{Slug: "argo-erp"}, wantErr: true},
		{name: "invalid health URL", input: ApplicationInput{Slug: "argo-erp", Name: "ERP", HealthURL: "javascript:alert(1)"}, wantErr: true},
		{name: "negative heartbeat", input: ApplicationInput{Slug: "argo-erp", Name: "ERP", ExpectedHeartbeatSeconds: -1}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateInput(test.input, true)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateInput() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestCredentialHashIsStableAndDoesNotExposeCredential(t *testing.T) {
	credential, hash, err := newCredential()
	if err != nil {
		t.Fatal(err)
	}
	if credential == "" || hash == "" {
		t.Fatal("credential and hash must be generated")
	}
	if credential == hash {
		t.Fatal("stored hash must not equal the plaintext credential")
	}
	if got := hashCredential(credential); got != hash {
		t.Fatalf("hashCredential() = %q, want %q", got, hash)
	}
}

func TestNormalizeHeartbeat(t *testing.T) {
	if got := normalizeHeartbeat(0); got != 300 {
		t.Fatalf("normalizeHeartbeat(0) = %d", got)
	}
	if got := normalizeHeartbeat(5); got != 30 {
		t.Fatalf("normalizeHeartbeat(5) = %d", got)
	}
	if got := normalizeHeartbeat(100000); got != 86400 {
		t.Fatalf("normalizeHeartbeat(100000) = %d", got)
	}
}
