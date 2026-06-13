import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import * as resumeService from '../services/resumeService'
import * as jobService from '../services/jobService'
import * as optimizationService from '../services/optimizationService'
import type { ResumeResponse } from '../types/resume'
import type { JobResponse } from '../types/job'
import Button from '../components/Button'
import Select from '../components/Select'
import styles from './Optimize.module.css'

function OptimizePage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const preselectedResumeId = searchParams.get('resume')

  const [resumes, setResumes] = useState<ResumeResponse[]>([])
  const [jobs, setJobs] = useState<JobResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedResume, setSelectedResume] = useState('')
  const [selectedJob, setSelectedJob] = useState('')
  const [optimizing, setOptimizing] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [resumeList, jobList] = await Promise.all([
          resumeService.list(),
          jobService.list(),
        ])
        setResumes(resumeList)
        setJobs(jobList)

        if (preselectedResumeId && resumeList.some((r) => r.id === preselectedResumeId)) {
          setSelectedResume(preselectedResumeId)
        }
      } catch {
        setError('Erro ao carregar dados.')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [preselectedResumeId])

  async function handleOptimize() {
    if (!selectedResume || !selectedJob) return
    setError('')
    setOptimizing(true)
    try {
      const result = await optimizationService.optimize(selectedResume, selectedJob)
      navigate(`/optimizations/${selectedResume}/${result.id}`)
    } catch (err) {
      const msg = (err as { message?: string }).message
      if (msg === 'currículo não encontrado' || msg === 'vaga não encontrada') {
        setError('Currículo ou vaga não encontrados.')
      } else if (msg === 'serviço de IA não configurado') {
        setError('O serviço de IA não está configurado.')
      } else {
        setError('Erro ao otimizar currículo. Tente novamente.')
      }
    } finally {
      setOptimizing(false)
    }
  }

  if (loading) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Otimizar Currículo</h1>
        </div>
        <div className={styles.skeleton}>
          <div className={styles.skeletonLine} />
          <div className={styles.skeletonLine} />
          <div className={styles.skeletonButton} />
        </div>
      </div>
    )
  }

  const resumeOptions = resumes.map((r) => ({
    value: r.id,
    label: r.originalName,
  }))

  const jobOptions = jobs.map((j) => ({
    value: j.id,
    label: j.title,
  }))

  const canOptimize = !!selectedResume && !!selectedJob

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <h1 className={styles.title}>Otimizar Currículo</h1>
      </div>

      <div className={styles.form}>
        {error && <div className={styles.error}>{error}</div>}

        <Select
          label="Currículo"
          placeholder="Selecione um currículo"
          options={resumeOptions}
          value={selectedResume}
          onChange={(e) => setSelectedResume(e.target.value)}
          disabled={optimizing}
        />

        <Select
          label="Vaga"
          placeholder="Selecione uma vaga"
          options={jobOptions}
          value={selectedJob}
          onChange={(e) => setSelectedJob(e.target.value)}
          disabled={optimizing}
        />

        <div>
          <Button
            onClick={handleOptimize}
            disabled={!canOptimize}
            loading={optimizing}
          >
            {optimizing ? 'Otimizando...' : 'Otimizar'}
          </Button>
        </div>
      </div>
    </div>
  )
}

export default OptimizePage
