import { useState, type FormEvent } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import Input from '../components/Input'
import Button from '../components/Button'
import styles from './Auth.module.css'

function RegisterPage() {
  const { register, isAuthenticated } = useAuth()
  const navigate = useNavigate()

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [errors, setErrors] = useState<{ name?: string; email?: string; password?: string }>({})
  const [apiError, setApiError] = useState('')
  const [loading, setLoading] = useState(false)

  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }

  function validate() {
    const newErrors: { name?: string; email?: string; password?: string } = {}

    if (!name.trim()) {
      newErrors.name = 'Informe seu nome.'
    }

    if (!email.trim()) {
      newErrors.email = 'Informe seu email.'
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Email inválido.'
    }

    if (!password) {
      newErrors.password = 'Informe uma senha.'
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
      await register(name, email, password)
      navigate('/')
    } catch (err) {
      const status = (err as { status?: number }).status
      if (status === 409) {
        setApiError('Este email já está cadastrado.')
      } else {
        setApiError(
          (err as { message?: string }).message || 'Erro ao criar conta. Tente novamente.',
        )
      }
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
        <p className={styles.subtitle}>Crie sua conta</p>

        <form className={styles.form} onSubmit={handleSubmit} noValidate>
          {apiError && <div className={styles.error}>{apiError}</div>}

          <Input
            label="Nome"
            type="text"
            placeholder="Seu nome"
            value={name}
            onChange={(e) => setName(e.target.value)}
            error={errors.name}
            disabled={loading}
          />

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
            placeholder="Mínimo de 6 caracteres"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={errors.password}
            disabled={loading}
          />

          <div className={styles.submit}>
            <Button type="submit" loading={loading} style={{ width: '100%' }}>
              Criar conta
            </Button>
          </div>
        </form>

        <p className={styles.footer}>
          Já tem conta?{' '}
          <Link to="/login">Entrar</Link>
        </p>
      </div>
    </div>
  )
}

export default RegisterPage
