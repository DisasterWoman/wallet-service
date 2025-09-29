package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/DisasterWoman/wallet-service/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WalletAPITestSuite struct {
	suite.Suite
	baseURL string
	client  *http.Client
}

func (suite *WalletAPITestSuite) SetupSuite() {
	suite.baseURL = "http://localhost:8080"
	suite.client = &http.Client{
		Timeout: 10 * time.Second,
	}
}

func (suite *WalletAPITestSuite) TestWalletAPI_EndToEnd() {
	walletID := uuid.New()

	// 1. Попробуем получить баланс несуществующего кошелька
	balance, err := suite.getBalance(walletID)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "wallet not found")

	// 2. Создаем кошелек через депозит
	err = suite.updateBalance(walletID, models.Deposit, 1000)
	assert.NoError(suite.T(), err)

	// 3. Проверяем баланс
	balance, err = suite.getBalance(walletID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1000), balance)

	// 4. Снимаем средства
	err = suite.updateBalance(walletID, models.Withdraw, 300)
	assert.NoError(suite.T(), err)

	// 5. Проверяем обновленный баланс
	balance, err = suite.getBalance(walletID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(700), balance)

	// 6. Пробуем снять больше чем есть
	err = suite.updateBalance(walletID, models.Withdraw, 1000)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "insufficient funds")

	// 7. Проверяем что баланс не изменился
	balance, err = suite.getBalance(walletID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(700), balance)

	// 8. Еще один депозит
	err = suite.updateBalance(walletID, models.Deposit, 500)
	assert.NoError(suite.T(), err)

	// 9. Финальная проверка баланса
	balance, err = suite.getBalance(walletID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1200), balance)
}

func (suite *WalletAPITestSuite) TestWalletAPI_ConcurrentEndToEnd() {
	walletID := uuid.New()

	// Начальный депозит
	err := suite.updateBalance(walletID, models.Deposit, 10000)
	assert.NoError(suite.T(), err)

	// 10 concurrent операций
	var successCount int
	var errorCount int

	for i := 0; i < 10; i++ {
		go func(amount int64) {
			err := suite.updateBalance(walletID, models.Deposit, amount)
			if err != nil {
				errorCount++
			} else {
				successCount++
			}
		}(int64((i + 1) * 100))
	}

	time.Sleep(2 * time.Second)

	assert.Equal(suite.T(), 10, successCount)
	assert.Equal(suite.T(), 0, errorCount)

	balance, err := suite.getBalance(walletID)
	assert.NoError(suite.T(), err)
	
	assert.Equal(suite.T(), int64(15500), balance)
}

func (suite *WalletAPITestSuite) getBalance(walletID uuid.UUID) (int64, error) {
	resp, err := suite.client.Get(suite.baseURL + "/api/v1/wallets/" + walletID.String())
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var result struct {
		Balance int64 `json:"balance"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result.Balance, err
}

func (suite *WalletAPITestSuite) updateBalance(walletID uuid.UUID, opType models.OperationType, amount int64) error {
	reqBody := models.OperationRequest{
		WalletID:      walletID,
		OperationType: opType,
		Amount:        amount,
	}

	body, _ := json.Marshal(reqBody)
	resp, err := suite.client.Post(suite.baseURL+"/api/v1/wallet", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return nil
}

func TestWalletAPITestSuite(t *testing.T) {
	suite.Run(t, new(WalletAPITestSuite))
}