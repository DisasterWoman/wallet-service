package load

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DisasterWoman/wallet-service/internal/models"
	"github.com/DisasterWoman/wallet-service/internal/repository"
	"github.com/DisasterWoman/wallet-service/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
)

func TestWalletService_Load_1000RPS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	connStr := "host=localhost port=5433 user=wallet_user password=wallet_password dbname=wallet_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewPostgresRepository(db)
	walletService := service.NewWalletService(repo)

	walletID := uuid.New()
	_, err = db.Exec("INSERT INTO wallets (id, balance) VALUES ($1, $2)", walletID, 0)
	assert.NoError(t, err)
	defer db.Exec("DELETE FROM wallets WHERE id = $1", walletID)

	totalRequests := 1000
	concurrentWorkers := 100
	requestsPerWorker := totalRequests / concurrentWorkers

	var wg sync.WaitGroup
	errorCh := make(chan error, totalRequests)
	successCh := make(chan bool, totalRequests)

	startTime := time.Now()

	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < requestsPerWorker; j++ {
				req := &models.OperationRequest{
					WalletID:      walletID,
					OperationType: models.Deposit,
					Amount:        1,
				}
				
				err := walletService.UpdateBalance(context.Background(), req)
				if err != nil {
					errorCh <- fmt.Errorf("worker %d, request %d: %w", workerID, j, err)
				} else {
					successCh <- true
				}
				
				time.Sleep(time.Microsecond * 100)
			}
		}(i)
	}

	wg.Wait()
	close(errorCh)
	close(successCh)

	duration := time.Since(startTime)
	actualRPS := float64(totalRequests) / duration.Seconds()

	successCount := len(successCh)
	errorCount := 0
	for range errorCh {
		errorCount++
	}

	finalBalance, err := walletService.GetBalance(context.Background(), walletID)
	assert.NoError(t, err)

	t.Logf("Load Test Results:")
	t.Logf("Duration: %v", duration)
	t.Logf("Total Requests: %d", totalRequests)
	t.Logf("Successful: %d", successCount)
	t.Logf("Errors: %d", errorCount)
	t.Logf("Actual RPS: %.2f", actualRPS)
	t.Logf("Final Balance: %d", finalBalance)

	// Проверяем что все запросы обработаны
	assert.Equal(t, totalRequests, successCount, "All requests should be processed successfully")
	assert.Equal(t, 0, errorCount, "No requests should fail")
	assert.Equal(t, int64(totalRequests), finalBalance, "Final balance should equal total deposits")
	assert.True(t, actualRPS >= 100, "Should handle at least 100 RPS, got %.2f", actualRPS)
}


func TestWalletService_Load_StableConcurrentDeposits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	connStr := "host=localhost port=5433 user=wallet_user password=wallet_password dbname=wallet_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	assert.NoError(t, err)
	defer db.Close()

	repo := repository.NewPostgresRepository(db)
	walletService := service.NewWalletService(repo)

	walletID := uuid.New()
	_, err = db.Exec("INSERT INTO wallets (id, balance) VALUES ($1, $2)", walletID, 0)
	assert.NoError(t, err)
	defer db.Exec("DELETE FROM wallets WHERE id = $1", walletID)

	totalRequests := 800
	concurrentWorkers := 80

	var wg sync.WaitGroup
	errorCh := make(chan error, totalRequests)
	successCount := int64(0)

	startTime := time.Now()

	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			requestsPerWorker := totalRequests / concurrentWorkers
			
			for j := 0; j < requestsPerWorker; j++ {
				req := &models.OperationRequest{
					WalletID:      walletID,
					OperationType: models.Deposit,
					Amount:        1,
				}
				
				err := walletService.UpdateBalance(context.Background(), req)
				if err != nil {
					errorCh <- fmt.Errorf("worker %d: %w", workerID, err)
				} else {
					successCount++
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorCh)

	duration := time.Since(startTime)
	actualRPS := float64(totalRequests) / duration.Seconds()

	errorCount := 0
	for range errorCh {
		errorCount++
	}

	finalBalance, err := walletService.GetBalance(context.Background(), walletID)
	assert.NoError(t, err)

	t.Logf("Stable Concurrent Deposits Test:")
	t.Logf("Duration: %v", duration)
	t.Logf("Total Requests: %d", totalRequests)
	t.Logf("Successful: %d", successCount)
	t.Logf("Errors: %d", errorCount)
	t.Logf("Actual RPS: %.2f", actualRPS)
	t.Logf("Final Balance: %d", finalBalance)

	assert.Equal(t, 0, errorCount, "Pure deposits should have no errors")
	assert.Equal(t, int64(totalRequests), finalBalance, "Final balance should match total deposits")
	assert.True(t, actualRPS >= 200, "Should handle at least 200 RPS for deposits, got %.2f", actualRPS)
}