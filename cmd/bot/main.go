package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/KVasquesMoviaUTN/cex-dex-arbitrage-challenge/internal/core/services"
	"github.com/KVasquesMoviaUTN/cex-dex-arbitrage-challenge/internal/engine"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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

	tradeSizesStr := viper.GetString("TRADE_SIZES")
	tradeSizes, err := engine.ParseTradeSizes(tradeSizesStr)
	if err != nil {
		slog.Warn("Failed to parse some trade sizes", "error", err)
	}
	if len(tradeSizes) == 0 {
		log.Fatal("No valid TRADE_SIZES configured")
	}

	minProfitStr := viper.GetString("MIN_PROFIT")
	minProfit, err := decimal.NewFromString(minProfitStr)
	if err != nil {
		log.Fatalf("Invalid MIN_PROFIT: %v", err)
	}

	cfg := engine.Config{
		Config: services.Config{
			Symbol:        viper.GetString("SYMBOL"),
			TokenInAddr:   viper.GetString("TOKEN_IN"),
			TokenOutAddr:  viper.GetString("TOKEN_OUT"),
			TokenInDec:    viper.GetInt32("TOKEN_IN_DEC"),
			TokenOutDec:   viper.GetInt32("TOKEN_OUT_DEC"),
			PoolFee:       viper.GetInt64("POOL_FEE"),
			MaxWorkers:    viper.GetInt("MAX_WORKERS"),
			CacheDuration: 10 * time.Second,
			TradeSizes:    tradeSizes,
			MinProfit:     minProfit,
		},
		EthNodeWS:     viper.GetString("ETH_NODE_WS"),
		EthNodeHTTP:   viper.GetString("ETH_NODE_HTTP"),
		MetricsPort:   viper.GetString("METRICS_PORT"),
		CEXProvider:   viper.GetString("CEX_PROVIDER"),
		BinanceAPIURL: viper.GetString("BINANCE_API_URL"),
	}

	eng, err := engine.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

	if err := eng.Run(context.Background()); err != nil {
		slog.Error("Engine exited with error", "error", err)
		os.Exit(1)
	}
}
