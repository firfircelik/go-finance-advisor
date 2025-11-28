import { Link } from 'react-router-dom';
import { Twitter, Linkedin, Github, Instagram } from 'lucide-react';

const Footer = () => {
    return (
        <footer className="bg-card border-t border-white/5 pt-20 pb-10">
            <div className="container mx-auto px-4">
                <div className="grid grid-cols-1 md:grid-cols-4 gap-12 mb-16">
                    <div className="col-span-1 md:col-span-1">
                        <Link to="/" className="flex items-center gap-2 mb-6">
                            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-600 to-purple-600 flex items-center justify-center text-white font-bold text-lg">
                                GF
                            </div>
                            <span className="text-lg font-bold text-foreground">GoFinance</span>
                        </Link>
                        <p className="text-muted-foreground text-sm leading-relaxed mb-6">
                            Empowering investors with institutional-grade analytics and AI-driven insights for a smarter financial future.
                        </p>
                        <div className="flex items-center gap-4">
                            <SocialLink icon={<Twitter className="w-4 h-4" />} href="#" />
                            <SocialLink icon={<Linkedin className="w-4 h-4" />} href="#" />
                            <SocialLink icon={<Github className="w-4 h-4" />} href="#" />
                            <SocialLink icon={<Instagram className="w-4 h-4" />} href="#" />
                        </div>
                    </div>

                    <div>
                        <h4 className="font-semibold text-foreground mb-6">Product</h4>
                        <ul className="space-y-4">
                            <FooterLink to="#">Features</FooterLink>
                            <FooterLink to="#">Pricing</FooterLink>
                            <FooterLink to="#">API</FooterLink>
                            <FooterLink to="#">Integrations</FooterLink>
                        </ul>
                    </div>

                    <div>
                        <h4 className="font-semibold text-foreground mb-6">Resources</h4>
                        <ul className="space-y-4">
                            <FooterLink to="#">Documentation</FooterLink>
                            <FooterLink to="#">Academy</FooterLink>
                            <FooterLink to="#">Blog</FooterLink>
                            <FooterLink to="#">Community</FooterLink>
                        </ul>
                    </div>

                    <div>
                        <h4 className="font-semibold text-foreground mb-6">Company</h4>
                        <ul className="space-y-4">
                            <FooterLink to="#">About</FooterLink>
                            <FooterLink to="#">Careers</FooterLink>
                            <FooterLink to="#">Legal</FooterLink>
                            <FooterLink to="#">Contact</FooterLink>
                        </ul>
                    </div>
                </div>

                <div className="pt-8 border-t border-white/5 flex flex-col md:flex-row items-center justify-between gap-4">
                    <p className="text-sm text-muted-foreground">
                        &copy; {new Date().getFullYear()} GoFinance Advisor. All rights reserved.
                    </p>
                    <div className="flex items-center gap-8">
                        <a href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Privacy Policy</a>
                        <a href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Terms of Service</a>
                    </div>
                </div>
            </div>
        </footer>
    );
};

const SocialLink = ({ icon, href }: { icon: React.ReactNode, href: string }) => (
    <a
        href={href}
        className="w-8 h-8 rounded-full bg-white/5 flex items-center justify-center text-muted-foreground hover:bg-primary hover:text-white transition-all"
    >
        {icon}
    </a>
);

const FooterLink = ({ to, children }: { to: string, children: React.ReactNode }) => (
    <li>
        <Link to={to} className="text-sm text-muted-foreground hover:text-primary transition-colors">
            {children}
        </Link>
    </li>
);

export default Footer;
