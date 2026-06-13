import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import * as resumeService from '../services/resumeService'
import * as jobService from '../services/jobService'
import * as evaluationService from '../services/evaluationService'
import type { ResumeResponse } from '../types/resume'
import type { JobResponse } from '../types/job'
import Button from '../components/Button'
import Select from '../components/Select'
import styles from './Evaluate.module.css'

function EvaluatePage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const preselectedResumeId = searchParams.get('resume')

  const [resumes, setResumes] = useState<ResumeResponse[]>([])
  const [jobs, setJobs] = useState<JobResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedResume, setSelectedResume] = useState('')
  const [selectedJob, setSelectedJob] = useState('')
  const [evaluating, setEvaluating] = useState(false)
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

  async function handleEvaluate() {
    if (!selectedResume || !selectedJob) return
    setError('')
    setEvaluating(true)
    try {
      const result = await evaluationService.evaluate(selectedResume, selectedJob)
      navigate(`/evaluations/${selectedResume}/${result.id}`)
    } catch (err) {
      const msg = (err as { message?: string }).message
      if (msg === 'currículo não encontrado' || msg === 'vaga não encontrada') {
        setError('Currículo ou vaga não encontrados.')
      } else if (msg === 'serviço de IA não configurado') {
        setError('O serviço de IA não está configurado.')
      } else {
        setError('Erro ao avaliar currículo. Tente novamente.')
      }
    } finally {
      setEvaluating(false)
    }
  }

  if (loading) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Avaliar Currículo</h1>
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

  const canEvaluate = !!selectedResume && !!selectedJob

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <h1 className={styles.title}>Avaliar Currículo</h1>
      </div>

      <div className={styles.form}>
        {error && <div className={styles.error}>{error}</div>}

        <Select
          label="Currículo"
          placeholder="Selecione um currículo"
          options={resumeOptions}
          value={selectedResume}
          onChange={(e) => setSelectedResume(e.target.value)}
          disabled={evaluating}
        />

        <Select
          label="Vaga"
          placeholder="Selecione uma vaga"
          options={jobOptions}
          value={selectedJob}
          onChange={(e) => setSelectedJob(e.target.value)}
          disabled={evaluating}
        />

        <div>
          <Button
            onClick={handleEvaluate}
            disabled={!canEvaluate}
            loading={evaluating}
          >
            {evaluating ? 'Avaliando...' : 'Avaliar'}
          </Button>
        </div>
      </div>
    </div>
  )
}

export default EvaluatePage
