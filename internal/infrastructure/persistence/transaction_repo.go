package persistence

import (
	"go-finance-advisor/internal/domain"

	"gorm.io/gorm"
)

type UserRepository struct{ DB *gorm.DB }

type TransactionRepository struct{ DB *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{DB: db}
}

func (r *UserRepository) Create(u *domain.User) error { return r.DB.Create(u).Error }
func (r *UserRepository) GetByID(id uint) (domain.User, error) {
	var u domain.User
	err := r.DB.First(&u, id).Error
	return u, err
}
func (r *UserRepository) Update(u *domain.User) error { return r.DB.Save(u).Error }

func (r *TransactionRepository) Create(tx *domain.Transaction) error { return r.DB.Create(tx).Error }
func (r *TransactionRepository) ListByUser(userID uint) ([]domain.Transaction, error) {
	var txs []domain.Transaction
	err := r.DB.Where("user_id = ?", userID).Order("date desc").Find(&txs).Error
	return txs, err
}
