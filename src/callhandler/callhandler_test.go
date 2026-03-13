package callhandler

import (
	"testing"
)

func TestCallHandler(t *testing.T) {
	go InitCallHandler()
	select {}
}
