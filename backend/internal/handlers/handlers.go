package handlers

import (
	"backend/config"
	"backend/internal/models"
	"backend/internal/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Incomes
func GetIncomes(c *gin.Context) {
	var incomes []models.Income
	if err := config.DB.Preload("User").Find(&incomes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, incomes)
}

func CreateIncome(c *gin.Context) {
	var income models.Income
	if err := c.ShouldBindJSON(&income); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set user_id from authenticated user (admin context)
	userID, _ := c.Get("userID")
	userIDUint, ok := userID.(uint)
	if ok {
		income.UserID = &userIDUint
	}

	if err := config.DB.Create(&income).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, income)
}

// Expenses
func GetExpenses(c *gin.Context) {
	var expenses []models.Expense
	if err := config.DB.Preload("Installments").Find(&expenses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, expenses)
}

func CreateExpense(c *gin.Context) {
	var req struct {
		ContractName     string    `json:"contract_name" binding:"required,max=70"`
		TotalAmount      float64   `json:"total_amount" binding:"required,gt=0"`
		Description      string    `json:"description"`
		Category         string    `json:"category"`
		InstallmentCount int       `json:"installment_count" binding:"gte=1,lte=120"`
		StartDate        string    `json:"start_date"`
		FirstPaymentDate string    `json:"first_payment_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cat := "Contrato" // default
	if req.Category != "" {
		if req.Category != "Contrato" && req.Category != "Custo" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Categoria inválida. Deve ser 'Contrato' ou 'Custo'"})
			return
		}
		cat = req.Category
	}

	var startDate time.Time
	var firstPaymentDate time.Time

	if req.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = parsed
		}
	} else {
		startDate = time.Now()
	}

	if req.FirstPaymentDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.FirstPaymentDate); err == nil {
			firstPaymentDate = parsed
		}
	}

	if !startDate.IsZero() && firstPaymentDate.IsZero() {
		firstPaymentDate = startDate
	}

	if !firstPaymentDate.IsZero() && !startDate.IsZero() && firstPaymentDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A data do primeiro pagamento deve ser igual ou posterior à data de início do contrato"})
		return
	}

	expense := models.Expense{
		ContractName:     req.ContractName,
		TotalAmount:      req.TotalAmount,
		Description:      req.Description,
		Category:         cat,
		Start_Date:       startDate,
		First_Payment_Date: firstPaymentDate,
	}

	if err := config.DB.Create(&expense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.InstallmentCount > 0 {
		amount := req.TotalAmount / float64(req.InstallmentCount)
		monthsInterval := 1
		if !firstPaymentDate.IsZero() && !startDate.IsZero() {
			daysDiff := int(firstPaymentDate.Sub(startDate).Hours() / 24)
			monthsDiff := daysDiff / 30
			if monthsDiff > 0 {
				monthsInterval = monthsDiff
			}
		}
		for i := 0; i < req.InstallmentCount; i++ {
			inst := models.Installment{
				ExpenseID: expense.ID,
				Amount:    amount,
				DueDate:   firstPaymentDate.AddDate(0, i*monthsInterval, 0),
				Status:    "Pendente",
			}
			config.DB.Create(&inst)
		}
	}
	c.JSON(http.StatusCreated, expense)
}

// Installment Payment — IDOR fix: verificar se user tem permissão
func PayInstallment(c *gin.Context) {
	id := c.Param("id")
	var installment models.Installment

	if err := config.DB.First(&installment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Installment not found"})
		return
	}

	// Admin podem pagar qualquer parcela
	userRole, exists := c.Get("role")
	if !exists || userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	installment.Status = "Pago"
	// time.Now() could be set to payment_date, but to keep simple we just update status
	if err := config.DB.Save(&installment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update installment"})
		return
	}

	// Carregar expense para obter user_id
	var expense models.Expense
	if err := config.DB.First(&expense, installment.ExpenseID).Error; err == nil {
		// Notificar usuário (quem pagou)
		userID := c.GetUint("userID")
		utils.CreateNotification(
			userID,
			"Parcela paga",
			fmt.Sprintf("Parcela de R$%.2f do contrato '%s' foi paga", installment.Amount, expense.ContractName),
			"success",
			map[string]interface{}{"expense_id": expense.ID, "installment_id": installment.ID},
		)
	}

	c.JSON(http.StatusOK, installment)
}

// Income (Update & Delete) — IDOR fix: verificar ownership
func UpdateIncome(c *gin.Context) {
	id := c.Param("id")
	var income models.Income

	if err := config.DB.First(&income, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Income not found"})
		return
	}

	// IDOR Fix: Verificar se usuário é dono da income ou admin
	userID, _ := c.Get("userID")
	userIDUint, ok := userID.(uint)
	if ok && income.UserID != nil && *income.UserID != userIDUint {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	// Admin pode atualizar qualquer income
	userRole, roleExists := c.Get("role")
	if !roleExists || userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	if err := c.ShouldBindJSON(&income); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Save(&income).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, income)
}

func DeleteIncome(c *gin.Context) {
	id := c.Param("id")

	// IDOR Fix: Verificar ownership antes de deletar
	var income models.Income
	if err := config.DB.First(&income, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Income not found"})
		return
	}

	// Admin podem deletar qualquer income
	userRole, roleExists := c.Get("role")
	if !roleExists || userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	if income.UserID != nil {
		userID, _ := c.Get("userID")
		userIDUint, ok := userID.(uint)
		if ok && *income.UserID != userIDUint {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
	}

	if err := config.DB.Delete(&income).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Income deleted"})
}

// Expense (Update & Delete) — IDOR fix
func UpdateExpense(c *gin.Context) {
	id := c.Param("id")
	var expense models.Expense

	if err := config.DB.First(&expense, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
		return
	}

	// Admin podem atualizar qualquer expense
	userRole, roleExists := c.Get("role")
	if !roleExists || userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Save(&expense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expense)
}

func DeleteExpense(c *gin.Context) {
	id := c.Param("id")

	var expense models.Expense
	if err := config.DB.First(&expense, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expense not found"})
		return
	}

	userRole, roleExists := c.Get("role")
	if !roleExists || userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	if err := config.DB.Delete(&expense).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Expense deleted"})
}

// Installment (Create, Update, Delete) — IDOR fix
func CreateInstallment(c *gin.Context) {
	var installment models.Installment
	if err := c.ShouldBindJSON(&installment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Create(&installment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, installment)
}

func UpdateInstallment(c *gin.Context) {
	id := c.Param("id")
	var installment models.Installment

	if err := config.DB.First(&installment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Installment not found"})
		return
	}

	// Admin podem atualizar qualquer installment
	userRole, roleExists := c.Get("role")
	if !roleExists || userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	if err := c.ShouldBindJSON(&installment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.DB.Save(&installment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, installment)
}

func DeleteInstallment(c *gin.Context) {
	id := c.Param("id")

	var installment models.Installment
	if err := config.DB.First(&installment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Installment not found"})
		return
	}

	userRole, roleExists := c.Get("role")
	if !roleExists || userRole.(string) != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	if err := config.DB.Delete(&installment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Installment deleted"})
}
