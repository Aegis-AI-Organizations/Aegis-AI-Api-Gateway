package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/api/handlers"
	agrpc "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc"
	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/agrpc/aegis/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBalanceHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockBilling := new(MockBillingServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			BillingService: mockBilling,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("company_id", "test-company")
	c.Request, _ = http.NewRequest("GET", "/billing/balance", nil)

	mockBilling.On("GetBalance", mock.Anything, &v1.GetBalanceRequest{CompanyId: "test-company"}).
		Return(&v1.GetBalanceResponse{CompanyId: "test-company", Balance: 100}, nil)

	api.GetBalanceHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp v1.GetBalanceResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), resp.Balance)
}

func TestAdjustTokensHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockBilling := new(MockBillingServiceClient)
	api := &handlers.API{
		GRPCClient: &agrpc.Client{
			BillingService: mockBilling,
		},
	}

	payload := map[string]interface{}{"amount": 50, "reason": "bonus"}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "test-company"}}
	c.Request, _ = http.NewRequest("POST", "/admin/companies/test-company/tokens/adjust", bytes.NewBuffer(body))

	mockBilling.On("AdjustTokens", mock.Anything, &v1.AdjustTokensRequest{
		CompanyId: "test-company",
		Amount:    50,
		Reason:    "bonus",
	}).Return(&v1.AdjustTokensResponse{CompanyId: "test-company", Balance: 150}, nil)

	api.AdjustTokensHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockBilling.AssertExpectations(t)
}
