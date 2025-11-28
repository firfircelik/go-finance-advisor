import { useEffect, useState } from 'react';
import { Loader2, Search, TrendingUp, TrendingDown } from 'lucide-react';
import api from '../api/client';

interface TickerItem {
    s: string; // Symbol
    p: number; // Price
    c: number; // Change
    v: number; // Volume
    h: number; // High
    l: number; // Low
}

const Markets = () => {
    const [items, setItems] = useState<TickerItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [search, setSearch] = useState('');

    useEffect(() => {
        const fetchTicker = async () => {
            try {
                const response = await api.get('/ticker');
                setItems(response.data);
            } catch (error) {
                console.error('Failed to fetch market data:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchTicker();
        const interval = setInterval(fetchTicker, 5000);
        return () => clearInterval(interval);
    }, []);

    const filteredItems = items.filter(item =>
        item.s.toLowerCase().includes(search.toLowerCase())
    );

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
                <h1 className="text-3xl font-bold font-heading">Markets</h1>
                <div className="relative w-64">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                    <input
                        type="text"
                        placeholder="Search symbol..."
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                        className="w-full bg-secondary/30 border border-white/10 rounded-xl py-2 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
                    />
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {filteredItems.map((item) => (
                    <div key={item.s} className="bg-card/50 backdrop-blur-xl border border-white/5 rounded-xl p-6 hover:border-primary/20 transition-all cursor-pointer group">
                        <div className="flex justify-between items-start mb-4">
                            <div>
                                <h3 className="font-bold text-lg">{item.s}</h3>
                                <p className="text-xs text-muted-foreground">Stock</p>
                            </div>
                            <div className={`flex items-center gap-1 text-sm font-bold px-2 py-1 rounded ${item.c >= 0 ? 'bg-emerald-500/10 text-emerald-500' : 'bg-red-500/10 text-red-500'}`}>
                                {item.c >= 0 ? <TrendingUp size={14} /> : <TrendingDown size={14} />}
                                {item.c >= 0 ? '+' : ''}{item.c.toFixed(2)}%
                            </div>
                        </div>
                        <div className="flex justify-between items-end">
                            <div>
                                <p className="text-2xl font-bold">${item.p.toFixed(2)}</p>
                            </div>
                            <div className="text-right text-xs text-muted-foreground">
                                <p>Vol: {(item.v / 1000).toFixed(1)}K</p>
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default Markets;
