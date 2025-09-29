package repository

import (
	"context"
	"database/sql"
	"sync"
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
	db   *sql.DB
	repo *PostgresRepository
}

func (suite *PostgresRepositoryTestSuite) SetupSuite() {
	connStr := "host=localhost port=5433 user=wallet_user password=wallet_password dbname=wallet_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		suite.T().Fatal(err)
	}

	suite.db = db
	suite.repo = NewPostgresRepository(db)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := db.PingContext(ctx); err != nil {
		suite.T().Fatalf("Failed to connect to database: %v", err)
	}
}

func (suite *PostgresRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *PostgresRepositoryTestSuite) SetupTest() {
	_, err := suite.db.Exec("DELETE FROM wallets")
	if err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *PostgresRepositoryTestSuite) TestGetBalance_Success() {
	walletID := uuid.New()
	_, err := suite.db.Exec("INSERT INTO wallets (id, balance) VALUES ($1, $2)", walletID, 1500)
	assert.NoError(suite.T(), err)

	balance, err := suite.repo.GetBalance(context.Background(), walletID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1500), balance)
}

func (suite *PostgresRepositoryTestSuite) TestGetBalance_WalletNotFound() {
	nonExistentWallet := uuid.New()
	balance, err := suite.repo.GetBalance(context.Background(), nonExistentWallet)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrWalletNotFound, err)
	assert.Equal(suite.T(), int64(0), balance)
}

func (suite *PostgresRepositoryTestSuite) TestUpdateBalance_Deposit() {
	walletID := uuid.New()
	_, err := suite.db.Exec("INSERT INTO wallets (id, balance) VALUES ($1, $2)", walletID, 1000)
	assert.NoError(suite.T(), err)

	err = suite.repo.UpdateBalance(context.Background(), walletID, 500)

	assert.NoError(suite.T(), err)

	var balance int64
	err = suite.db.QueryRow("SELECT balance FROM wallets WHERE id = $1", walletID).Scan(&balance)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1500), balance)
}

func (suite *PostgresRepositoryTestSuite) TestUpdateBalance_Withdraw() {
	walletID := uuid.New()
	_, err := suite.db.Exec("INSERT INTO wallets (id, balance) VALUES ($1, $2)", walletID, 1000)
	assert.NoError(suite.T(), err)

	err = suite.repo.UpdateBalance(context.Background(), walletID, -300)

	assert.NoError(suite.T(), err)

	var balance int64
	err = suite.db.QueryRow("SELECT balance FROM wallets WHERE id = $1", walletID).Scan(&balance)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(700), balance)
}

func (suite *PostgresRepositoryTestSuite) TestUpdateBalance_InsufficientFunds() {
	walletID := uuid.New()
	_, err := suite.db.Exec("INSERT INTO wallets (id, balance) VALUES ($1, $2)", walletID, 500)
	assert.NoError(suite.T(), err)

	err = suite.repo.UpdateBalance(context.Background(), walletID, -1000)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), models.ErrInsufficientFunds, err)

	var balance int64
	err = suite.db.QueryRow("SELECT balance FROM wallets WHERE id = $1", walletID).Scan(&balance)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(500), balance)
}

func (suite *PostgresRepositoryTestSuite) TestUpdateBalance_WalletNotFound() {
	nonExistentWallet := uuid.New()
	err := suite.repo.UpdateBalance(context.Background(), nonExistentWallet, 1000)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrWalletNotFound, err)
}

func (suite *PostgresRepositoryTestSuite) TestUpdateBalance_Concurrent() {
	walletID := uuid.New()
	_, err := suite.db.Exec("INSERT INTO wallets (id, balance) VALUES ($1, $2)", walletID, 1000)
	assert.NoError(suite.T(), err)

	iterations := 10
	var wg sync.WaitGroup
	errCh := make(chan error, iterations)

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			errCh <- suite.repo.UpdateBalance(ctx, walletID, 100)
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NoError(suite.T(), err)
	}

	var balance int64
	err = suite.db.QueryRow("SELECT balance FROM wallets WHERE id = $1", walletID).Scan(&balance)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2000), balance)
}

func (suite *PostgresRepositoryTestSuite) TestUpdateBalance_ConcurrentMixed() {
	walletID := uuid.New()
	_, err := suite.db.Exec("INSERT INTO wallets (id, balance) VALUES ($1, $2)", walletID, 1000)
	assert.NoError(suite.T(), err)

	var wg sync.WaitGroup
	errCh := make(chan error, 8)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- suite.repo.UpdateBalance(context.Background(), walletID, 200)
		}()
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errCh <- suite.repo.UpdateBalance(context.Background(), walletID, -100)
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NoError(suite.T(), err)
	}

	var balance int64
	err = suite.db.QueryRow("SELECT balance FROM wallets WHERE id = $1", walletID).Scan(&balance)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1700), balance)
}

func TestPostgresRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresRepositoryTestSuite))
}