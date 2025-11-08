package queue

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestQueueMessage_Structure(t *testing.T) {
	// Test QueueMessage marshaling/unmarshaling
	msg := QueueMessage{
		Type:      "user_created",
		Payload:   []byte(`{"userId":"123","username":"test"}`),
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Marshal
	data, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.NotNil(t, data)

	// Unmarshal
	var unmarshaled QueueMessage
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, msg.Type, unmarshaled.Type)
	assert.Equal(t, msg.Payload, unmarshaled.Payload)
}

func TestQueueMessage_KratosPayload(t *testing.T) {
	// Test Kratos payload structure
	kratosPayload := struct {
		UserId   string `json:"userId"`
		Username string `json:"username"`
	}{
		UserId:   uuid.New().String(),
		Username: "testuser",
	}

	payloadBytes, err := json.Marshal(kratosPayload)
	assert.NoError(t, err)

	// Unmarshal back
	var unmarshaled struct {
		UserId   string `json:"userId"`
		Username string `json:"username"`
	}
	err = json.Unmarshal(payloadBytes, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, kratosPayload.UserId, unmarshaled.UserId)
	assert.Equal(t, kratosPayload.Username, unmarshaled.Username)
}

func TestQueueMessage_InvalidJSON(t *testing.T) {
	invalidJSON := `{"type":"user_created","payload":invalid}`

	var msg QueueMessage
	err := json.Unmarshal([]byte(invalidJSON), &msg)
	assert.Error(t, err)
}

func TestQueueMessage_TypeValidation(t *testing.T) {
	validTypes := []string{"user_created", "problem_updated", "contest_started"}

	for _, msgType := range validTypes {
		msg := QueueMessage{
			Type:      msgType,
			Payload:   []byte("{}"),
			CreatedAt: time.Now().Format(time.RFC3339),
		}

		assert.NotEmpty(t, msg.Type)
		assert.NotNil(t, msg.Payload)
	}
}

func TestFailedMessage_Structure(t *testing.T) {
	// Test failed message structure (used in DLQ)
	type FailedMessage struct {
		OriginalMessage string `json:"original_message"`
		Error           string `json:"error"`
		FailedAt        string `json:"failed_at"`
	}

	failed := FailedMessage{
		OriginalMessage: `{"type":"test"}`,
		Error:           "processing failed",
		FailedAt:        time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(failed)
	assert.NoError(t, err)
	assert.NotNil(t, data)

	var unmarshaled FailedMessage
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, failed.OriginalMessage, unmarshaled.OriginalMessage)
	assert.Equal(t, failed.Error, unmarshaled.Error)
}
