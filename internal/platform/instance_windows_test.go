package platform

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSecondaryActivatesOriginalInstance(t *testing.T) {
	name := fmt.Sprintf(`Local\CodexLimitTray.Test.%d`, os.Getpid())
	opened := make(chan struct{}, 1)
	closeFirst, first, err := ownEvent(name, func() { opened <- struct{}{} })
	if err != nil || !first {
		t.Fatalf("first instance: %v, %v", first, err)
	}
	defer closeFirst()
	_, primary, err := ownEvent(name, func() {})
	if err != nil || primary {
		t.Fatalf("secondary instance: %v, %v", primary, err)
	}
	select {
	case <-opened:
	case <-time.After(3 * time.Second):
		t.Fatal("original instance did not receive activation")
	}
}
