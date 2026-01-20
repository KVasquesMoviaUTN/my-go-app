'use client';

import Header from '@/components/dashboard/Header';
import LiveFeed from '@/components/dashboard/LiveFeed';
import SpreadChart from '@/components/dashboard/SpreadChart';
import { useArbitrageSocket } from '@/hooks/useArbitrageSocket';

export default function Dashboard() {
  // Initialize WebSocket connection (or Mock Mode)
  useArbitrageSocket();

  return (
    <div className="min-h-screen bg-slate-950 text-slate-200 font-sans selection:bg-emerald-500/30">
      <Header />

      <main className="p-6 grid grid-cols-1 lg:grid-cols-2 gap-6 h-[calc(100vh-80px)]">
        {/* Left Column: Chart */}
        <div className="h-[400px] lg:h-auto">
          <SpreadChart />
        </div>

        {/* Right Column: Feed */}
        <div className="h-[400px] lg:h-auto">
          <LiveFeed />
        </div>
      </main>
    </div>
  );
}
