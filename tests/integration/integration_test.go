package integration

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KVasquesMoviaUTN/cex-dex-arbitrage-challenge/internal/adapters/binance"
	"github.com/KVasquesMoviaUTN/cex-dex-arbitrage-challenge/internal/adapters/ethereum"
	"github.com/KVasquesMoviaUTN/cex-dex-arbitrage-challenge/internal/core/domain"
	"github.com/KVasquesMoviaUTN/cex-dex-arbitrage-challenge/internal/core/services"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBlockchainListener mocks the blockchain listener
type MockBlockchainListener struct {
	mock.Mock
}

func (m *MockBlockchainListener) SubscribeNewHeads(ctx context.Context) (<-chan *domain.Block, <-chan error, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan *domain.Block), args.Get(1).(<-chan error), args.Error(2)
}

// MockNotificationService mocks the notification service
type MockNotificationService struct {
	mock.Mock
	events chan domain.ArbitrageEvent
}

func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{
		events: make(chan domain.ArbitrageEvent, 10),
	}
}

func (m *MockNotificationService) Broadcast(event domain.ArbitrageEvent) {
	m.events <- event
}

func TestEndToEndArbitrageFlow(t *testing.T) {
	// 1. Mock Binance API
	binanceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/depth" {
			// Return a mock order book where ETH price is ~2000 USDC
			response := map[string]interface{}{
				"lastUpdateId": 12345,
				"bids": [][]string{
					{"2000.00", "10.0"}, // Buy at 2000
				},
				"asks": [][]string{
					{"2001.00", "10.0"}, // Sell at 2001
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer binanceServer.Close()

	// 2. Mock Ethereum Node (JSON-RPC)
	ethServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string        `json:"method"`
			Params []interface{} `json:"params"`
			ID     int           `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var result interface{}

		switch req.Method {
		case "eth_chainId":
			result = "0x1"
		case "eth_gasPrice":
			result = "0x4a817c800" // 20 Gwei
		case "eth_call":
			// Mock QuoterV2 response
			// We need to return encoded ABI data for (amountOut, sqrtPriceX96After, initializedTicksCrossed, gasEstimate)
			// Let's assume DEX price is significantly higher to trigger arbitrage (e.g., 2050 USDC)
			// 1 ETH -> 2050 USDC
			
			// Simple mock: we just return a successful call. 
			// In a real integration test, we might want to decode the input to return specific values.
			// For now, let's return a generic success response that the adapter can unpack.
			// The adapter expects 4 *big.Int.
			
			// 2050 * 10^6 = 2050000000 (AmountOut)
			// GasEstimate = 100000
			
			// We need to construct the hex string for the ABI encoded return values.
			// Since this is complex to construct manually without abi.Pack, we might need a helper or a simplified approach.
			// Alternatively, we can use the real ethereum adapter but point it to a mock that returns pre-calculated hex.
			
			// For this test, let's assume the adapter works if we give it valid hex.
			// 2050000000 = 0x7a31c700
			// 100000 = 0x186a0
			
			// 32 bytes per word
			// Word 1: AmountOut (2050000000)
			// Word 2: SqrtPriceX96After (0)
			// Word 3: InitializedTicksCrossed (0)
			// Word 4: GasEstimate (100000)
			
			result = "0x000000000000000000000000000000000000000000000000000000007a31c700" + // AmountOut
				"0000000000000000000000000000000000000000000000000000000000000000" + // SqrtPrice
				"0000000000000000000000000000000000000000000000000000000000000000" + // Ticks
				"00000000000000000000000000000000000000000000000000000000000186a0"   // GasEstimate
		default:
			result = "0x0"
		}

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  result,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ethServer.Close()

	// 3. Setup Dependencies
	cexAdapter := binance.NewAdapter(binanceServer.URL)
	dexAdapter, err := ethereum.NewAdapter(ethServer.URL)
	assert.NoError(t, err)

	mockListener := &MockBlockchainListener{}
	mockNotifier := NewMockNotificationService()

	// Channels for the listener
	blockChan := make(chan *domain.Block)
	errChan := make(chan error)
	mockListener.On("SubscribeNewHeads", mock.Anything).Return((<-chan *domain.Block)(blockChan), (<-chan error)(errChan), nil)

	// 4. Initialize Manager
	cfg := services.Config{
		Symbol:        "ETHUSDC",
		TokenInAddr:   "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
		TokenOutAddr:  "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // USDC
		TokenInDec:    18,
		TokenOutDec:   6,
		PoolFee:       3000,
		TradeSizes:    []*big.Int{big.NewInt(1000000000000000000)}, // 1 ETH
		MinProfit:     decimal.NewFromFloat(1.0),
		MaxWorkers:    1,
		CacheDuration: time.Second,
	}

	manager := services.NewManager(cfg, cexAdapter, dexAdapter, mockListener, mockNotifier)

	// 5. Start Manager in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := manager.Start(ctx)
		assert.NoError(t, err)
	}()

	// 6. Simulate a Block
	// Wait a bit for start
	time.Sleep(100 * time.Millisecond)

	block := &domain.Block{
		Number:    big.NewInt(1000),
		Timestamp: time.Now(),
	}
	blockChan <- block

	// 7. Verify Output
	// We expect a HEARTBEAT first, then potentially an OPPORTUNITY
	
	timeout := time.After(2 * time.Second)
	var heartbeatReceived, opportunityReceived bool

	for {
		select {
		case event := <-mockNotifier.events:
			if event.Type == "HEARTBEAT" {
				heartbeatReceived = true
			} else if event.Type == "OPPORTUNITY" {
				opportunityReceived = true
				// Verify Data
				assert.Equal(t, "ETHUSDC", event.Data.Symbol)
				assert.Equal(t, "CEX -> DEX", event.Data.Direction)
				// CEX Price: ~2001 (Ask)
				// DEX Price: ~2050 (from mock)
				// Spread should be positive
				assert.Greater(t, event.Data.SpreadPct, 0.0)
				assert.Greater(t, event.Data.EstimatedProfit, 0.0)
				
				// We can break if we found the opportunity
				goto Done
			}
		case <-timeout:
			t.Fatal("Timeout waiting for arbitrage event")
		}
	}

Done:
	assert.True(t, heartbeatReceived, "Should receive heartbeat")
	assert.True(t, opportunityReceived, "Should receive arbitrage opportunity")
}
