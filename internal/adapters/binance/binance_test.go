package binance

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGetOrderBook(t *testing.T) {
	// Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Request
		if r.URL.Path != "/depth" {
			t.Errorf("Expected path /depth, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "ETHUSDC" {
			t.Errorf("Expected symbol ETHUSDC, got %s", r.URL.Query().Get("symbol"))
		}

		// Mock Response
		response := `{
			"lastUpdateId": 1027024,
			"bids": [
				["4.00000000", "431.00000000"]
			],
			"asks": [
				["4.00000200", "12.00000000"]
			]
		}`
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, response)
	}))
	defer ts.Close()

	// Initialize Adapter with Mock Server URL
	adapter := NewAdapter(ts.URL)

	// Test
	ctx := context.Background()
	ob, err := adapter.GetOrderBook(ctx, "ETHUSDC")

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, ob)
	assert.Len(t, ob.Bids, 1)
	assert.Len(t, ob.Asks, 1)

	expectedBidPrice := decimal.NewFromFloat(4.0)
	expectedBidAmt := decimal.NewFromFloat(431.0)
	assert.True(t, ob.Bids[0].Price.Equal(expectedBidPrice), "Bid price mismatch")
	assert.True(t, ob.Bids[0].Amount.Equal(expectedBidAmt), "Bid amount mismatch")

	expectedAskPrice := decimal.NewFromFloat(4.000002)
	expectedAskAmt := decimal.NewFromFloat(12.0)
	assert.True(t, ob.Asks[0].Price.Equal(expectedAskPrice), "Ask price mismatch")
	assert.True(t, ob.Asks[0].Amount.Equal(expectedAskAmt), "Ask amount mismatch")
}

func TestGetOrderBook_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	adapter := NewAdapter(ts.URL)
	_, err := adapter.GetOrderBook(context.Background(), "ETHUSDC")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "binance api returned status: 500")
}
