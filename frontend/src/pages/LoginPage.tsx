import { useState, type FormEvent } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import Input from '../components/Input'
import Button from '../components/Button'
import styles from './Auth.module.css'

function LoginPage() {
  const { login, isAuthenticated } = useAuth()
  const navigate = useNavigate()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [errors, setErrors] = useState<{ email?: string; password?: string }>({})
  const [apiError, setApiError] = useState('')
  const [loading, setLoading] = useState(false)

  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }

  function validate() {
    const newErrors: { email?: string; password?: string } = {}

    if (!email.trim()) {
      newErrors.email = 'Informe seu email.'
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Email inválido.'
    }

    if (!password) {
      newErrors.password = 'Informe sua senha.'
    } else if (password.length < 6) {
      newErrors.password = 'A senha deve ter no mínimo 6 caracteres.'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setApiError('')

    if (!validate()) return

    setLoading(true)
    try {
      await login(email, password)
      navigate('/')
    } catch (err) {
      setApiError(
        (err as { message?: string }).message || 'Credenciais inválidas. Verifique seu email e senha.',
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <h1 className={styles.brand}>
          Resume<span className={styles.brandAccent}>Optimizer</span>
        </h1>
        <p className={styles.subtitle}>Entre na sua conta</p>

        <form className={styles.form} onSubmit={handleSubmit} noValidate>
          {apiError && <div className={styles.error}>{apiError}</div>}

          <Input
            label="Email"
            type="email"
            placeholder="seu@email.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            error={errors.email}
            disabled={loading}
          />

          <Input
            label="Senha"
            type="password"
            placeholder="Sua senha"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={errors.password}
            disabled={loading}
          />

          <div className={styles.submit}>
            <Button type="submit" loading={loading} style={{ width: '100%' }}>
              Entrar
            </Button>
          </div>
        </form>

        <p className={styles.footer}>
          Não tem conta?{' '}
          <Link to="/register">Criar conta</Link>
        </p>
      </div>
    </div>
  )
}

export default LoginPage
