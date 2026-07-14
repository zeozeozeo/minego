package minego

import "testing"

func TestOfflineUUID(t *testing.T) {
	got := offlineUUID("Notch").String()
	if got != "b50ad385-829d-3141-a216-7e7d7539ba7f" {
		t.Fatalf("offline UUID = %s", got)
	}
}
