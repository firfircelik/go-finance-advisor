import { useEffect, useState } from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts';
import { Loader2, TrendingUp, TrendingDown } from 'lucide-react';
import api from '../api/client';

interface Holding {
    symbol: string;
    quantity: number;
    avg_price: number;
    current_price: number;
    total_value: number;
    gain_loss: number;
    gain_loss_percent: number;
}

interface Allocation {
    name: string;
    value: number;
}

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8'];

const Portfolio = () => {
    const [holdings, setHoldings] = useState<Holding[]>([]);
    const [allocation, setAllocation] = useState<Allocation[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [holdingsRes, allocationRes] = await Promise.all([
                    api.get('/holdings'),
                    api.get('/allocation')
                ]);
                setHoldings(holdingsRes.data);
                setAllocation(allocationRes.data);
            } catch (error) {
                console.error('Failed to fetch portfolio data:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

    if (loading) {
        return (
            <div className="flex items-center justify-center h-full">
                <Loader2 className="w-8 h-8 animate-spin text-primary" />
            </div>
        );
    }

    return (
        <div className="p-8 space-y-8">
            <h1 className="text-3xl font-bold font-heading">Portfolio</h1>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Asset Allocation Chart */}
                <div className="bg-card/50 backdrop-blur-xl border border-white/5 rounded-2xl p-6">
                    <h3 className="text-lg font-semibold mb-6">Asset Allocation</h3>
                    <div className="h-[300px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <PieChart>
                                <Pie
                                    data={allocation}
                                    cx="50%"
                                    cy="50%"
                                    innerRadius={60}
                                    outerRadius={80}
                                    fill="#8884d8"
                                    paddingAngle={5}
                                    dataKey="value"
                                >
                                    {allocation.map((entry, index) => (
                                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                                    ))}
                                </Pie>
                                <Tooltip
                                    contentStyle={{ backgroundColor: '#0f172a', borderColor: '#1e293b', borderRadius: '8px' }}
                                    itemStyle={{ color: '#fff' }}
                                />
                            </PieChart>
                        </ResponsiveContainer>
                    </div>
                    <div className="flex flex-wrap justify-center gap-4 mt-4">
                        {allocation.map((entry, index) => (
                            <div key={entry.name} className="flex items-center gap-2 text-sm">
                                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: COLORS[index % COLORS.length] }} />
                                <span className="text-muted-foreground">{entry.name}</span>
                            </div>
                        ))}
                    </div>
                </div>

                {/* Holdings Table */}
                <div className="lg:col-span-2 bg-card/50 backdrop-blur-xl border border-white/5 rounded-2xl p-6 overflow-hidden">
                    <h3 className="text-lg font-semibold mb-6">Current Holdings</h3>
                    <div className="overflow-x-auto">
                        <table className="w-full">
                            <thead>
                                <tr className="border-b border-white/10 text-left">
                                    <th className="pb-4 font-medium text-muted-foreground">Asset</th>
                                    <th className="pb-4 font-medium text-muted-foreground text-right">Quantity</th>
                                    <th className="pb-4 font-medium text-muted-foreground text-right">Avg. Price</th>
                                    <th className="pb-4 font-medium text-muted-foreground text-right">Current Price</th>
                                    <th className="pb-4 font-medium text-muted-foreground text-right">Total Value</th>
                                    <th className="pb-4 font-medium text-muted-foreground text-right">Gain/Loss</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-white/5">
                                {holdings.map((holding) => (
                                    <tr key={holding.symbol} className="group hover:bg-white/5 transition-colors">
                                        <td className="py-4 font-bold">{holding.symbol}</td>
                                        <td className="py-4 text-right">{holding.quantity}</td>
                                        <td className="py-4 text-right text-muted-foreground">${holding.avg_price.toFixed(2)}</td>
                                        <td className="py-4 text-right">${holding.current_price.toFixed(2)}</td>
                                        <td className="py-4 text-right font-medium">${holding.total_value.toLocaleString()}</td>
                                        <td className={`py-4 text-right font-medium ${holding.gain_loss >= 0 ? 'text-emerald-500' : 'text-red-500'}`}>
                                            <div className="flex items-center justify-end gap-1">
                                                {holding.gain_loss >= 0 ? <TrendingUp size={14} /> : <TrendingDown size={14} />}
                                                {holding.gain_loss >= 0 ? '+' : ''}{holding.gain_loss_percent.toFixed(2)}%
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Portfolio;
