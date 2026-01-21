export type ArbitrageEvent = {
  type: 'HEARTBEAT' | 'OPPORTUNITY';
  blockNumber: number;
  timestamp: string;
  data?: {
    cexPrice: number;
    dexPrice: number;
    spreadPct: number;
    estimatedProfit: number;
    gasCost: number;
    symbol: string; // e.g., "ETH-USDC"
    direction: string; // "CEX -> DEX" or "DEX -> CEX"
  }
}

export type DashboardState = {
  isConnected: boolean;
  lastBlock: number;
  latency: number; // ms
  events: ArbitrageEvent[];
  chartData: { block: number; spread: number }[];
  
  setConnected: (status: boolean) => void;
  updateBlock: (block: number) => void;
  addEvent: (event: ArbitrageEvent) => void;
  updateLatency: (ms: number) => void;
}
