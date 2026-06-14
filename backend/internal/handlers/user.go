package handlers

import (
	"backend/config"
	"backend/internal/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Default role — Remover possibilidade de criar admin via API
	// Somente o primeiro usuário criado será admin
	var role string
	var firstUser models.User
	if config.DB.First(&firstUser).Error != nil {
				// Nenhum usuário existe, criar como admin
		role = "admin"
	} else {
				// Usuário existe, novo usuário será comum
		role = "comum"
	}
	// Ignorar role enviado pelo cliente (prevenir privilege escalation)

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	user.Password = "" // clear before sending
	c.JSON(http.StatusCreated, user)
}

func GetUsers(c *gin.Context) {
	params := ParsePaginateParams(c)
	resp := Paginate(config.DB, params, &[]models.User{})

	// Filter passwords from paginated result
	if data, ok := resp.Data.(*[]models.User); ok {
		for i := range *data {
			(*data)[i].Password = ""
		}
	}

	c.JSON(http.StatusOK, resp)
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateUser: bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Name = req.Name
	user.Email = req.Email

	if err := config.DB.Save(&user).Error; err != nil {
		log.Printf("UpdateUser: save error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	user.Password = ""
	log.Printf("User updated: %s (id=%d)", user.Name, user.ID)
	c.JSON(http.StatusOK, user)
}

type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

func ResetPassword(c *gin.Context) {
	id := c.Param("id")
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ResetPassword: bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha deve ter pelo menos 8 caracteres"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ResetPassword: hash error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Password = string(hashedPassword)
	if err := config.DB.Save(&user).Error; err != nil {
		log.Printf("ResetPassword: save error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	user.Password = ""
	log.Printf("Password reset for user: %s (id=%d)", user.Name, user.ID)
	c.JSON(http.StatusOK, user)
}

type DisableUserRequest struct {
	Disabled *bool `json:"disabled"`
}

func DisableUser(c *gin.Context) {
	id := c.Param("id")
	var req DisableUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("DisableUser: bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Toggle or set disabled
	newState := *req.Disabled
	user.Disabled = &newState

	if err := config.DB.Save(&user).Error; err != nil {
		log.Printf("DisableUser: save error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	user.Password = ""
	if newState {
		log.Printf("User DISABLED: %s (id=%d)", user.Name, user.ID)
	} else {
		log.Printf("User RE-ENABLED: %s (id=%d)", user.Name, user.ID)
	}
	c.JSON(http.StatusOK, user)
}
