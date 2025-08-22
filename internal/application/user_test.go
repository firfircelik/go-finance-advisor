package application

import (
	"testing"

	"go-finance-advisor/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupUserTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	return db
}

func TestUserService_Register(t *testing.T) {
	db := setupUserTestDB(t)
	userService := &UserService{DB: db}

	t.Run("successful registration", func(t *testing.T) {
		user, err := userService.Register("test@example.com", "password123", "John", "Doe")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, "moderate", user.RiskTolerance)
		assert.NotEmpty(t, user.Password)

		// Verify password is hashed
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
		assert.NoError(t, err)
	})

	t.Run("duplicate email registration", func(t *testing.T) {
		// First registration
		_, err := userService.Register("duplicate@example.com", "password123", "Jane", "Doe")
		assert.NoError(t, err)

		// Second registration with same email
		user, err := userService.Register("duplicate@example.com", "password456", "John", "Smith")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user already exists")
	})

	t.Run("empty fields", func(t *testing.T) {
		user, err := userService.Register("", "", "", "")

		// Should still create user but with empty fields
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "moderate", user.RiskTolerance)
	})
}

func TestUserService_Login(t *testing.T) {
	db := setupUserTestDB(t)
	userService := &UserService{DB: db}

	// Create a test user
	testEmail := "login@example.com"
	testPassword := "testpassword"
	_, err := userService.Register(testEmail, testPassword, "Test", "User")
	require.NoError(t, err)

	t.Run("successful login", func(t *testing.T) {
		user, err := userService.Login(testEmail, testPassword)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testEmail, user.Email)
		assert.Equal(t, "Test", user.FirstName)
		assert.Equal(t, "User", user.LastName)
	})

	t.Run("invalid email", func(t *testing.T) {
		user, err := userService.Login("nonexistent@example.com", testPassword)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("invalid password", func(t *testing.T) {
		user, err := userService.Login(testEmail, "wrongpassword")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("empty credentials", func(t *testing.T) {
		user, err := userService.Login("", "")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "invalid credentials")
	})
}

func TestUserService_GetByEmail(t *testing.T) {
	db := setupUserTestDB(t)
	userService := &UserService{DB: db}

	// Create a test user
	testEmail := "getbyemail@example.com"
	createdUser, err := userService.Register(testEmail, "password", "Get", "ByEmail")
	require.NoError(t, err)

	t.Run("existing user", func(t *testing.T) {
		user, err := userService.GetByEmail(testEmail)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, testEmail, user.Email)
		assert.Equal(t, "Get", user.FirstName)
		assert.Equal(t, "ByEmail", user.LastName)
	})

	t.Run("non-existing user", func(t *testing.T) {
		user, err := userService.GetByEmail("nonexistent@example.com")

		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("empty email", func(t *testing.T) {
		user, err := userService.GetByEmail("")

		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserService_GetByID(t *testing.T) {
	db := setupUserTestDB(t)
	userService := &UserService{DB: db}

	// Create a test user
	createdUser, err := userService.Register("getbyid@example.com", "password", "Get", "ByID")
	require.NoError(t, err)

	t.Run("existing user", func(t *testing.T) {
		user, err := userService.GetByID(createdUser.ID)

		assert.NoError(t, err)
		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, "getbyid@example.com", user.Email)
		assert.Equal(t, "Get", user.FirstName)
		assert.Equal(t, "ByID", user.LastName)
	})

	t.Run("non-existing user", func(t *testing.T) {
		user, err := userService.GetByID(99999)

		assert.Error(t, err)
		assert.Equal(t, uint(0), user.ID)
	})
}

func TestUserService_Create(t *testing.T) {
	db := setupUserTestDB(t)
	userService := &UserService{DB: db}

	t.Run("create user", func(t *testing.T) {
		user := &domain.User{
			Email:         "create@example.com",
			Password:      "hashedpassword",
			FirstName:     "Create",
			LastName:      "Test",
			RiskTolerance: "aggressive",
		}

		err := userService.Create(user)

		assert.NoError(t, err)
		assert.NotZero(t, user.ID)

		// Verify user was created
		retrievedUser, err := userService.GetByEmail("create@example.com")
		assert.NoError(t, err)
		assert.Equal(t, user.ID, retrievedUser.ID)
	})

	t.Run("create user with duplicate email", func(t *testing.T) {
		user1 := &domain.User{
			Email:     "duplicate@example.com",
			Password:  "password1",
			FirstName: "User",
			LastName:  "One",
		}

		err := userService.Create(user1)
		assert.NoError(t, err)

		user2 := &domain.User{
			Email:     "duplicate@example.com",
			Password:  "password2",
			FirstName: "User",
			LastName:  "Two",
		}

		err = userService.Create(user2)
		assert.Error(t, err) // Should fail due to unique constraint
	})
}

func TestUserService_Update(t *testing.T) {
	db := setupUserTestDB(t)
	userService := &UserService{DB: db}

	// Create a test user
	user, err := userService.Register("update@example.com", "password", "Update", "Test")
	require.NoError(t, err)

	t.Run("update user fields", func(t *testing.T) {
		user.FirstName = "Updated"
		user.LastName = "Name"
		user.RiskTolerance = "conservative"

		err := userService.Update(user)
		assert.NoError(t, err)

		// Verify update
		updatedUser, err := userService.GetByID(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated", updatedUser.FirstName)
		assert.Equal(t, "Name", updatedUser.LastName)
		assert.Equal(t, "conservative", updatedUser.RiskTolerance)
	})

	t.Run("update non-existing user", func(t *testing.T) {
		nonExistentUser := &domain.User{
			ID:        99999,
			Email:     "nonexistent@example.com",
			FirstName: "Non",
			LastName:  "Existent",
		}

		err := userService.Update(nonExistentUser)
		// GORM will create a new record if ID doesn't exist
		assert.NoError(t, err)
	})
}
