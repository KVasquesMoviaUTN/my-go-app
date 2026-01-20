import { create } from 'zustand';
import { DashboardState } from './types';

export const useStore = create<DashboardState>((set) => ({
  isConnected: false,
  lastBlock: 0,
  latency: 0,
  events: [],
  chartData: [],

  setConnected: (status) => set({ isConnected: status }),
  
  updateBlock: (block) => set({ lastBlock: block }),
  
  addEvent: (event) => set((state) => {
    // Keep last 100 events
    const newEvents = [event, ...state.events].slice(0, 100);
    
    // Update chart data if it's an opportunity
    let newChartData = state.chartData;
    if (event.type === 'OPPORTUNITY' && event.data) {
      newChartData = [...state.chartData, { 
        block: event.blockNumber, 
        spread: event.data.spreadPct 
      }].slice(-50); // Keep last 50 points
    }

    return { 
      events: newEvents,
      chartData: newChartData,
      lastBlock: Math.max(state.lastBlock, event.blockNumber)
    };
  }),

  updateLatency: (ms) => set({ latency: ms }),
}));
