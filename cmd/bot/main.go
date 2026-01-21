package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/KVasquesMoviaUTN/arbitrage-bot-go/internal/adapters/binance"
	"github.com/KVasquesMoviaUTN/arbitrage-bot-go/internal/adapters/blockchain"
	"github.com/KVasquesMoviaUTN/arbitrage-bot-go/internal/adapters/ethereum"
	"github.com/KVasquesMoviaUTN/arbitrage-bot-go/internal/adapters/kraken"
	"github.com/KVasquesMoviaUTN/arbitrage-bot-go/internal/adapters/okx"
	"github.com/KVasquesMoviaUTN/arbitrage-bot-go/internal/adapters/websocket"
	"github.com/KVasquesMoviaUTN/arbitrage-bot-go/internal/core/ports"
	"github.com/KVasquesMoviaUTN/arbitrage-bot-go/internal/core/services"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	go func() {
		port := viper.GetString("METRICS_PORT")
		addr := ":" + port
		http.Handle("/metrics", promhttp.Handler())
		slog.Info("Starting metrics server", "port", port)
		if err := http.ListenAndServe(addr, nil); err != nil {
			slog.Error("Metrics server failed", "error", err)
		}
	}()

	viper.SetDefault("ETH_NODE_WS", "wss://mainnet.infura.io/ws/v3/YOUR_KEY")
	viper.SetDefault("ETH_NODE_HTTP", "https://mainnet.infura.io/v3/YOUR_KEY")
	viper.SetDefault("SYMBOL", "ETHUSDC")
	viper.SetDefault("TOKEN_IN", "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	viper.SetDefault("TOKEN_OUT", "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	viper.SetDefault("TOKEN_IN_DEC", 18)
	viper.SetDefault("TOKEN_OUT_DEC", 6)
	viper.SetDefault("POOL_FEE", 3000)
	viper.SetDefault("TRADE_SIZES", "1000000000000000000,10000000000000000000")
	viper.SetDefault("MIN_PROFIT", "10.0")
	viper.SetDefault("MAX_WORKERS", 5)
	viper.SetDefault("METRICS_PORT", "8085")
	viper.SetDefault("CEX_PROVIDER", "binance")
	viper.SetDefault("BINANCE_API_URL", "https://api.binance.com/api/v3")

	viper.AutomaticEnv()

	cfg := services.Config{
		Symbol:        viper.GetString("SYMBOL"),
		TokenInAddr:   viper.GetString("TOKEN_IN"),
		TokenOutAddr:  viper.GetString("TOKEN_OUT"),
		TokenInDec:    viper.GetInt32("TOKEN_IN_DEC"),
		TokenOutDec:   viper.GetInt32("TOKEN_OUT_DEC"),
		PoolFee:       viper.GetInt64("POOL_FEE"),
		MaxWorkers:    viper.GetInt("MAX_WORKERS"),
		CacheDuration: 10 * time.Second,
	}

	tradeSizesStr := viper.GetString("TRADE_SIZES")
	tradeSizes, err := parseTradeSizes(tradeSizesStr)
	if err != nil {
		slog.Warn("Failed to parse some trade sizes", "error", err)
	}
	if len(tradeSizes) == 0 {
		log.Fatal("No valid TRADE_SIZES configured")
	}
	cfg.TradeSizes = tradeSizes

	minProfitStr := viper.GetString("MIN_PROFIT")
	minProfit, err := decimal.NewFromString(minProfitStr)
	if err != nil {
		log.Fatalf("Invalid MIN_PROFIT: %v", err)
	}
	cfg.MinProfit = minProfit

	cexProvider := viper.GetString("CEX_PROVIDER")
	cex := createCEXAdapter(cexProvider)
	slog.Info("Using CEX provider", "provider", cexProvider)

	ethNodeHTTP := viper.GetString("ETH_NODE_HTTP")
	dex, err := ethereum.NewAdapter(ethNodeHTTP)
	if err != nil {
		log.Fatalf("Failed to create DEX adapter: %v", err)
	}

	ethNodeWS := viper.GetString("ETH_NODE_WS")
	listener := blockchain.NewListener(ethNodeWS)

	notifier := websocket.NewServer()
	go func() {
		notifier.Start(":8080")
	}()

	manager := services.NewManager(cfg, cex, dex, listener, notifier)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Shutdown signal received")
		cancel()
	}()

	slog.Info("Starting arbitrage bot")
	if err := manager.Start(ctx); err != nil {
		slog.Error("Manager failed", "error", err)
	}
}

func createCEXAdapter(provider string) ports.ExchangeAdapter {
	switch strings.ToLower(provider) {
	case "kraken":
		return kraken.NewAdapter()
	case "okx":
		return okx.NewAdapter()
	case "binance":
		fallthrough
	default:
		return binance.NewAdapter(viper.GetString("BINANCE_API_URL"))
	}
}

func parseTradeSizes(s string) ([]*big.Int, error) {
	parts := strings.Split(s, ",")
	sizes := make([]*big.Int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		size, ok := new(big.Int).SetString(p, 10)
		if !ok {
			return sizes, fmt.Errorf("invalid trade size: %s", p)
		}
		sizes = append(sizes, size)
	}
	return sizes, nil
}
