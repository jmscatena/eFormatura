package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `json:"name" binding:"required"`
	Email      string    `gorm:"unique" json:"email" binding:"required,email"`
	Password   string    `json:"password,omitempty" binding:"required"` // only for creation
	Role       string    `json:"role"`                                   // "admin" or "comum"
	Disabled   *bool     `json:"disabled"`                               // soft delete flag
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Income struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Title    string    `json:"title" binding:"required"`
	Amount   float64   `json:"amount" binding:"required"`
	Date     time.Time `json:"date"`
	Category string    `json:"category"` // "Rifa", "Evento", "Mensalidade"
	UserID   *uint     `json:"user_id"`  // Who paid or sold
	User     *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type Expense struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	ContractName    string        `json:"contract_name" binding:"required"`
	TotalAmount     float64       `json:"total_amount" binding:"required"`
	Description     string        `json:"description"`
	Category        string        `json:"category"` // "Contrato" or "Custo"
	Start_Date      time.Time     `json:"start_date"`
	First_Payment_Date time.Time `json:"first_payment_date"`
	Installments    []Installment `json:"installments,omitempty"`
}

type Installment struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ExpenseID   uint       `json:"expense_id"`
	Amount      float64    `json:"amount" binding:"required"`
	DueDate     time.Time  `json:"due_date" binding:"required"`
	Status      string     `json:"status"` // "Pendente", "Pago"
	PaymentDate *time.Time `json:"payment_date"`
}

// Notification model for user alerts
type Notification struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uint       `json:"user_id" binding:"required"`
	Title     string    `json:"title" binding:"required"`
	Message   string    `json:"message"`
	Type      string    `json:"type" gorm:"default:'info'"` // info, success, warning, error
	Read      bool      `json:"read" gorm:"default:false"`
	Metadata  string    `json:"metadata,omitempty"` // JSON extra fields
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Income{}, &Expense{}, &Installment{}, &Notification{})
}
