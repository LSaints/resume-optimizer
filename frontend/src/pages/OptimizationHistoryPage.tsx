import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import * as optimizationService from '../services/optimizationService'
import * as resumeService from '../services/resumeService'
import * as jobService from '../services/jobService'
import * as renderService from '../services/renderService'
import type { OptimizeSummaryResponse } from '../types/optimization'
import Button from '../components/Button'
import Modal from '../components/Modal'
import styles from './OptimizationHistory.module.css'

interface HistoryItem extends OptimizeSummaryResponse {
  resumeName: string
  jobTitle: string
  svgThumbnail: string | null
}

function OptimizationHistoryPage() {
  const navigate = useNavigate()
  const { id: resumeId } = useParams<{ id: string }>()

  const [items, setItems] = useState<HistoryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    if (!resumeId) return

    async function load() {
      try {
        const [optimizations, resumes, jobs] = await Promise.all([
          optimizationService.listByResume(resumeId!),
          resumeService.list(),
          jobService.list(),
        ])

        const resume = resumes.find((r) => r.id === resumeId) ?? null

        const itemsWithMeta: HistoryItem[] = await Promise.all(
          optimizations.map(async (opt) => {
            const job = jobs.find((j) => j.id === opt.jobId) ?? null
            let svgThumbnail: string | null = null
            try {
              svgThumbnail = await renderService.getRenderSVG(opt.id)
            } catch {
              /* thumbnail unavailable */
            }
            return {
              ...opt,
              resumeName: resume?.originalName ?? '—',
              jobTitle: job?.title ?? '—',
              svgThumbnail,
            }
          }),
        )

        setItems(itemsWithMeta)
      } catch {
        setError('Erro ao carregar histórico de otimizações.')
      } finally {
        setLoading(false)
      }
    }

    load()
  }, [resumeId])

  async function handleDelete(optimizationId: string) {
    if (!resumeId) return
    setDeleting(true)
    try {
      await optimizationService.remove(resumeId, optimizationId)
      setItems((prev) => prev.filter((i) => i.id !== optimizationId))
    } catch {
      /* silently fail */
    } finally {
      setDeleting(false)
      setDeleteTarget(null)
    }
  }

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
          <h1 className={styles.title}>Histórico de Otimizações</h1>
        </div>
        <div className={styles.list}>
          {[1, 2, 3].map((i) => (
            <div key={i} className={styles.skeletonCard}>
              <div className={styles.skeletonThumb} />
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
          <h1 className={styles.title}>Histórico de Otimizações</h1>
        </div>
        <p className={styles.error}>{error}</p>
        <Button variant="primary" onClick={() => navigate('/optimize')}>
          Nova otimização
        </Button>
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Histórico de Otimizações</h1>
        </div>
        <div className={styles.empty}>
          <p>Nenhuma otimização encontrada.</p>
          <Button variant="primary" onClick={() => navigate('/optimize')}>
            Nova otimização
          </Button>
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
        <h1 className={styles.title}>Histórico de Otimizações</h1>
      </div>

      <div className={styles.list}>
        {items.map((item) => (
          <div
            key={item.id}
            className={styles.card}
            onClick={() => navigate(`/optimizations/${resumeId}/${item.id}`)}
          >
            <div className={styles.thumb}>
              {item.svgThumbnail ? (
                <div
                  className={styles.thumbSvg}
                  dangerouslySetInnerHTML={{ __html: item.svgThumbnail }}
                />
              ) : (
                <div className={styles.thumbFallback}>Typst</div>
              )}
            </div>
            <div className={styles.info}>
              <span className={styles.resumeName}>{item.resumeName}</span>
              <span className={styles.jobTitle}>{item.jobTitle}</span>
              <span className={styles.date}>{formatDate(item.createdAt)}</span>
            </div>
            <button
              className={styles.deleteBtn}
              onClick={(e) => {
                e.stopPropagation()
                setDeleteTarget(item.id)
              }}
              title="Excluir otimização"
            >
              Excluir
            </button>
          </div>
        ))}
      </div>

      <Modal
        open={deleteTarget !== null}
        title="Excluir otimização"
        message="Tem certeza que deseja excluir esta otimização? Esta ação não pode ser desfeita."
        confirmLabel="Excluir"
        cancelLabel="Cancelar"
        loading={deleting}
        onConfirm={() => deleteTarget && handleDelete(deleteTarget)}
        onCancel={() => setDeleteTarget(null)}
      />
    </div>
  )
}

export default OptimizationHistoryPage
