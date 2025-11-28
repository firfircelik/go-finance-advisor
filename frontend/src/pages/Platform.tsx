import { Routes, Route, Navigate } from 'react-router-dom';
import Sidebar from '../components/Sidebar';
import Dashboard from './Dashboard';
import Portfolio from './Portfolio';
import Markets from './Markets';
import Settings from './Settings';

const Platform = () => {
    return (
        <div className="flex min-h-screen bg-background text-foreground font-sans selection:bg-primary/30">
            <Sidebar />

            <main className="flex-1 overflow-y-auto h-screen bg-gradient-to-br from-background to-secondary/20">
                <div className="max-w-7xl mx-auto">
                    <Routes>
                        <Route path="/" element={<Dashboard />} />
                        <Route path="/portfolio" element={<Portfolio />} />
                        <Route path="/markets" element={<Markets />} />
                        <Route path="/settings" element={<Settings />} />
                        <Route path="*" element={<Navigate to="/platform" replace />} />
                    </Routes>
                </div>
            </main>
        </div>
    );
};

export default Platform;
