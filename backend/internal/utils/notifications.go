package utils

import (
	"backend/config"
	"backend/internal/models"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// CreateNotification cria uma notificação para um usuário
func CreateNotification(userID uint, title, message, notifType string, metadata map[string]interface{}) error {
	// Criar a notificação no banco
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	notification := models.Notification{
		UserID:   userID,
		Title:    title,
		Message:  message,
		Type:     notifType,
		Read:     false,
		Metadata: string(metadataJSON),
	}

	// Salvar no banco
	if err := config.DB.Create(&notification).Error; err != nil {
		return err
	}

	// Enviar para o Django via webhook
	go ForwardNotificationToDjango(notification)

	return nil
}

// ForwardNotificationToDjango envia a notificação para o frontend via webhook
func ForwardNotificationToDjango(notification models.Notification) {
	// Configurar endpoint
	djangoWebhook := "http://localhost:8000/api/webhook/notifications/"
	if val, exists := os.LookupEnv("DJANGO_WEBHOOK_URL"); exists {
		djangoWebhook = val
	}

	// Montar payload
	payload := map[string]interface{}{
		"user_id":    notification.UserID,
		"title":      notification.Title,
		"message":    notification.Message,
		"type":       notification.Type,
		"read":       notification.Read,
		"created_at": notification.CreatedAt,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Erro ao serializar notificação: %v", err)
		return
	}

	// Enviar requisição
	resp, err := http.Post(
		djangoWebhook,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Printf("Erro ao enviar notificação para Django: %v", err)
		return
	}
	defer resp.Body.Close()

	// Log de resposta
	if resp.StatusCode != http.StatusOK {
		log.Printf("Webhook retorno não-OK: %d", resp.StatusCode)
	}
}
