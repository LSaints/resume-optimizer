import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import * as jobService from '../services/jobService'
import type { JobResponse } from '../types/job'
import Button from '../components/Button'
import Modal from '../components/Modal'
import styles from './JobList.module.css'

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('pt-BR', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function JobListPage() {
  const navigate = useNavigate()
  const [jobs, setJobs] = useState<JobResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [deleting, setDeleting] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<JobResponse | null>(null)

  useEffect(() => {
    jobService
      .list()
      .then(setJobs)
      .catch(() => setJobs([]))
      .finally(() => setLoading(false))
  }, [])

  async function handleDelete() {
    if (!deleteTarget) return
    setDeleting(deleteTarget.id)
    try {
      await jobService.remove(deleteTarget.id)
      setJobs((prev) => prev.filter((j) => j.id !== deleteTarget.id))
    } catch {
      /* silently fail */
    } finally {
      setDeleting(null)
      setDeleteTarget(null)
    }
  }

  if (loading) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Minhas Vagas</h1>
        </div>
        <div className={styles.skeleton}>
          {[1, 2, 3].map((i) => (
            <div key={i} className={styles.skeletonCard}>
              <div className={`${styles.skeletonLine} ${styles.skeletonLineNarrow}`} />
              <div className={`${styles.skeletonLine} ${styles.skeletonLineWide}`} />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <h1 className={styles.title}>Minhas Vagas</h1>
        <Link to="/jobs/new">
          <Button>Nova vaga</Button>
        </Link>
      </div>

      {jobs.length === 0 ? (
        <div className={styles.empty}>
          <div className={styles.emptyIcon}>💼</div>
          <p className={styles.emptyText}>
            Nenhuma vaga cadastrada ainda.
          </p>
          <Link to="/jobs/new">
            <Button>Cadastrar vaga</Button>
          </Link>
        </div>
      ) : (
        <div className={styles.list}>
          {jobs.map((job) => (
            <div key={job.id} className={styles.card}>
              <span className={styles.cardIcon}>💼</span>
              <div className={styles.cardInfo}>
                <div className={styles.cardTitle}>{job.title}</div>
                <div className={styles.cardPreview}>{job.rawDescription}</div>
                <div className={styles.cardDate}>{formatDate(job.createdAt)}</div>
              </div>
              <div className={styles.cardActions}>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => navigate(`/jobs/${job.id}/edit`)}
                >
                  Editar
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setDeleteTarget(job)}
                >
                  Excluir
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      <Modal
        open={!!deleteTarget}
        title="Excluir vaga"
        message={
          deleteTarget
            ? `Tem certeza que deseja excluir "${deleteTarget.title}"? Esta ação não pode ser desfeita.`
            : ''
        }
        confirmLabel="Excluir"
        loading={!!deleting}
        onConfirm={handleDelete}
        onCancel={() => setDeleteTarget(null)}
      />
    </div>
  )
}

export default JobListPage
