package argo_middleware

import "testing"

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		detail string
		want   string
	}{
		{name: "success", status: 200, want: ""},
		{name: "auth", status: 401, want: "AUTH_FAILED"},
		{name: "offline", status: 500, detail: "no active session found", want: "INSTANCE_OFFLINE"},
		{name: "invalid recipient", status: 400, detail: "invalid phone number", want: "INVALID_RECIPIENT"},
		{name: "timeout", status: 500, detail: "context deadline exceeded", want: "WHATSAPP_TIMEOUT"},
		{name: "media", status: 500, detail: "media upload failed", want: "MEDIA_FAILED"},
		{name: "validation", status: 400, detail: "message body is required", want: "VALIDATION_FAILED"},
		{name: "internal", status: 500, detail: "unexpected failure", want: "INTERNAL_ERROR"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := classifyError(test.status, test.detail); got != test.want {
				t.Fatalf("classifyError() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestResponseMessageID(t *testing.T) {
	body := []byte(`{"message":"success","data":{"Info":{"ID":"3EB0ABC"}}}`)
	if got := responseMessageID(body); got != "3EB0ABC" {
		t.Fatalf("responseMessageID() = %q, want %q", got, "3EB0ABC")
	}
}

func TestCleanHeader(t *testing.T) {
	if got := cleanHeader("  value\r\ninjected  ", 10); got != "value  inj" {
		t.Fatalf("cleanHeader() = %q", got)
	}
}

