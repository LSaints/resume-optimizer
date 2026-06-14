import { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import * as evaluationService from '../services/evaluationService'
import * as resumeService from '../services/resumeService'
import * as jobService from '../services/jobService'
import type { EvaluationResponse } from '../types/evaluation'
import Button from '../components/Button'
import styles from './EvaluationResult.module.css'

function EvaluationResultPage() {
  const navigate = useNavigate()
  const { resumeId, evaluationId } = useParams<{
    resumeId: string
    evaluationId: string
  }>()

  const [evaluation, setEvaluation] = useState<EvaluationResponse | null>(null)
  const [resumeName, setResumeName] = useState('')
  const [jobTitle, setJobTitle] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!resumeId || !evaluationId) {
      setError('Dados da avaliação não encontrados.')
      setLoading(false)
      return
    }

    async function load() {
      try {
        const [evalResult, resumes, jobs] = await Promise.all([
          evaluationService.getByID(resumeId!, evaluationId!),
          resumeService.list(),
          jobService.list(),
        ])

        setEvaluation(evalResult)

        const resume = resumes.find((r) => r.id === evalResult.resumeId) ?? null
        if (resume) setResumeName(resume.originalName)

        const job = jobs.find((j) => j.id === evalResult.jobId) ?? null
        if (job) setJobTitle(job.title)
      } catch {
        setError('Avaliação não encontrada.')
      } finally {
        setLoading(false)
      }
    }

    load()
  }, [resumeId, evaluationId])

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

  function barClass(pct: number): string {
    if (pct >= 0.7) return styles.barHigh
    if (pct >= 0.4) return styles.barMedium
    return styles.barLow
  }

  function hasBreakdown(): boolean {
    return (
      evaluation!.breakdownKeywordMatch > 0 ||
      evaluation!.breakdownTechnical > 0 ||
      evaluation!.breakdownExperience > 0 ||
      evaluation!.breakdownImpact > 0 ||
      evaluation!.breakdownReadability > 0
    )
  }

  const breakdownItems = [
    { key: 'breakdownKeywordMatch' as const, label: 'Correspondência de Palavras-chave', weight: '30%', max: 3.0 },
    { key: 'breakdownTechnical' as const, label: 'Compatibilidade Técnica', weight: '25%', max: 2.5 },
    { key: 'breakdownExperience' as const, label: 'Experiência Profissional', weight: '20%', max: 2.0 },
    { key: 'breakdownImpact' as const, label: 'Impacto e Resultados', weight: '15%', max: 1.5 },
    { key: 'breakdownReadability' as const, label: 'Legibilidade ATS', weight: '10%', max: 1.0 },
  ]

  if (loading) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <h1 className={styles.title}>Resultado da Avaliação</h1>
        </div>
        <div className={styles.skeletonScore} />
        <div className={styles.skeletonMeta}>
          {[1, 2, 3].map((i) => (
            <div key={i} className={styles.skeletonMetaItem}>
              <div className={`${styles.skeletonLine} ${styles.skeletonLineShort}`} />
              <div className={`${styles.skeletonLine} ${styles.skeletonLineMedium}`} />
            </div>
          ))}
        </div>
        <div className={styles.skeletonBlock} />
      </div>
    )
  }

  if (error || !evaluation) {
    return (
      <div className={styles.page}>
        <div className={styles.header}>
          <button className={styles.backButton} onClick={() => navigate('/evaluate')}>
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
        <button className={styles.backButton} onClick={() => navigate('/evaluate')}>
          ←
        </button>
        <h1 className={styles.title}>Resultado da Avaliação</h1>
      </div>

      <div className={styles.scoreSection}>
        <div className={`${styles.score} ${scoreClass(evaluation.score)}`}>
          {evaluation.score.toFixed(1)}
        </div>
        <p className={styles.summary}>{evaluation.summary}</p>
      </div>

      {hasBreakdown() && (
        <div className={styles.breakdown}>
          <h2 className={styles.breakdownTitle}>Pontuação por Critério</h2>
          <div className={styles.breakdownList}>
            {breakdownItems.map((item) => {
              const value = evaluation[item.key]
              const pct = item.max > 0 ? value / item.max : 0
              return (
                <div key={item.key} className={styles.breakdownItem}>
                  <div className={styles.breakdownHeader}>
                    <span className={styles.breakdownLabel}>{item.label}</span>
                    <span className={styles.breakdownWeight}>{item.weight}</span>
                  </div>
                  <div className={styles.breakdownBarTrack}>
                    <div
                      className={`${styles.breakdownBarFill} ${barClass(pct)}`}
                      style={{ width: `${Math.min(pct * 100, 100)}%` }}
                    />
                  </div>
                  <span className={styles.breakdownScore}>
                    {value.toFixed(1)} / {item.max.toFixed(1)}
                  </span>
                </div>
              )
            })}
          </div>
        </div>
      )}

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
          <span className={styles.metaValue}>{formatDate(evaluation.createdAt)}</span>
        </div>
      </div>

      <div className={styles.actions}>
        <Button variant="secondary" onClick={() => navigate('/evaluate')}>
          Nova avaliação
        </Button>
        <Button variant="ghost" onClick={() => navigate(`/resumes/${resumeId}/evaluations`)}>
          Ver histórico
        </Button>
      </div>

      {(evaluation.matchedKeywords.length > 0 || evaluation.missingKeywords.length > 0) && (
        <div className={styles.keywords}>
          {evaluation.matchedKeywords.length > 0 && (
            <div className={styles.keywordSection}>
              <h2 className={styles.keywordSectionTitle}>Palavras-chave Encontradas</h2>
              <div className={styles.keywordTags}>
                {evaluation.matchedKeywords.map((kw) => (
                  <span key={kw} className={`${styles.keywordTag} ${styles.keywordTagMatch}`}>
                    {kw}
                  </span>
                ))}
              </div>
            </div>
          )}
          {evaluation.missingKeywords.length > 0 && (
            <div className={styles.keywordSection}>
              <h2 className={styles.keywordSectionTitle}>Palavras-chave Ausentes</h2>
              <div className={styles.keywordTags}>
                {evaluation.missingKeywords.map((kw) => (
                  <span key={kw} className={`${styles.keywordTag} ${styles.keywordTagMissing}`}>
                    {kw}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {evaluation.recommendations.length > 0 && (
        <div className={styles.recommendations}>
          <h2 className={styles.recommendationsTitle}>Recomendações</h2>
          <ol className={styles.recommendationsList}>
            {evaluation.recommendations.map((rec, i) => (
              <li key={i} className={styles.recommendationItem}>
                {rec}
              </li>
            ))}
          </ol>
        </div>
      )}

      {evaluation.details && (
        <div className={styles.details}>
          <h2 className={styles.detailsTitle}>Detalhamento</h2>
          <p className={styles.detailsText}>{evaluation.details}</p>
        </div>
      )}
    </div>
  )
}

export default EvaluationResultPage
