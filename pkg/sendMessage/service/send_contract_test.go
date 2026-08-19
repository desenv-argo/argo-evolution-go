package send_service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMessageSendStructExposesInternalInstanceID(t *testing.T) {
	payload, err := json.Marshal(MessageSendStruct{InstanceID: "5df58306-3611-4b25-83a9-c40282250f57"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(payload), `"instanceId":"5df58306-3611-4b25-83a9-c40282250f57"`) {
		t.Fatalf("send payload does not expose internal instanceId: %s", payload)
	}
}
