import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import * as resumeService from '../services/resumeService'
import type { ResumeResponse } from '../types/resume'
import Button from '../components/Button'
import Modal from '../components/Modal'
import styles from './ResumeList.module.css'

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  return d.toLocaleDateString('pt-BR', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

function ResumeListPage() {
  const navigate = useNavigate()
  const [resumes, setResumes] = useState<ResumeResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [deleting, setDeleting] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<ResumeResponse | null>(null)

  useEffect(() => {
    resumeService
      .list()
      .then(setResumes)
      .catch(() => setResumes([]))
      .finally(() => setLoading(false))
  }, [])

  async function handleDelete() {
    if (!deleteTarget) return
    setDeleting(deleteTarget.id)
    try {
      await resumeService.remove(deleteTarget.id)
      setResumes((prev) => prev.filter((r) => r.id !== deleteTarget.id))
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
          <h1 className={styles.title}>Meus Currículos</h1>
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
        <h1 className={styles.title}>Meus Currículos</h1>
        <Link to="/resumes/new">
          <Button>Enviar currículo</Button>
        </Link>
      </div>

      {resumes.length === 0 ? (
        <div className={styles.empty}>
          <div className={styles.emptyIcon}>📄</div>
          <p className={styles.emptyText}>
            Nenhum currículo enviado ainda.
          </p>
          <Link to="/resumes/new">
            <Button>Enviar currículo</Button>
          </Link>
        </div>
      ) : (
        <div className={styles.list}>
          {resumes.map((resume) => (
            <div key={resume.id} className={styles.card}>
              <span className={styles.cardIcon}>📄</span>
              <div className={styles.cardInfo}>
                <div className={styles.cardName}>{resume.originalName}</div>
                <div className={styles.cardDate}>{formatDate(resume.uploadAt)}</div>
              </div>
              <div className={styles.cardActions}>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => navigate(`/optimize?resume=${resume.id}`)}
                >
                  Otimizar
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setDeleteTarget(resume)}
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
        title="Excluir currículo"
        message={
          deleteTarget
            ? `Tem certeza que deseja excluir "${deleteTarget.originalName}"? Esta ação não pode ser desfeita.`
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

export default ResumeListPage
