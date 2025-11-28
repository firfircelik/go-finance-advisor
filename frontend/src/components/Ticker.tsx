import { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import api from '../api/client';

interface TickerItem {
    s: string; // Symbol
    p: number; // Price
    c: number; // Change
}

const Ticker = () => {
    const [items, setItems] = useState<TickerItem[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchTicker = async () => {
            try {
                const response = await api.get('/ticker');
                setItems(response.data);
            } catch (error) {
                console.error('Failed to fetch ticker data:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchTicker();
        const interval = setInterval(fetchTicker, 5000); // Refresh every 5 seconds

        return () => clearInterval(interval);
    }, []);

    if (loading && items.length === 0) return null;

    return (
        <div className="w-full bg-background/50 backdrop-blur-md border-b border-white/5 overflow-hidden py-3 z-50 relative">
            <motion.div
                className="flex whitespace-nowrap"
                animate={{ x: [0, -1000] }}
                transition={{
                    repeat: Infinity,
                    duration: 40,
                    ease: "linear"
                }}
            >
                {[...items, ...items].map((item, index) => (
                    <div key={`${item.s}-${index}`} className="inline-flex items-center mx-8 group cursor-pointer">
                        <span className="font-bold mr-2 text-sm group-hover:text-primary transition-colors">{item.s}</span>
                        <span className="mr-2 text-sm font-medium">${item.p.toFixed(2)}</span>
                        <span className={`text-xs font-bold px-1.5 py-0.5 rounded ${item.c >= 0 ? 'bg-emerald-500/10 text-emerald-500' : 'bg-red-500/10 text-red-500'}`}>
                            {item.c >= 0 ? '+' : ''}{item.c.toFixed(2)}%
                        </span>
                    </div>
                ))}
            </motion.div>
        </div>
    );
};

export default Ticker;
