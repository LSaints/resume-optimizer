import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import * as resumeService from '../services/resumeService'
import FileUpload from '../components/FileUpload'
import Button from '../components/Button'
import LoadingSpinner from '../components/LoadingSpinner'
import styles from './ResumeUpload.module.css'

function ResumeUploadPage() {
  const navigate = useNavigate()
  const [file, setFile] = useState<File | null>(null)
  const [uploading, setUploading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  async function handleSubmit() {
    if (!file) return
    setError('')
    setUploading(true)
    try {
      await resumeService.upload(file)
      setSuccess(true)
      setTimeout(() => navigate('/resumes'), 1500)
    } catch (err) {
      const status = (err as { status?: number }).status
      if (status === 413) {
        setError('Arquivo muito grande. O limite é de 10MB.')
      } else {
        setError('Erro ao enviar currículo. Tente novamente.')
      }
    } finally {
      setUploading(false)
    }
  }

  if (success) {
    return (
      <div className={styles.page}>
        <div className={styles.success}>
          <div className={styles.successIcon}>✅</div>
          <p className={styles.successText}>Currículo enviado com sucesso!</p>
          <LoadingSpinner size="sm" color="var(--color-success)" />
        </div>
      </div>
    )
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <button className={styles.backButton} onClick={() => navigate('/resumes')}>
          ←
        </button>
        <h1 className={styles.title}>Enviar Currículo</h1>
      </div>

      <div className={styles.form}>
        <FileUpload
          selectedFile={file}
          onFileSelected={setFile}
          onFileRemoved={() => setFile(null)}
          error={error}
          disabled={uploading}
        />

        <div className={styles.actions}>
          <Button variant="ghost" onClick={() => navigate('/resumes')} disabled={uploading}>
            Cancelar
          </Button>
          <Button onClick={handleSubmit} loading={uploading} disabled={!file}>
            Enviar
          </Button>
        </div>
      </div>
    </div>
  )
}

export default ResumeUploadPage
