'use client';

import { useStore } from '@/lib/store';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine, Area, AreaChart } from 'recharts';
import { TrendingUp } from 'lucide-react';

export default function SpreadChart() {
	const chartData = useStore((state) => state.chartData);

	const gradientOffset = () => {
		const dataMax = Math.max(...chartData.map((i) => i.spread));
		const dataMin = Math.min(...chartData.map((i) => i.spread));

		if (dataMax <= 0) {
			return 0;
		}
		if (dataMin >= 0) {
			return 1;
		}

		return dataMax / (dataMax - dataMin);
	};

	const off = gradientOffset();

	return (
		<div className="h-full flex flex-col bg-gradient-to-br from-slate-900 via-slate-900 to-slate-800 rounded-xl border border-slate-700/50 overflow-hidden shadow-2xl">
			{/* Header with gradient */}
			<div className="relative px-4 py-3 bg-gradient-to-r from-slate-800 to-slate-900 border-b border-slate-700/50">
				<div className="absolute inset-0 bg-gradient-to-r from-purple-500/5 to-blue-500/5" />
				<div className="relative flex items-center gap-2">
					<TrendingUp className="w-4 h-4 text-purple-400" />
					<h2 className="text-sm font-bold text-white uppercase tracking-wider">Spread Analysis</h2>
					<span className="text-xs text-slate-500 ml-auto font-mono">Last 50 Opportunities</span>
				</div>
			</div>

			<div className="flex-1 p-4 min-h-[300px]">
				<ResponsiveContainer width="100%" height="100%">
					<AreaChart data={chartData}>
						<defs>
							<linearGradient id="splitColor" x1="0" y1="0" x2="0" y2="1">
								<stop offset={off} stopColor="#10b981" stopOpacity={0.8} />
								<stop offset={off} stopColor="#ef4444" stopOpacity={0.8} />
							</linearGradient>
							<linearGradient id="fillGradient" x1="0" y1="0" x2="0" y2="1">
								<stop offset={off} stopColor="#10b981" stopOpacity={0.2} />
								<stop offset={off} stopColor="#ef4444" stopOpacity={0.2} />
							</linearGradient>
						</defs>
						<CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} opacity={0.3} />
						<XAxis
							dataKey="block"
							stroke="#64748b"
							tick={{ fontSize: 11, fill: '#94a3b8' }}
							tickLine={{ stroke: '#475569' }}
							domain={['auto', 'auto']}
							type="number"
						/>
						<YAxis
							stroke="#64748b"
							tick={{ fontSize: 11, fill: '#94a3b8' }}
							tickLine={{ stroke: '#475569' }}
							unit="%"
							domain={['auto', 'auto']}
						/>
						<Tooltip
							contentStyle={{
								backgroundColor: '#0f172a',
								borderColor: '#334155',
								borderRadius: '8px',
								boxShadow: '0 10px 40px rgba(0,0,0,0.5)'
							}}
							labelStyle={{ color: '#94a3b8', marginBottom: '4px' }}
							formatter={(value: number | undefined) => {
								if (value === undefined) return ['', ''];
								const color = value > 0 ? '#34d399' : value < 0 ? '#ef4444' : '#ffffff';
								return [
									<span style={{ color, fontWeight: 600 }}>{value.toFixed(3)}%</span>,
									'Spread'
								];
							}}
						/>
						<ReferenceLine y={0} stroke="#94a3b8" strokeDasharray="3 3" strokeWidth={2} />
						<Area
							type="monotone"
							dataKey="spread"
							stroke="url(#splitColor)"
							fill="url(#fillGradient)"
							strokeWidth={3}
							isAnimationActive={false}
						/>
					</AreaChart>
				</ResponsiveContainer>
			</div>
		</div>
	);
}
