'use client';

import { useStore } from '@/lib/store';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from 'recharts';

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
		<div className="h-full flex flex-col bg-slate-950 rounded-lg border border-slate-800 overflow-hidden">
			<div className="px-4 py-2 bg-slate-900 border-b border-slate-800">
				<h2 className="text-sm font-semibold text-slate-300 uppercase tracking-wider">Spread Analysis (Last 50 Ops)</h2>
			</div>

			<div className="flex-1 p-4 min-h-[300px]">
				<ResponsiveContainer width="100%" height="100%">
					<LineChart data={chartData}>
						<CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
						<XAxis
							dataKey="block"
							stroke="#64748b"
							tick={{ fontSize: 10 }}
							domain={['auto', 'auto']}
							type="number"
						/>
						<YAxis
							stroke="#64748b"
							tick={{ fontSize: 10 }}
							unit="%"
							domain={['auto', 'auto']}
						/>
						<Tooltip
							contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', color: '#f8fafc' }}
							itemStyle={{ color: '#34d399' }}
							labelStyle={{ color: '#94a3b8' }}
						/>
						<defs>
							<linearGradient id="splitColor" x1="0" y1="0" x2="0" y2="1">
								<stop offset={off} stopColor="#10b981" stopOpacity={1} />
								<stop offset={off} stopColor="#ef4444" stopOpacity={1} />
							</linearGradient>
						</defs>
						<ReferenceLine y={0} stroke="#94a3b8" strokeDasharray="3 3" />
						<Line
							type="monotone"
							dataKey="spread"
							stroke="url(#splitColor)"
							strokeWidth={2}
							dot={false}
							activeDot={{ r: 4, fill: '#10b981' }}
							isAnimationActive={false} // Disable animation for performance
						/>
					</LineChart>
				</ResponsiveContainer>
			</div>
		</div>
	);
}
