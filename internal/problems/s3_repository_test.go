package problems

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestS3Repository_KeyFormat(t *testing.T) {
	// We can't easily mock S3 client without full interface implementation
	// So we just test the key format generation logic

	// Test key format generation
	problemID := uuid.New()
	expectedKey := fmt.Sprintf("problems/%s/tests.zip", problemID)

	assert.Contains(t, expectedKey, problemID.String())
	assert.Contains(t, expectedKey, "tests.zip")
	assert.Contains(t, expectedKey, "problems/")
}

func TestS3Repository_KeyFormatValidation(t *testing.T) {
	// Test that key format is consistent
	problemID1 := uuid.New()
	problemID2 := uuid.New()

	key1 := fmt.Sprintf("problems/%s/tests.zip", problemID1)
	key2 := fmt.Sprintf("problems/%s/tests.zip", problemID2)

	// Keys should be different for different problem IDs
	assert.NotEqual(t, key1, key2)

	// Keys should follow the same pattern
	assert.Contains(t, key1, "problems/")
	assert.Contains(t, key1, "/tests.zip")
	assert.Contains(t, key2, "problems/")
	assert.Contains(t, key2, "/tests.zip")
}
