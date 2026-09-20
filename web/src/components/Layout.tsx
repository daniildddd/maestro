import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { AudioLines, Gauge, LineChart, ScrollText, UserRound, UsersRound } from "lucide-react";
import { useAuth } from "../auth/AuthContext";

export function Layout() {
  const { user, signOut } = useAuth();
  const navigate = useNavigate();

  const isAdmin = user?.role === "admin";

  const handleSignOut = async () => {
    await signOut();
    navigate("/login");
  };

  const navClass = ({ isActive }: { isActive: boolean }) => "nav-link" + (isActive ? " active" : "");

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <img className="brand-mark" src="/maestro.png" alt="Maestro" />
          <span className="brand-name">Maestro</span>
        </div>
        <nav className="nav">
          <NavLink to="/dashboard" className={navClass}>
            <Gauge size={16} strokeWidth={2} aria-hidden="true" />
            Dashboard
          </NavLink>
          <NavLink to="/connectors" className={navClass}>
            <AudioLines size={16} strokeWidth={2} aria-hidden="true" />
            Connectors
          </NavLink>
          <NavLink to="/metrics" className={navClass}>
            <LineChart size={16} strokeWidth={2} aria-hidden="true" />
            Metrics
          </NavLink>
          {isAdmin && (
            <NavLink to="/users" className={navClass}>
              <UsersRound size={16} strokeWidth={2} aria-hidden="true" />
              Users
            </NavLink>
          )}
          {isAdmin && (
            <NavLink to="/audit" className={navClass}>
              <ScrollText size={16} strokeWidth={2} aria-hidden="true" />
              Audit log
            </NavLink>
          )}
          <NavLink to="/profile" className={navClass}>
            <UserRound size={16} strokeWidth={2} aria-hidden="true" />
            Profile
          </NavLink>
        </nav>
        <div className="sidebar-footer">
          <div className="who">
            <div className="who-name">{user?.username}</div>
            <div className="who-role">{user?.role}</div>
          </div>
          <button className="btn btn-ghost" onClick={handleSignOut} type="button">
            Sign out
          </button>
        </div>
      </aside>
      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}