import { Routes, Route } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'

import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardPage from './pages/DashboardPage'
import ResumeListPage from './pages/ResumeListPage'
import ResumeUploadPage from './pages/ResumeUploadPage'
import JobListPage from './pages/JobListPage'
import JobFormPage from './pages/JobFormPage'
import OptimizePage from './pages/OptimizePage'
import TypstViewerPage from './pages/TypstViewerPage'
import OptimizationHistoryPage from './pages/OptimizationHistoryPage'

import Layout from './components/Layout'
import ProtectedRoute from './components/ProtectedRoute'

function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />

        <Route element={<ProtectedRoute />}>
          <Route element={<Layout />}>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/resumes" element={<ResumeListPage />} />
            <Route path="/resumes/new" element={<ResumeUploadPage />} />
            <Route path="/jobs" element={<JobListPage />} />
            <Route path="/jobs/new" element={<JobFormPage />} />
            <Route path="/jobs/:id/edit" element={<JobFormPage />} />
            <Route path="/optimize" element={<OptimizePage />} />
            <Route path="/optimizations/:resumeId/:optimizationId" element={<TypstViewerPage />} />
            <Route path="/resumes/:id/optimizations" element={<OptimizationHistoryPage />} />
          </Route>
        </Route>
      </Routes>
    </AuthProvider>
  )
}

export default App
