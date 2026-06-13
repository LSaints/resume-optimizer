import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import * as resumeService from '../services/resumeService'
import * as jobService from '../services/jobService'
import * as optimizationService from '../services/optimizationService'
import styles from './Dashboard.module.css'

interface DashboardData {
  resumeCount: number
  jobCount: number
  optimizationCount: number
}

function DashboardPage() {
  const { user } = useAuth()
  const [data, setData] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function fetchData() {
      try {
        const [resumes, jobs] = await Promise.all([
          resumeService.list(),
          jobService.list(),
        ])

        const optimizationCounts = await Promise.all(
          resumes.map((r) => optimizationService.listByResume(r.id)),
        )
        const optimizationCount = optimizationCounts.reduce(
          (sum, list) => sum + list.length, 0,
        )

        setData({
          resumeCount: resumes.length,
          jobCount: jobs.length,
          optimizationCount,
        })
      } catch {
        setData({ resumeCount: 0, jobCount: 0, optimizationCount: 0 })
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [])

  if (loading) {
    return (
      <div className={styles.page}>
        <h1 className={styles.greeting}>
          Olá, <span className={styles.greetingName}>—</span>
        </h1>
        <div className={styles.skeleton}>
          {[1, 2, 3].map((i) => (
            <div key={i} className={styles.skeletonCard}>
              <div className={`${styles.skeletonLine} ${styles.skeletonLineShort}`} />
              <div className={styles.skeletonCount} />
              <div className={`${styles.skeletonLine} ${styles.skeletonLineMedium}`} />
            </div>
          ))}
        </div>
      </div>
    )
  }

  const cards = [
    {
      icon: '📄',
      count: data?.resumeCount ?? 0,
      label: 'currículo enviado',
      plural: 'currículos enviados',
      link: '/resumes',
    },
    {
      icon: '💼',
      count: data?.jobCount ?? 0,
      label: 'vaga cadastrada',
      plural: 'vagas cadastradas',
      link: '/jobs',
    },
    {
      icon: '⚡',
      count: data?.optimizationCount ?? 0,
      label: 'otimização realizada',
      plural: 'otimizações realizadas',
      link: '/optimize',
    },
  ]

  return (
    <div className={styles.page}>
      <h1 className={styles.greeting}>
        Olá, <span className={styles.greetingName}>{user?.name}</span>
      </h1>

      <div className={styles.grid}>
        {cards.map((card) => (
          <Link key={card.label} to={card.link} className={styles.card}>
            <div className={styles.cardIcon}>{card.icon}</div>
            <div className={styles.cardCount}>{card.count}</div>
            <div className={styles.cardLabel}>
              {card.count === 1 ? card.label : card.plural}
            </div>
            <div className={styles.cardArrow}>→</div>
          </Link>
        ))}
      </div>
    </div>
  )
}

export default DashboardPage
