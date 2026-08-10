package task

import (
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

func TestNewSendMessageTaskAndParse(t *testing.T) {
	p := SendMessagePayload{
		SenderID: 1, SenderType: "admin", RecipientIDs: []uint{2, 3},
		Title: "通知", Content: "你好",
	}
	task, err := NewSendMessageTask(p)
	if err != nil {
		t.Fatal(err)
	}
	if task.Type() != TypeSendMessage {
		t.Errorf("type=%s want %s", task.Type(), TypeSendMessage)
	}
	got, err := ParseSendMessagePayload(task)
	if err != nil {
		t.Fatal(err)
	}
	if got.SenderID != 1 || len(got.RecipientIDs) != 2 || got.Title != "通知" || got.Content != "你好" {
		t.Errorf("parsed mismatch: %+v", got)
	}
}

func TestParseCorruptPayload(t *testing.T) {
	task := asynq.NewTask(TypeSendMessage, []byte("{not-json"))
	if _, err := ParseSendMessagePayload(task); err == nil {
		t.Fatal("want error for corrupt payload")
	}
}

func TestDefaultRetention(t *testing.T) {
	if defaultRetention != 7*24*time.Hour {
		t.Errorf("defaultRetention=%v want 7d", defaultRetention)
	}
}
