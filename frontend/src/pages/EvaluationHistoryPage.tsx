import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import * as evaluationService from '../services/evaluationService'
import * as resumeService from '../services/resumeService'
import * as jobService from '../services/jobService'
import type { EvaluationSummaryResponse } from '../types/evaluation'
import Button from '../components/Button'
import styles from './EvaluationHistory.module.css'

interface HistoryItem extends EvaluationSummaryResponse {
  resumeName: string
  jobTitle: string
}

function EvaluationHistoryPage() {
  const navigate = useNavigate()
  const { id: resumeId } = useParams<{ id: string }>()

  const [items, setItems] = useState<HistoryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!resumeId) return

    async function load() {
      try {
        const [evaluations, resumes, jobs] = await Promise.all([
          evaluationService.listByResume(resumeId!),
          resumeService.list(),
          jobService.list(),
        ])

        const resume = resumes.find((r) => r.id === resumeId) ?? null

        const itemsWithMeta: HistoryItem[] = evaluations.map((evalItem) => {
          const job = jobs.find((j) => j.id === evalItem.jobId) ?? null
          return {
            ...evalItem,
            resumeName: resume?.originalName ?? '—',
            jobTitle: job?.title ?? '—',
          }
        })

        setItems(itemsWithMeta)
      } catch {
        setError('Erro ao carregar histórico de avaliações.')
      } finally {
        setLoading(false)
      }
    }

    load()
  }, [resumeId])

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

  function scoreClass(score: number): string {
    if (score >= 7) return styles.scoreHigh
    if (score >= 4) return styles.scoreMedium
    return styles.scoreLow
  }

  if (loading) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Histórico de Avaliações</h1>
        </div>
        <div className={styles.list}>
          {[1, 2, 3].map((i) => (
            <div key={i} className={styles.skeletonCard}>
              <div className={styles.skeletonBadge} />
              <div className={styles.skeletonInfo}>
                <div className={`${styles.skeletonLine} ${styles.skeletonLineLong}`} />
                <div className={`${styles.skeletonLine} ${styles.skeletonLineShort}`} />
              </div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Histórico de Avaliações</h1>
        </div>
        <p className={styles.error}>{error}</p>
        <Button variant="primary" onClick={() => navigate('/evaluate')}>
          Nova avaliação
        </Button>
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Histórico de Avaliações</h1>
        </div>
        <div className={styles.empty}>
          <p>Nenhuma avaliação encontrada.</p>
          <Button variant="primary" onClick={() => navigate('/evaluate')}>
            Nova avaliação
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <button className={styles.backButton} onClick={() => navigate('/evaluate')}>
          ←
        </button>
        <h1 className={styles.title}>Histórico de Avaliações</h1>
      </div>

      <div className={styles.list}>
        {items.map((item) => (
          <div
            key={item.id}
            className={styles.card}
            onClick={() => navigate(`/evaluations/${resumeId}/${item.id}`)}
          >
            <div className={`${styles.scoreBadge} ${scoreClass(item.score)}`}>
              {item.score.toFixed(1)}
            </div>
            <div className={styles.info}>
              <span className={styles.summary}>{item.summary}</span>
              <span className={styles.jobTitle}>{item.jobTitle}</span>
              <span className={styles.date}>{formatDate(item.createdAt)}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default EvaluationHistoryPage
