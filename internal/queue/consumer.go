package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gate149/core/internal/models"
	"github.com/gate149/core/internal/users"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	redisClient *redis.Client
	usersUC     *users.UsersUseCase
}

type QueueMessage struct {
	Type      string `json:"type"`
	Payload   []byte `json:"payload"`
	CreatedAt string `json:"created_at"`
}

func NewConsumer(redisClient *redis.Client, usersUC *users.UsersUseCase) *Consumer {
	return &Consumer{
		redisClient: redisClient,
		usersUC:     usersUC,
	}
}

func (c *Consumer) StartConsuming(ctx context.Context, queueName string) {
	log.Printf("Starting to consume queue: %s", queueName)

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping queue consumer")
			return
		default:
			// Blocking pop from Redis queue
			result, err := c.redisClient.BLPop(ctx, 1*time.Second, queueName).Result()
			if err != nil {
				if err == redis.Nil {
					// No messages in queue, continue
					continue
				}
				log.Printf("Error reading from queue: %v", err)
				continue
			}

			if len(result) < 2 {
				log.Printf("Invalid queue result: %v", result)
				continue
			}

			messageData := result[1]
			if err := c.processMessage(ctx, messageData); err != nil {
				log.Printf("Error processing message: %v", err)
				if retryErr := c.handleFailedMessage(ctx, queueName, messageData, err); retryErr != nil {
					log.Printf("Failed to handle failed message: %v", retryErr)
				}
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, messageData string) error {
	var message QueueMessage
	if err := json.Unmarshal([]byte(messageData), &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.Printf("Processing message: %+v", message)

	switch message.Type {
	case "user_created":
		return c.handleUserCreated(ctx, message)
	default:
		log.Printf("Unknown message type: %s", message.Type)
		return nil
	}
}

func (c *Consumer) handleUserCreated(ctx context.Context, message QueueMessage) error {
	// Parse the Kratos webhook payload
	var kratosPayload struct {
		UserId   string `json:"userId"`
		Username string `json:"username"`
	}

	if err := json.Unmarshal(message.Payload, &kratosPayload); err != nil {
		return fmt.Errorf("failed to parse Kratos payload: %w", err)
	}

	// Create user in tester database
	testerUserId := uuid.New()
	userCreation := models.UserCreation{
		Id:       testerUserId,
		KratosId: &kratosPayload.UserId,
		Username: kratosPayload.Username,
		Role:     "user", // Default role for new users
	}

	_, err := c.usersUC.CreateUser(ctx, &userCreation)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("Successfully created user with Kratos ID: %s", kratosPayload.UserId)
	return nil
}

func (c *Consumer) handleFailedMessage(ctx context.Context, queueName, messageData string, processingErr error) error {
	deadLetterQueue := queueName + ":dlq"

	type FailedMessage struct {
		OriginalMessage string `json:"original_message"`
		Error           string `json:"error"`
		FailedAt        string `json:"failed_at"`
	}

	failed := FailedMessage{
		OriginalMessage: messageData,
		Error:           processingErr.Error(),
		FailedAt:        time.Now().UTC().Format(time.RFC3339),
	}

	failedJSON, err := json.Marshal(failed)
	if err != nil {
		return fmt.Errorf("failed to marshal failed message: %w", err)
	}

	if err := c.redisClient.RPush(ctx, deadLetterQueue, failedJSON).Err(); err != nil {
		return fmt.Errorf("failed to push to dead letter queue: %w", err)
	}

	log.Printf("Message moved to dead letter queue: %s", deadLetterQueue)
	return nil
}
