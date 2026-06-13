import { useState, useEffect, type FormEvent, type ChangeEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import * as jobService from '../services/jobService'
import Input from '../components/Input'
import Button from '../components/Button'
import styles from './JobForm.module.css'

function JobFormPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const isEditing = !!id

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [errors, setErrors] = useState<{ title?: string; description?: string }>({})
  const [apiError, setApiError] = useState('')
  const [loading, setLoading] = useState(false)
  const [pageLoading, setPageLoading] = useState(isEditing)

  useEffect(() => {
    if (!isEditing) return
    jobService
      .getById(id!)
      .then((job) => {
        setTitle(job.title)
        setDescription(job.rawDescription)
      })
      .catch(() => navigate('/jobs'))
      .finally(() => setPageLoading(false))
  }, [id, isEditing, navigate])

  function validate() {
    const newErrors: { title?: string; description?: string } = {}

    if (!title.trim()) {
      newErrors.title = 'Informe o título da vaga.'
    } else if (title.trim().length < 3) {
      newErrors.title = 'O título deve ter no mínimo 3 caracteres.'
    }

    if (!description.trim()) {
      newErrors.description = 'Informe a descrição da vaga.'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setApiError('')

    if (!validate()) return

    const data = { title: title.trim(), rawDescription: description.trim() }

    setLoading(true)
    try {
      if (isEditing) {
        await jobService.update(id!, data)
      } else {
        await jobService.create(data)
      }
      navigate('/jobs')
    } catch (err) {
      setApiError(
        (err as { message?: string }).message || 'Erro ao salvar vaga. Tente novamente.',
      )
    } finally {
      setLoading(false)
    }
  }

  if (pageLoading) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Carregando...</h1>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <button className={styles.backButton} onClick={() => navigate('/jobs')}>
          ←
        </button>
        <h1 className={styles.title}>
          {isEditing ? 'Editar Vaga' : 'Nova Vaga'}
        </h1>
      </div>

      <form className={styles.form} onSubmit={handleSubmit} noValidate>
        {apiError && <div className={styles.error}>{apiError}</div>}

        <Input
          label="Título"
          type="text"
          placeholder="Ex: Engenheiro de Software Sênior"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          error={errors.title}
          disabled={loading}
        />

        <div>
          <label
            style={{
              display: 'block',
              fontFamily: 'var(--font-display)',
              fontSize: '0.875rem',
              fontWeight: 600,
              color: 'var(--color-text)',
              marginBottom: 'var(--space-xs)',
            }}
            htmlFor="job-description"
          >
            Descrição
          </label>
          <textarea
            id="job-description"
            className={`${styles.textarea} ${errors.description ? styles.hasError : ''}`}
            placeholder="Descreva a vaga, requisitos e responsabilidades..."
            value={description}
            onChange={(e: ChangeEvent<HTMLTextAreaElement>) => setDescription(e.target.value)}
            disabled={loading}
          />
          {errors.description && (
            <span
              style={{
                fontSize: '0.8125rem',
                color: 'var(--color-error)',
                marginTop: 'var(--space-xs)',
                display: 'block',
              }}
            >
              {errors.description}
            </span>
          )}
        </div>

        <div className={styles.actions}>
          <Button variant="ghost" onClick={() => navigate('/jobs')} disabled={loading}>
            Cancelar
          </Button>
          <Button type="submit" loading={loading}>
            {isEditing ? 'Salvar' : 'Criar vaga'}
          </Button>
        </div>
      </form>
    </div>
  )
}

export default JobFormPage
