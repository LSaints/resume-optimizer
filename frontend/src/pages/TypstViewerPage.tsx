import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import * as optimizationService from '../services/optimizationService'
import * as resumeService from '../services/resumeService'
import * as jobService from '../services/jobService'
import type { OptimizeResponse } from '../types/optimization'
import TypstRenderer from '../components/TypstRenderer'
import Button from '../components/Button'
import styles from './TypstViewer.module.css'

function TypstViewerPage() {
  const navigate = useNavigate()
  const { resumeId, optimizationId } = useParams<{
    resumeId: string
    optimizationId: string
  }>()

  const [optimization, setOptimization] = useState<OptimizeResponse | null>(null)
  const [resumeName, setResumeName] = useState('')
  const [jobTitle, setJobTitle] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!resumeId || !optimizationId) {
      setError('Dados da otimização não encontrados.')
      setLoading(false)
      return
    }

    async function load() {
      try {
        const [opt, resumes, jobs] = await Promise.all([
          optimizationService.getByID(resumeId!, optimizationId!),
          resumeService.list(),
          jobService.list(),
        ])

        setOptimization(opt)

        const resume = resumes.find((r) => r.id === opt.resumeId) ?? null
        if (resume) setResumeName(resume.originalName)

        const job = jobs.find((j) => j.id === opt.jobId) ?? null
        if (job) setJobTitle(job.title)
      } catch {
        setError('Otimização não encontrada.')
      } finally {
        setLoading(false)
      }
    }

    load()
  }, [resumeId, optimizationId])

  function formatDate(dateStr: string): string {
    const d = new Date(dateStr)
    return d.toLocaleDateString('pt-BR', {
      day: '2-digit',
      month: 'long',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  if (loading) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Resultado da Otimização</h1>
        </div>
        <div className={styles.skeletonMeta}>
          {[1, 2, 3].map((i) => (
            <div key={i} className={styles.skeletonMetaItem}>
              <div className={`${styles.skeletonLine} ${styles.skeletonLineShort}`} />
              <div className={`${styles.skeletonLine} ${styles.skeletonLineMedium}`} />
            </div>
          ))}
        </div>
        <div className={styles.skeletonViewer} />
      </div>
    )
  }

  if (error || !optimization) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <button className={styles.backButton} onClick={() => navigate('/optimize')}>
            ←
          </button>
          <h1 className={styles.title}>{error || 'Erro'}</h1>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <button className={styles.backButton} onClick={() => navigate('/optimize')}>
          ←
        </button>
        <h1 className={styles.title}>Resultado da Otimização</h1>
      </div>

      <div className={styles.meta}>
        <div className={styles.metaItem}>
          <span className={styles.metaLabel}>Currículo</span>
          <span className={styles.metaValue}>{resumeName || '—'}</span>
        </div>
        <div className={styles.metaItem}>
          <span className={styles.metaLabel}>Vaga</span>
          <span className={styles.metaValue}>{jobTitle || '—'}</span>
        </div>
        <div className={styles.metaItem}>
          <span className={styles.metaLabel}>Data</span>
          <span className={styles.metaValue}>{formatDate(optimization.createdAt)}</span>
        </div>
      </div>

      <div className={styles.actions}>
        <Button variant="secondary" onClick={() => navigate('/optimize')}>
          Nova otimização
        </Button>
        <Button variant="ghost" onClick={() => navigate(`/resumes/${resumeId}/optimizations`)}>
          Ver histórico
        </Button>
      </div>

      <TypstRenderer content={optimization.typstContent} optimizationID={optimizationId} />
    </div>
  )
}

export default TypstViewerPage
