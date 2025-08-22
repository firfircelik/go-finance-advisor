package persistence

import (
	"testing"
	"time"

	"go-finance-advisor/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	return gormDB, mock
}

func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		user    *domain.User
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful user creation",
			user: &domain.User{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Password:  "hashedpassword",
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").
					WithArgs("john@example.com", "hashedpassword", "John", "Doe", 30, "moderate", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "database error during creation",
			user: &domain.User{
				FirstName: "Jane",
				LastName:  "Doe",
				Email:     "jane@example.com",
				Password:  "hashedpassword",
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").
					WillReturnError(gorm.ErrInvalidDB)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer func() {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			}()

			repo := NewUserRepository(db)
			tt.setup(mock)

			err := repo.Create(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	tests := []struct {
		name     string
		userID   uint
		setup    func(mock sqlmock.Sqlmock)
		expected domain.User
		wantErr  bool
	}{
		{
			name:   "successful user retrieval",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "email", "password"}).
					AddRow(1, time.Now(), time.Now(), nil, "John", "Doe", "john@example.com", "hashedpassword")
				mock.ExpectQuery("SELECT \\* FROM `users`").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			expected: domain.User{
				ID:        1,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Password:  "hashedpassword",
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 999,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `users`").
					WithArgs(999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expected: domain.User{},
			wantErr:  true,
		},
		{
			name:   "database error during retrieval",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `users`").
					WithArgs(1, 1).
					WillReturnError(gorm.ErrInvalidDB)
			},
			expected: domain.User{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer func() {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			}()

			repo := NewUserRepository(db)
			tt.setup(mock)

			user, err := repo.GetByID(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, user.ID)
				assert.Equal(t, tt.expected.FirstName, user.FirstName)
				assert.Equal(t, tt.expected.LastName, user.LastName)
				assert.Equal(t, tt.expected.Email, user.Email)
				assert.Equal(t, tt.expected.Password, user.Password)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	tests := []struct {
		name    string
		user    *domain.User
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "successful user update",
			user: &domain.User{
				ID:            1,
				FirstName:     "John",
				LastName:      "Updated",
				Age:           0,
				Email:         "john.updated@example.com",
				Password:      "newhashedpassword",
				RiskTolerance: "moderate",
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users`").
					WithArgs("john.updated@example.com", "newhashedpassword", "John", "Updated", 0, "moderate", sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "database error during update",
			user: &domain.User{
				ID:            1,
				FirstName:     "John",
				LastName:      "Updated",
				Age:           0,
				Email:         "john.updated@example.com",
				Password:      "newhashedpassword",
				RiskTolerance: "moderate",
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `users`").
					WillReturnError(gorm.ErrInvalidDB)
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer func() {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			}()

			repo := NewUserRepository(db)
			tt.setup(mock)

			err := repo.Update(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNewUserRepository(t *testing.T) {
	db, mock := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewUserRepository(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.DB)
	assert.NoError(t, mock.ExpectationsWereMet())
}
