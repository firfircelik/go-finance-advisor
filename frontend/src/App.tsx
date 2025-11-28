import { BrowserRouter as Router, Routes, Route, useLocation, Navigate } from 'react-router-dom';
import LandingPage from './pages/LandingPage';
import Login from './pages/Login';
import Signup from './pages/Signup';
import Platform from './pages/Platform';
import Navbar from './components/Navbar';
import Footer from './components/Footer';

// Layout wrapper to conditionally show Navbar/Footer
const Layout = ({ children }: { children: React.ReactNode }) => {
  const location = useLocation();
  const isPlatform = location.pathname.startsWith('/platform');
  const isAuth = ['/login', '/signup'].includes(location.pathname);

  // Don't show default Navbar/Footer on Platform or Auth pages (they have their own layouts)
  if (isPlatform || isAuth) {
    return <>{children}</>;
  }

  return (
    <div className="min-h-screen flex flex-col bg-background text-foreground font-sans selection:bg-primary/30">
      <Navbar />
      <main className="flex-grow">
        {children}
      </main>
      <Footer />
    </div>
  );
};

const ProtectedRoute = ({ children }: { children: React.ReactElement }) => {
  const token = localStorage.getItem('token');

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return children;
};

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<LandingPage />} />
          <Route path="/login" element={<Login />} />
          <Route path="/signup" element={<Signup />} />
          <Route
            path="/platform/*"
            element={
              <ProtectedRoute>
                <Platform />
              </ProtectedRoute>
            }
          />
        </Routes>
      </Layout>
    </Router>
  );
}

export default App;
