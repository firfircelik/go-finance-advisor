import React from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ArrowRight, BarChart2, Shield, Zap, TrendingUp, Globe, Lock, CheckCircle2 } from 'lucide-react';
import Ticker from '../components/Ticker';

const LandingPage = () => {
    return (
        <div className="flex flex-col">
            <Ticker />

            {/* Hero Section */}
            <section className="relative pt-20 pb-32 md:pt-32 md:pb-48 overflow-hidden">
                {/* Background Elements */}
                <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full h-full max-w-7xl pointer-events-none">
                    <div className="absolute top-20 left-10 w-[500px] h-[500px] bg-primary/10 rounded-full blur-[120px] animate-pulse-slow" />
                    <div className="absolute top-40 right-10 w-[400px] h-[400px] bg-purple-500/10 rounded-full blur-[100px] animate-pulse-slow delay-1000" />
                </div>

                <div className="container mx-auto px-4 relative z-10 text-center">
                    <motion.div
                        initial={{ opacity: 0, y: 30 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.8, ease: "easeOut" }}
                    >
                        <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-secondary/50 border border-white/10 backdrop-blur-sm mb-8 hover:bg-secondary/70 transition-colors cursor-default">
                            <span className="relative flex h-2 w-2">
                                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                                <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                            </span>
                            <span className="text-sm font-medium text-muted-foreground">v2.0 is now live</span>
                        </div>

                        <h1 className="text-5xl md:text-7xl lg:text-8xl font-bold tracking-tight mb-8 font-heading">
                            Wealth Management <br />
                            <span className="text-gradient-primary animate-gradient-x">Reimagined</span>
                        </h1>

                        <p className="text-xl md:text-2xl text-muted-foreground mb-12 max-w-3xl mx-auto leading-relaxed font-light">
                            Experience the future of finance with AI-driven insights, real-time market data, and institutional-grade analytics.
                        </p>

                        <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
                            <Link to="/signup">
                                <button className="group px-8 py-4 bg-primary text-primary-foreground rounded-full font-semibold text-lg hover:bg-primary/90 transition-all flex items-center gap-2 shadow-lg shadow-primary/25 hover:-translate-y-1">
                                    Start Investing <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                                </button>
                            </Link>
                            <Link to="/login">
                                <button className="px-8 py-4 bg-secondary/50 border border-white/10 text-foreground rounded-full font-semibold text-lg hover:bg-secondary transition-all hover:-translate-y-1 backdrop-blur-sm">
                                    View Demo
                                </button>
                            </Link>
                        </div>
                    </motion.div>

                    {/* Dashboard Preview */}
                    <motion.div
                        initial={{ opacity: 0, y: 100, rotateX: 20 }}
                        animate={{ opacity: 1, y: 0, rotateX: 0 }}
                        transition={{ delay: 0.4, duration: 1, type: "spring" }}
                        className="mt-20 relative mx-auto max-w-6xl perspective-1000"
                    >
                        <div className="relative rounded-xl border border-white/10 bg-card/50 backdrop-blur-xl shadow-2xl overflow-hidden aspect-[16/9] group">
                            <div className="absolute inset-0 bg-gradient-to-tr from-primary/5 to-purple-500/5" />

                            {/* Abstract UI Representation */}
                            <div className="absolute inset-0 p-8 flex flex-col">
                                {/* Header */}
                                <div className="flex items-center justify-between mb-8 border-b border-white/5 pb-6">
                                    <div className="flex items-center gap-4">
                                        <div className="w-12 h-12 rounded-full bg-secondary animate-pulse" />
                                        <div className="space-y-2">
                                            <div className="w-32 h-4 bg-secondary rounded animate-pulse" />
                                            <div className="w-20 h-3 bg-secondary/50 rounded animate-pulse" />
                                        </div>
                                    </div>
                                    <div className="flex gap-4">
                                        <div className="w-24 h-10 bg-secondary rounded-lg animate-pulse" />
                                        <div className="w-10 h-10 bg-secondary rounded-lg animate-pulse" />
                                    </div>
                                </div>

                                {/* Main Content Grid */}
                                <div className="grid grid-cols-12 gap-6 flex-1">
                                    {/* Sidebar */}
                                    <div className="col-span-2 hidden md:flex flex-col gap-4">
                                        {[1, 2, 3, 4, 5].map((i) => (
                                            <div key={i} className="w-full h-10 bg-secondary/30 rounded-lg animate-pulse" style={{ opacity: 1 - i * 0.15 }} />
                                        ))}
                                    </div>

                                    {/* Chart Area */}
                                    <div className="col-span-12 md:col-span-7 bg-secondary/20 rounded-xl border border-white/5 p-6 relative overflow-hidden">
                                        <div className="absolute inset-0 bg-gradient-to-t from-primary/5 to-transparent" />
                                        <div className="flex items-end justify-between h-full gap-2 px-4 pb-4">
                                            {[...Array(12)].map((_, i) => (
                                                <div
                                                    key={i}
                                                    className="w-full bg-primary/20 rounded-t-sm hover:bg-primary/40 transition-colors duration-500"
                                                    style={{ height: `${Math.random() * 60 + 20}%` }}
                                                />
                                            ))}
                                        </div>
                                    </div>

                                    {/* Right Panel */}
                                    <div className="col-span-12 md:col-span-3 flex flex-col gap-6">
                                        <div className="flex-1 bg-secondary/20 rounded-xl border border-white/5 p-6">
                                            <div className="w-16 h-16 rounded-full border-4 border-primary/30 border-t-primary mx-auto mb-4 animate-spin" style={{ animationDuration: '3s' }} />
                                            <div className="w-full h-4 bg-secondary/50 rounded animate-pulse mb-2" />
                                            <div className="w-2/3 h-4 bg-secondary/30 rounded animate-pulse mx-auto" />
                                        </div>
                                        <div className="flex-1 bg-secondary/20 rounded-xl border border-white/5 p-6">
                                            <div className="space-y-3">
                                                {[1, 2, 3].map(i => (
                                                    <div key={i} className="flex items-center gap-3">
                                                        <div className="w-8 h-8 rounded-full bg-secondary/50" />
                                                        <div className="flex-1 h-2 bg-secondary/50 rounded" />
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            {/* Overlay Content */}
                            <div className="absolute inset-0 flex items-center justify-center bg-background/40 backdrop-blur-[2px] opacity-0 group-hover:opacity-100 transition-opacity duration-500">
                                <div className="bg-background/90 backdrop-blur-md px-8 py-4 rounded-full border border-white/10 shadow-2xl transform translate-y-4 group-hover:translate-y-0 transition-transform duration-500">
                                    <span className="font-semibold bg-clip-text text-transparent bg-gradient-to-r from-primary to-purple-400 text-lg">Interactive Dashboard Demo</span>
                                </div>
                            </div>
                        </div>
                    </motion.div>
                </div>
            </section>

            {/* Features Grid */}
            <section className="py-32 px-4 relative bg-secondary/5">
                <div className="container mx-auto max-w-6xl">
                    <div className="text-center mb-20">
                        <h2 className="text-3xl md:text-5xl font-bold mb-6 font-heading">Everything you need to succeed</h2>
                        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
                            Powerful tools designed for modern investors, all in one place.
                        </p>
                    </div>

                    <div className="grid md:grid-cols-3 gap-8">
                        <FeatureCard
                            icon={<BarChart2 className="w-8 h-8 text-blue-500" />}
                            title="Advanced Analytics"
                            description="Deep dive into your portfolio performance with institutional-grade charting and metrics."
                            delay={0}
                        />
                        <FeatureCard
                            icon={<Globe className="w-8 h-8 text-purple-500" />}
                            title="Global Markets"
                            description="Access real-time data from BIST, NYSE, NASDAQ, and Crypto markets instantly."
                            delay={0.1}
                        />
                        <FeatureCard
                            icon={<Zap className="w-8 h-8 text-yellow-500" />}
                            title="AI Insights"
                            description="Get personalized investment recommendations powered by state-of-the-art machine learning."
                            delay={0.2}
                        />
                        <FeatureCard
                            icon={<Shield className="w-8 h-8 text-green-500" />}
                            title="Bank-Grade Security"
                            description="Your assets are protected with military-grade encryption and security protocols."
                            delay={0.3}
                        />
                        <FeatureCard
                            icon={<TrendingUp className="w-8 h-8 text-red-500" />}
                            title="Real-Time Alerts"
                            description="Never miss a market movement with customizable price and news alerts."
                            delay={0.4}
                        />
                        <FeatureCard
                            icon={<Lock className="w-8 h-8 text-indigo-500" />}
                            title="Private & Secure"
                            description="We prioritize your privacy. Your data is encrypted and never shared with third parties."
                            delay={0.5}
                        />
                    </div>
                </div>
            </section>

            {/* Social Proof / Trust */}
            <section className="py-20 border-y border-white/5 bg-background">
                <div className="container mx-auto px-4 text-center">
                    <p className="text-sm font-medium text-muted-foreground mb-8 uppercase tracking-wider">Trusted by investors from</p>
                    <div className="flex flex-wrap justify-center gap-12 opacity-50 grayscale hover:grayscale-0 transition-all duration-500">
                        {/* Placeholders for logos */}
                        <div className="text-2xl font-bold flex items-center gap-2"><Globe className="w-6 h-6" /> GlobalCorp</div>
                        <div className="text-2xl font-bold flex items-center gap-2"><Zap className="w-6 h-6" /> FastTrade</div>
                        <div className="text-2xl font-bold flex items-center gap-2"><Shield className="w-6 h-6" /> SecureBank</div>
                        <div className="text-2xl font-bold flex items-center gap-2"><TrendingUp className="w-6 h-6" /> GrowthFund</div>
                    </div>
                </div>
            </section>

            {/* CTA Section */}
            <section className="py-32 px-4">
                <div className="container mx-auto max-w-5xl">
                    <div className="relative rounded-3xl overflow-hidden bg-primary px-6 py-20 text-center group">
                        <div className="absolute inset-0 bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20" />
                        <div className="absolute inset-0 bg-gradient-to-br from-primary via-blue-600 to-purple-700 transition-all duration-1000 group-hover:scale-110" />

                        <div className="relative z-10">
                            <h2 className="text-3xl md:text-5xl font-bold text-white mb-6 font-heading">Ready to take control?</h2>
                            <p className="text-xl text-blue-100 mb-10 max-w-2xl mx-auto">
                                Join thousands of investors who are already growing their wealth with GoFinance.
                            </p>
                            <div className="flex flex-col sm:flex-row gap-4 justify-center">
                                <Link to="/signup">
                                    <button className="px-10 py-4 bg-white text-primary rounded-full font-bold text-lg hover:bg-blue-50 transition-all shadow-xl hover:shadow-2xl hover:-translate-y-1 w-full sm:w-auto">
                                        Get Started for Free
                                    </button>
                                </Link>
                                <ul className="flex flex-col items-start gap-2 text-blue-100 text-sm mt-4 sm:mt-0 sm:ml-8 text-left">
                                    <li className="flex items-center gap-2"><CheckCircle2 className="w-4 h-4" /> No credit card required</li>
                                    <li className="flex items-center gap-2"><CheckCircle2 className="w-4 h-4" /> 14-day free trial</li>
                                    <li className="flex items-center gap-2"><CheckCircle2 className="w-4 h-4" /> Cancel anytime</li>
                                </ul>
                            </div>
                        </div>
                    </div>
                </div>
            </section>
        </div>
    );
};

const FeatureCard = ({ icon, title, description, delay }: { icon: React.ReactNode, title: string, description: string, delay: number }) => (
    <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ delay, duration: 0.5 }}
        whileHover={{ y: -5 }}
        className="p-8 rounded-2xl bg-card border border-white/5 shadow-lg hover:shadow-2xl hover:border-primary/20 transition-all group"
    >
        <div className="mb-6 p-4 rounded-xl bg-secondary/50 w-fit group-hover:bg-primary/10 transition-colors">
            {icon}
        </div>
        <h3 className="text-xl font-bold mb-3 text-foreground">{title}</h3>
        <p className="text-muted-foreground leading-relaxed">{description}</p>
    </motion.div>
);

export default LandingPage;

