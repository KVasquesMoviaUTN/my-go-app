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
    // Check for duplicate (same data as last event)
    const lastEvent = state.events[0];
    const isDuplicate = lastEvent &&
      event.type === 'OPPORTUNITY' &&
      lastEvent.type === 'OPPORTUNITY' &&
      event.data?.symbol === lastEvent.data?.symbol &&
      event.data?.direction === lastEvent.data?.direction &&
      Math.abs((event.data?.spreadPct || 0) - (lastEvent.data?.spreadPct || 0)) < 0.0001 &&
      Math.abs((event.data?.cexPrice || 0) - (lastEvent.data?.cexPrice || 0)) < 0.01 &&
      Math.abs((event.data?.dexPrice || 0) - (lastEvent.data?.dexPrice || 0)) < 0.01;

    if (isDuplicate) {
      return {
        lastBlock: Math.max(state.lastBlock, event.blockNumber)
      };
    }

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
