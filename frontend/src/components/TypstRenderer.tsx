import { useState, useCallback, useEffect } from 'react'
import { get } from '../services/api'
import type { RenderResponse } from '../types/render'
import LoadingSpinner from './LoadingSpinner'
import styles from './TypstRenderer.module.css'

interface TypstRendererProps {
  content: string
  optimizationID?: string
}

type ViewMode = 'rendered' | 'source'

function TypstRenderer({ content, optimizationID }: TypstRendererProps) {
  const [viewMode, setViewMode] = useState<ViewMode>(optimizationID ? 'rendered' : 'source')
  const [svgContent, setSvgContent] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!optimizationID) {
      setViewMode('source')
      return
    }
    setLoading(true)
    setError(null)
    get<RenderResponse>(`/optimizations/${optimizationID}/render`)
      .then((res) => {
        setSvgContent(res.svgContent)
        setViewMode('rendered')
      })
      .catch((err) => {
        setError(err?.message || 'Erro ao renderizar documento')
        setViewMode('source')
      })
      .finally(() => setLoading(false))
  }, [optimizationID])

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(content)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      /* silently fail */
    }
  }, [content])

  const handleViewToggle = useCallback(() => {
    setViewMode((prev) => (prev === 'rendered' ? 'source' : 'rendered'))
  }, [])

  return (
    <div className={styles.container}>
      <div className={styles.toolbar}>
        <span className={styles.toolbarLabel}>Typst</span>
        <div className={styles.toolbarActions}>
          {optimizationID && (
            <button
              className={styles.viewToggle}
              onClick={handleViewToggle}
            >
              {viewMode === 'rendered' ? 'Ver código-fonte' : 'Ver renderizado'}
            </button>
          )}
          {viewMode === 'source' && (
            <button
              className={`${styles.copyBtn} ${copied ? styles.copied : ''}`}
              onClick={handleCopy}
            >
              {copied ? '✓ Copiado!' : 'Copiar código'}
            </button>
          )}
          {viewMode === 'rendered' && optimizationID && (
            <a
              className={styles.downloadBtn}
              href={`http://localhost:8080/v1/optimizations/${optimizationID}/render/pdf`}
              download
            >
              Baixar PDF
            </a>
          )}
        </div>
      </div>
      <div className={styles.content}>
        {loading && (
          <div className={styles.loading}>
            <LoadingSpinner />
            <span>Renderizando documento...</span>
          </div>
        )}
        {error && (
          <div className={styles.error}>
            <p>{error}</p>
          </div>
        )}
        {!loading && viewMode === 'rendered' && svgContent && (
          <div
            className={styles.svgContainer}
            dangerouslySetInnerHTML={{ __html: svgContent }}
          />
        )}
        {!loading && viewMode === 'source' && (
          <pre className={styles.codeBlock}>{content}</pre>
        )}
      </div>
    </div>
  )
}

export default TypstRenderer
