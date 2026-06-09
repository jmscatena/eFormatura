package handlers

import (
	"backend/config"
	"backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListNotifications returns user's notifications
func ListNotifications(c *gin.Context) {
	userID := c.GetUint("userID")
	var notifications []models.Notification

	// Buscar do banco
	if err := config.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// MarkNotificationAsRead marks a single notification as read
func MarkNotificationAsRead(c *gin.Context) {
	userID := c.GetUint("userID")
	notificationID := c.Param("id")

	var notification models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	notification.Read = true
	if err := config.DB.Save(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// MarkAllNotificationsAsRead marks all user's notifications as read
func MarkAllNotificationsAsRead(c *gin.Context) {
	userID := c.GetUint("userID")

	if err := config.DB.Model(&models.Notification{}).Where("user_id = ?", userID).Update("read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
