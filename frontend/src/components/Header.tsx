import { useState } from 'react'
import { Link, NavLink, useLocation } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import styles from './Header.module.css'

const navItems = [
  { to: '/', label: 'Dashboard' },
  { to: '/resumes', label: 'Currículos' },
  { to: '/jobs', label: 'Vagas' },
  { to: '/optimize', label: 'Otimizar' },
  { to: '/evaluate', label: 'Avaliar' },
]

function Header() {
  const { user, logout } = useAuth()
  const [drawerOpen, setDrawerOpen] = useState(false)
  const location = useLocation()

  const closeDrawer = () => setDrawerOpen(false)

  return (
    <>
      <header className={styles.header}>
        <Link to="/" className={styles.brand}>
          Resume<span className={styles.brandAccent}>Optimizer</span>
        </Link>

        <nav className={styles.desktopNav}>
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) =>
                `${styles.link} ${isActive ? styles.active : ''}`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className={styles.userSection}>
          <span className={styles.userName}>{user?.name}</span>
          <button className={styles.logoutBtn} onClick={logout}>
            Sair
          </button>
        </div>

        <button
          className={styles.hamburger}
          onClick={() => setDrawerOpen(true)}
          aria-label="Abrir menu"
        >
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M3 12h18M3 6h18M3 18h18" />
          </svg>
        </button>
      </header>

      {drawerOpen && (
        <div className={styles.overlay} onClick={closeDrawer} />
      )}

      <aside className={`${styles.drawer} ${drawerOpen ? styles.drawerOpen : ''}`}>
        <Link to="/" className={styles.brand} onClick={closeDrawer}>
          Resume<span className={styles.brandAccent}>Optimizer</span>
        </Link>

        <nav className={styles.drawerNav}>
          {navItems.map((item) => {
            const isActive = item.to === '/'
              ? location.pathname === '/'
              : location.pathname.startsWith(item.to)
            return (
              <Link
                key={item.to}
                to={item.to}
                className={`${styles.drawerLink} ${isActive ? styles.active : ''}`}
                onClick={closeDrawer}
              >
                {item.label}
              </Link>
            )
          })}
        </nav>

        <div className={styles.drawerUser}>
          <div className={styles.drawerUserName}>{user?.name}</div>
          <button className={styles.logoutBtn} onClick={() => { logout(); closeDrawer() }}>
            Sair
          </button>
        </div>
      </aside>
    </>
  )
}

export default Header
