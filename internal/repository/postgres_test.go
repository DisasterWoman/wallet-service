package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DisasterWoman/wallet-service/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	_ "github.com/lib/pq"
)

type PostgresRepositoryTestSuite struct {
	suite.Suite
	db         *sql.DB
	repo       *PostgresRepository
	testWallet uuid.UUID
}

func (suite *PostgresRepositoryTestSuite) SetupSuite() {
	connStr := "host=localhost port=5433 user=wallet_user password=wallet_password dbname=wallet_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.db = db
	suite.repo = NewPostgresRepository(db)
	suite.testWallet = uuid.New()

	_, err = suite.db.Exec(
		"INSERT INTO wallets (id, balance) VALUES ($1, $2)",
		suite.testWallet, 1000,
	)
	if err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *PostgresRepositoryTestSuite) TearDownSuite() {
	suite.db.Exec("DELETE FROM wallets WHERE id = $1", suite.testWallet)
	suite.db.Close()
}

func (suite *PostgresRepositoryTestSuite) TestGetBalance() {
	balance, err := suite.repo.GetBalance(context.Background(), suite.testWallet)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1000), balance)
}

func (suite *PostgresRepositoryTestSuite) TestGetBalance_WalletNotFound() {
	nonExistentWallet := uuid.New()
	balance, err := suite.repo.GetBalance(context.Background(), nonExistentWallet)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrWalletNotFound, err)
	assert.Equal(suite.T(), int64(0), balance)
}

func (suite *PostgresRepositoryTestSuite) TestUpdateBalance_Deposit() {
	err := suite.repo.UpdateBalance(context.Background(), suite.testWallet, 500)

	assert.NoError(suite.T(), err)

	balance, _ := suite.repo.GetBalance(context.Background(), suite.testWallet)
	assert.Equal(suite.T(), int64(1500), balance)
}

func (suite *PostgresRepositoryTestSuite) TestUpdateBalance_Withdraw() {
	err := suite.repo.UpdateBalance(context.Background(), suite.testWallet, -300)

	assert.NoError(suite.T(), err)

	balance, _ := suite.repo.GetBalance(context.Background(), suite.testWallet)
	assert.Equal(suite.T(), int64(700), balance)
}

func (suite *PostgresRepositoryTestSuite) TestUpdateBalance_InsufficientFunds() {
	err := suite.repo.UpdateBalance(context.Background(), suite.testWallet, -2000)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrInsufficientFunds, err)

	balance, _ := suite.repo.GetBalance(context.Background(), suite.testWallet)
	assert.Equal(suite.T(), int64(1000), balance)
}

func (suite *PostgresRepositoryTestSuite) TestUpdateBalance_Concurrent() {
	iterations := 10
	errCh := make(chan error, iterations)

	for i := 0; i < iterations; i++ {
		go func(amount int64) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			errCh <- suite.repo.UpdateBalance(ctx, suite.testWallet, amount)
		}(100) // Все пополнения на 100
	}

	for i := 0; i < iterations; i++ {
		err := <-errCh
		assert.NoError(suite.T(), err)
	}

	balance, err := suite.repo.GetBalance(context.Background(), suite.testWallet)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2000), balance)
}

func TestPostgresRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresRepositoryTestSuite))
}