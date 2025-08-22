package api

import (
	"net/http"
	"strconv"

	"go-finance-advisor/internal/domain"
	"go-finance-advisor/internal/infrastructure/middleware"

	"github.com/gin-gonic/gin"
)

// Risk tolerance constants
const (
	riskToleranceConservative = "conservative"
	riskToleranceModerate     = "moderate"
	riskToleranceAggressive   = "aggressive"
)

// UserServiceInterface defines the contract for user service operations
type UserServiceInterface interface {
	Create(user *domain.User) error
	GetByID(id uint) (domain.User, error)
	Update(user *domain.User) error
	GetByEmail(email string) (*domain.User, error)
	Register(email, password, firstName, lastName string) (*domain.User, error)
	Login(email, password string) (*domain.User, error)
}

type UserHandler struct {
	Service UserServiceInterface
}

func NewUserHandler(service UserServiceInterface) *UserHandler {
	return &UserHandler{Service: service}
}

func (h *UserHandler) Create(c *gin.Context) {
	var u domain.User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if u.RiskTolerance == "" {
		u.RiskTolerance = riskToleranceModerate
	}
	if err := h.Service.Create(&u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (h *UserHandler) Get(c *gin.Context) {
	idStr := c.Param("userId")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	u, err := h.Service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register creates a new user account
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.Service.Register(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if err.Error() == "user already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	// Generate JWT token for immediate login
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
		},
		"token": token,
	})
}

// Login authenticates user with email and password
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.Service.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
		},
		"token": token,
	})
}

type RiskUpdateRequest struct {
	RiskTolerance string `json:"risk_tolerance"`
}

func (h *UserHandler) UpdateRisk(c *gin.Context) {
	idStr := c.Param("userId")
	id, _ := strconv.Atoi(idStr)

	var req RiskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate risk tolerance
	if req.RiskTolerance != riskToleranceConservative &&
		req.RiskTolerance != riskToleranceModerate &&
		req.RiskTolerance != riskToleranceAggressive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid risk tolerance"})
		return
	}

	if id < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.Service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	user.RiskTolerance = req.RiskTolerance
	if err := h.Service.Update(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, user)
}
