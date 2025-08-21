package persistence

import (
	"testing"
	"time"

	"go-finance-advisor/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestTransactionRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		transaction *domain.Transaction
		setup       func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name: "successful transaction creation",
			transaction: &domain.Transaction{
				UserID:      1,
				CategoryID:  1,
				Amount:      100.50,
				Description: "Test transaction",
				Type:        "expense",
				Date:        time.Now(),
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `transactions`").
					WithArgs(1, 1, "expense", "Test transaction", 100.50, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "database error during creation",
			transaction: &domain.Transaction{
				UserID:      1,
				CategoryID:  2,
				Amount:      50.25,
				Description: "Failed transaction",
				Type:        "income",
				Date:        time.Now(),
			},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `transactions`").
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

			repo := NewTransactionRepository(db)
			tt.setup(mock)

			err := repo.Create(tt.transaction)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTransactionRepository_ListByUser(t *testing.T) {
	testDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		userID   uint
		setup    func(mock sqlmock.Sqlmock)
		expected []domain.Transaction
		wantErr  bool
	}{
		{
			name:   "successful transaction list retrieval",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "user_id", "category_id", "amount", "description", "type", "date"}).
				AddRow(1, testDate, testDate, nil, 1, 1, 100.50, "Grocery shopping", "expense", testDate).
				AddRow(2, testDate, testDate, nil, 1, 2, 2500.00, "Monthly salary", "income", testDate)
				mock.ExpectQuery("SELECT \\* FROM `transactions`").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expected: []domain.Transaction{
				{
					ID:          1,
					UserID:      1,
					CategoryID:  1,
					Amount:      100.50,
					Description: "Grocery shopping",
					Type:        "expense",
					Date:        testDate,
				},
				{
					ID:          2,
					UserID:      1,
					CategoryID:  2,
					Amount:      2500.00,
					Description: "Monthly salary",
					Type:        "income",
					Date:        testDate,
				},
			},
			wantErr: false,
		},
		{
			name:   "empty transaction list",
			userID: 2,
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "user_id", "category_id", "amount", "description", "type", "date"})
				mock.ExpectQuery("SELECT \\* FROM `transactions`").
					WithArgs(2).
					WillReturnRows(rows)
			},
			expected: []domain.Transaction{},
			wantErr:  false,
		},
		{
			name:   "database error during retrieval",
			userID: 1,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `transactions`").
					WithArgs(1).
					WillReturnError(gorm.ErrInvalidDB)
			},
			expected: nil,
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

			repo := NewTransactionRepository(db)
			tt.setup(mock)

			transactions, err := repo.ListByUser(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, transactions)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(transactions))
				for i, expected := range tt.expected {
					if i < len(transactions) {
						assert.Equal(t, expected.ID, transactions[i].ID)
						assert.Equal(t, expected.UserID, transactions[i].UserID)
						assert.Equal(t, expected.Amount, transactions[i].Amount)
						assert.Equal(t, expected.Description, transactions[i].Description)
						assert.Equal(t, expected.Type, transactions[i].Type)
						assert.Equal(t, expected.CategoryID, transactions[i].CategoryID)
						assert.Equal(t, expected.Date.Unix(), transactions[i].Date.Unix())
					}
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNewTransactionRepository(t *testing.T) {
	db, mock := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	repo := NewTransactionRepository(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.DB)
	assert.NoError(t, mock.ExpectationsWereMet())
}