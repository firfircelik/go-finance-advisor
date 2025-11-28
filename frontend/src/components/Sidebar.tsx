import { NavLink } from 'react-router-dom';
import { LayoutDashboard, PieChart, TrendingUp, Settings, LogOut, Wallet, Bell, Search } from 'lucide-react';

const Sidebar = () => {
    return (
        <aside className="w-72 border-r border-white/5 bg-card/50 backdrop-blur-xl flex flex-col h-screen sticky top-0">
            <div className="p-8">
                <div className="flex items-center gap-3 mb-8">
                    <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-600 to-purple-600 flex items-center justify-center text-white font-bold text-xl shadow-lg shadow-blue-500/20">
                        GF
                    </div>
                    <span className="text-xl font-bold font-heading tracking-tight">GoFinance</span>
                </div>

                <div className="relative mb-6">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                    <input
                        type="text"
                        placeholder="Search assets..."
                        className="w-full bg-secondary/30 border border-white/5 rounded-xl py-2.5 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 transition-all"
                    />
                </div>

                <nav className="space-y-2">
                    <SidebarItem to="/platform" end icon={<LayoutDashboard size={20} />} label="Dashboard" />
                    <SidebarItem to="/platform/portfolio" icon={<PieChart size={20} />} label="Portfolio" />
                    <SidebarItem to="/platform/markets" icon={<TrendingUp size={20} />} label="Markets" />
                    <SidebarItem to="/platform/wallet" icon={<Wallet size={20} />} label="Wallet" />
                    <SidebarItem to="/platform/notifications" icon={<Bell size={20} />} label="Notifications" badge="3" />
                </nav>
            </div>

            <div className="mt-auto p-8 border-t border-white/5">
                <nav className="space-y-2 mb-6">
                    <SidebarItem to="/platform/settings" icon={<Settings size={20} />} label="Settings" />
                </nav>

                <div className="p-4 rounded-xl bg-gradient-to-br from-primary/10 to-purple-500/10 border border-white/5 mb-4">
                    <h4 className="font-semibold text-sm mb-1">Pro Plan</h4>
                    <p className="text-xs text-muted-foreground mb-3">Get advanced analytics</p>
                    <button className="w-full py-2 bg-primary text-primary-foreground rounded-lg text-xs font-bold hover:bg-primary/90 transition-colors">
                        Upgrade Now
                    </button>
                </div>

                <button className="flex items-center gap-3 w-full px-4 py-3 text-muted-foreground hover:text-red-400 hover:bg-red-500/10 rounded-xl transition-all group">
                    <LogOut size={20} className="group-hover:-translate-x-1 transition-transform" />
                    <span className="font-medium">Log Out</span>
                </button>
            </div>
        </aside>
    );
};

const SidebarItem = ({ to, icon, label, end, badge }: { to: string, icon: React.ReactNode, label: string, end?: boolean, badge?: string }) => (
    <NavLink
        to={to}
        end={end}
        className={({ isActive }) => `
            flex items-center justify-between w-full px-4 py-3 rounded-xl transition-all duration-200 group
            ${isActive
                ? 'bg-primary text-primary-foreground shadow-lg shadow-primary/25'
                : 'text-muted-foreground hover:bg-secondary/50 hover:text-foreground'
            }
        `}
    >
        <div className="flex items-center gap-3">
            {icon}
            <span className="font-medium">{label}</span>
        </div>
        {badge && (
            <span className="px-2 py-0.5 rounded-full bg-red-500 text-white text-[10px] font-bold">
                {badge}
            </span>
        )}
    </NavLink>
);

export default Sidebar;
