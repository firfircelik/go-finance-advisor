import { useEffect, useState } from 'react';
import { TrendingUp, TrendingDown, DollarSign, Activity, ArrowUpRight, Loader2 } from 'lucide-react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import api from '../api/client';

interface OverviewData {
    total_balance: number;
    total_profit: number;
    active_positions: number;
    yield_percentage: number;
}

interface PerformanceData {
    date: string;
    value: number;
}

interface Transaction {
    id: string;
    type: string;
    asset: string;
    amount: number;
    date: string;
}

const Dashboard = () => {
    const [overview, setOverview] = useState<OverviewData | null>(null);
    const [performance, setPerformance] = useState<PerformanceData[]>([]);
    const [transactions, setTransactions] = useState<Transaction[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const [overviewRes, performanceRes, transactionsRes] = await Promise.all([
                    api.get('/overview'),
                    api.get('/performance'),
                    api.get('/transactions')
                ]);

                setOverview(overviewRes.data);
                setPerformance(performanceRes.data);
                setTransactions(transactionsRes.data);
            } catch (error) {
                console.error('Failed to fetch dashboard data:', error);
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
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold font-heading mb-2">Dashboard</h1>
                    <p className="text-muted-foreground">Welcome back. Here's what's happening with your portfolio.</p>
                </div>
                <div className="flex gap-3">
                    <button className="px-4 py-2 bg-secondary/50 border border-white/10 rounded-lg text-sm font-medium hover:bg-secondary transition-colors">
                        Download Report
                    </button>
                    <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors shadow-lg shadow-primary/20">
                        Add Funds
                    </button>
                </div>
            </div>

            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard
                    title="Total Balance"
                    value={`$${overview?.total_balance.toLocaleString()}`}
                    change="+12.5%"
                    trend="up"
                    icon={<DollarSign className="w-5 h-5 text-emerald-500" />}
                />
                <StatCard
                    title="Total Profit"
                    value={`$${overview?.total_profit.toLocaleString()}`}
                    change="+5.2%"
                    trend="up"
                    icon={<TrendingUp className="w-5 h-5 text-blue-500" />}
                />
                <StatCard
                    title="Active Positions"
                    value={overview?.active_positions.toString() || '0'}
                    change="-2"
                    trend="down"
                    icon={<Activity className="w-5 h-5 text-purple-500" />}
                />
                <StatCard
                    title="Yield"
                    value={`${overview?.yield_percentage}%`}
                    change="+1.2%"
                    trend="up"
                    icon={<ArrowUpRight className="w-5 h-5 text-orange-500" />}
                />
            </div>

            {/* Main Chart Section */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                <div className="lg:col-span-2 bg-card/50 backdrop-blur-xl border border-white/5 rounded-2xl p-6">
                    <div className="flex items-center justify-between mb-6">
                        <h3 className="text-lg font-semibold">Portfolio Performance</h3>
                        <div className="flex gap-2">
                            {['1D', '1W', '1M', '1Y', 'ALL'].map((period) => (
                                <button key={period} className="px-3 py-1 rounded-lg text-xs font-medium hover:bg-secondary/50 transition-colors">
                                    {period}
                                </button>
                            ))}
                        </div>
                    </div>
                    <div className="h-[300px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={performance}>
                                <defs>
                                    <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                                        <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                                    </linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" stroke="#ffffff10" vertical={false} />
                                <XAxis dataKey="date" stroke="#64748b" fontSize={12} tickLine={false} axisLine={false} />
                                <YAxis stroke="#64748b" fontSize={12} tickLine={false} axisLine={false} tickFormatter={(value) => `$${value}`} />
                                <Tooltip
                                    contentStyle={{ backgroundColor: '#0f172a', borderColor: '#1e293b', borderRadius: '8px' }}
                                    itemStyle={{ color: '#fff' }}
                                />
                                <Area type="monotone" dataKey="value" stroke="#3b82f6" strokeWidth={3} fillOpacity={1} fill="url(#colorValue)" />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                </div>

                <div className="bg-card/50 backdrop-blur-xl border border-white/5 rounded-2xl p-6">
                    <h3 className="text-lg font-semibold mb-6">Recent Activity</h3>
                    <div className="space-y-6">
                        {transactions.map((tx) => (
                            <ActivityItem
                                key={tx.id}
                                title={`${tx.type === 'buy' ? 'Bought' : 'Sold'} ${tx.asset}`}
                                subtitle={tx.type.toUpperCase()}
                                amount={`${tx.type === 'buy' ? '-' : '+'}$${tx.amount.toLocaleString()}`}
                                date={new Date(tx.date).toLocaleDateString()}
                                icon={
                                    <div className={`w-10 h-10 rounded-full flex items-center justify-center ${tx.type === 'buy' ? 'bg-blue-500/20 text-blue-500' : 'bg-emerald-500/20 text-emerald-500'}`}>
                                        {tx.type === 'buy' ? <TrendingUp size={18} /> : <TrendingDown size={18} />}
                                    </div>
                                }
                            />
                        ))}
                        {transactions.length === 0 && (
                            <p className="text-center text-muted-foreground text-sm">No recent transactions</p>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

const StatCard = ({ title, value, change, trend, icon }: { title: string, value: string, change: string, trend: 'up' | 'down', icon: React.ReactNode }) => (
    <div className="bg-card/50 backdrop-blur-xl border border-white/5 p-6 rounded-2xl hover:border-primary/20 transition-colors group">
        <div className="flex justify-between items-start mb-4">
            <div className="p-2 bg-secondary/30 rounded-lg group-hover:bg-primary/10 transition-colors">
                {icon}
            </div>
            <span className={`text-xs font-bold px-2 py-1 rounded-full ${trend === 'up' ? 'bg-emerald-500/10 text-emerald-500' : 'bg-red-500/10 text-red-500'}`}>
                {change}
            </span>
        </div>
        <h3 className="text-muted-foreground text-sm font-medium mb-1">{title}</h3>
        <p className="text-2xl font-bold font-heading">{value}</p>
    </div>
);

const ActivityItem = ({ title, subtitle, amount, date, icon }: { title: string, subtitle: string, amount: string, date: string, icon: React.ReactNode }) => (
    <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
            {icon}
            <div>
                <h4 className="font-semibold text-sm">{title}</h4>
                <p className="text-xs text-muted-foreground">{subtitle}</p>
            </div>
        </div>
        <div className="text-right">
            <p className={`font-bold text-sm ${amount.startsWith('+') ? 'text-emerald-500' : 'text-foreground'}`}>{amount}</p>
            <p className="text-xs text-muted-foreground">{date}</p>
        </div>
    </div>
);

export default Dashboard;
