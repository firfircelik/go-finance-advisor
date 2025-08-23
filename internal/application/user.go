// Package application contains the business logic and use cases
// for the Go Finance Advisor application.
package application

import (
	"errors"

	"go-finance-advisor/internal/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

// Register creates a new user with email and password
func (s *UserService) Register(email, password, firstName, lastName string) (*domain.User, error) {
	// Check if user already exists
	var existingUser domain.User
	if err := s.DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Email:         email,
		Password:      string(hashedPassword),
		FirstName:     firstName,
		LastName:      lastName,
		RiskTolerance: "moderate", // default
	}

	if err := s.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates user with email and password
func (s *UserService) Login(email, password string) (*domain.User, error) {
	var user domain.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

// GetByEmail finds user by email
func (s *UserService) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) Create(u *domain.User) error {
	return s.DB.Create(u).Error
}

func (s *UserService) GetByID(id uint) (domain.User, error) {
	var u domain.User
	err := s.DB.First(&u, id).Error
	return u, err
}

func (s *UserService) Update(u *domain.User) error {
	return s.DB.Save(u).Error
}
