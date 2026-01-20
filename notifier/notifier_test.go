package notifier

import (
	"testing"
)

type mockNotifier struct {
	messages []string
}

func (m *mockNotifier) Notify(message string) error {
	m.messages = append(m.messages, message)
	return nil
}

func TestManager_Notify(t *testing.T) {
	n1 := &mockNotifier{}
	n2 := &mockNotifier{}
	m := &Manager{
		notifiers: []Notifier{n1, n2},
	}

	testMsg := "test message"
	m.Notify(testMsg)

	if len(n1.messages) != 1 || n1.messages[0] != testMsg {
		t.Errorf("n1 did not receive correct message, got %v", n1.messages)
	}
	if len(n2.messages) != 1 || n2.messages[0] != testMsg {
		t.Errorf("n2 did not receive correct message, got %v", n2.messages)
	}
}
