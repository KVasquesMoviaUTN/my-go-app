# System Architecture

## Overview

The CEX-DEX Arbitrage Bot is designed as a high-performance, concurrent system that monitors price discrepancies between Binance (Centralized Exchange) and Uniswap V3 (Decentralized Exchange). It follows the **Hexagonal Architecture** (Ports & Adapters) pattern to decouple core business logic from external infrastructure.

## 1. Component Diagram (Hexagonal Architecture)

This diagram illustrates how the core logic is isolated from the outside world via Ports (Interfaces) and Adapters (Implementations).

```mermaid
graph TD
    subgraph "External World"
        Binance[Binance API]
        Ethereum[Ethereum Node (RPC/WS)]
        Prometheus[Prometheus]
    end

    subgraph "Application"
        subgraph "Adapters (Infrastructure)"
            BinanceAdapter[Binance Adapter]
            EthAdapter[Ethereum Adapter]
            Listener[Blockchain Listener]
            Metrics[Metrics Handler]
        end

        subgraph "Core (Business Logic)"
            Ports{Ports / Interfaces}
            Manager[Manager Service]
            Domain[Domain Models]
        end
    end

    %% Relationships
    Manager --> Ports
    BinanceAdapter ..|> Ports : Implements
    EthAdapter ..|> Ports : Implements
    Listener ..|> Ports : Implements

    BinanceAdapter <--> Binance
    EthAdapter <--> Ethereum
    Listener <--> Ethereum
    Metrics <--> Prometheus

    Manager --> Domain
    Ports --> Domain
```

## 2. Class Diagram

This diagram details the internal structure of the Go codebase, showing the relationships between the Manager, Interfaces, and Domain entities.

```mermaid
classDiagram
    %% Domain
    class OrderBook {
        +Asks []PriceLevel
        +Bids []PriceLevel
        +Timestamp Time
        +CalculateEffectivePrice(side, amount) (Decimal, bool)
    }
    class PriceQuote {
        +Price Decimal
        +GasEstimate BigInt
        +Timestamp Time
    }
    class ArbitrageOpportunity {
        +BuyOn String
        +SellOn String
        +Profit Decimal
    }

    %% Ports
    class ExchangeAdapter {
        <<interface>>
        +GetOrderBook(symbol) (*OrderBook, error)
    }
    class PriceProvider {
        <<interface>>
        +GetQuote(tokenIn, tokenOut, amount, fee) (*PriceQuote, error)
    }
    class BlockchainListener {
        <<interface>>
        +SubscribeNewHeads() (<-chan BigInt, <-chan error, error)
    }

    %% Core
    class Manager {
        -cfg Config
        -cex ExchangeAdapter
        -dex PriceProvider
        -listener BlockchainListener
        -sem chan struct{}
        +Start(ctx) error
        -processBlock(ctx, blockNum)
        -checkArbitrageForSize(ctx, ob, amount)
    }

    %% Adapters
    class BinanceAdapter {
        -client HTTPClient
        +GetOrderBook(symbol) (*OrderBook, error)
    }
    class EthereumAdapter {
        -client EthClient
        +GetQuote(...) (*PriceQuote, error)
    }
    class ListenerImpl {
        -clientURL String
        +SubscribeNewHeads()
    }

    %% Relationships
    Manager --> ExchangeAdapter
    Manager --> PriceProvider
    Manager --> BlockchainListener
    Manager ..> OrderBook : Uses
    Manager ..> PriceQuote : Uses
    Manager ..> ArbitrageOpportunity : Creates

    BinanceAdapter ..|> ExchangeAdapter
    EthereumAdapter ..|> PriceProvider
    ListenerImpl ..|> BlockchainListener
```

## 3. Sequence Diagram (Block Processing Flow)

This diagram shows the runtime flow of the system when a new block is received from the Ethereum network.

```mermaid
sequenceDiagram
    participant Eth as Ethereum Node
    participant L as Listener
    participant M as Manager
    participant CEX as Binance Adapter
    participant DEX as Ethereum Adapter
    participant Metrics as Observability

    Note over Eth, Metrics: System Startup
    M->>L: SubscribeNewHeads()
    L->>Eth: eth_subscribe("newHeads")
    Eth-->>L: Subscription ID

    Note over Eth, Metrics: Runtime Loop
    loop Every Block (~12s)
        Eth->>L: New Block Header (Block N)
        L->>M: Block Number N
        
        M->>Metrics: Inc(BlocksProcessed)
        
        alt Worker Pool Full
            M->>M: Drop Block
        else Worker Available
            M->>M: Acquire Semaphore
            par Fetch CEX
                M->>CEX: GetOrderBook("ETHUSDC")
                CEX->>Binance API: GET /depth
                Binance API-->>CEX: JSON Response
                CEX-->>M: OrderBook
            and Fetch DEX
                M->>DEX: GetQuote(1 ETH)
                DEX->>Eth: eth_call(QuoterV2)
                Eth-->>DEX: Result (AmountOut)
                DEX-->>M: PriceQuote
            end

            M->>M: CalculateEffectivePrice(CEX)
            M->>M: Calculate Profit (DEX - CEX - Fees)

            alt Profit > MinProfit
                M->>Metrics: Inc(ArbitrageOpsFound)
                M->>M: Log Opportunity
            end
            
            M->>M: Release Semaphore
        end
    end
```

## 4. Directory Structure

- **`cmd/bot`**: The entry point. Initializes adapters, loads config, and starts the Manager.
- **`internal/core/domain`**: Pure business entities. No external dependencies.
- **`internal/core/ports`**: Interfaces defining the contract between Core and Adapters.
- **`internal/core/services`**: The `Manager` contains the orchestration logic.
- **`internal/adapters`**: Implementations of the ports (Binance, Ethereum, Blockchain).
- **`internal/observability`**: Prometheus metrics definitions.
- **`docs`**: This documentation.

## 5. Design Patterns Used

### 1. Hexagonal Architecture (Ports & Adapters)
- **Pattern**: Separation of core business logic (`internal/core`) from external infrastructure (`internal/adapters`).
- **Usage**: The `Manager` depends only on interfaces (`ExchangeAdapter`, `PriceProvider`), not on concrete implementations like `BinanceAdapter`. This allows easy swapping of exchanges or mocking for tests.

### 2. Worker Pool Pattern
- **Pattern**: Limiting the number of concurrent tasks to prevent resource exhaustion.
- **Usage**: The `Manager` uses a buffered channel semaphore (`sem`) to limit the number of blocks processed concurrently. If the pool is full, new blocks are dropped to prioritize freshness over completeness (crucial for arbitrage).

### 3. Dependency Injection
- **Pattern**: Passing dependencies to an object rather than creating them internally.
- **Usage**: The `NewManager` constructor accepts `ExchangeAdapter`, `PriceProvider`, and `BlockchainListener` as arguments. This facilitates testing by allowing the injection of mocks.

### 4. Adapter Pattern
- **Pattern**: Converting the interface of a class into another interface clients expect.
- **Usage**: `BinanceAdapter` and `EthereumAdapter` adapt the specific APIs of Binance and Uniswap into the generic `ExchangeAdapter` and `PriceProvider` interfaces used by the domain.

### 5. Graceful Shutdown
- **Pattern**: Catching OS signals to clean up resources before exiting.
- **Usage**: The `main` function listens for `SIGINT` and `SIGTERM` to cancel the context, allowing the `Manager` and `Listener` to stop processing and close connections cleanly.
