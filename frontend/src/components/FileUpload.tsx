import { useState, useRef, type ChangeEvent, type DragEvent } from 'react'
import styles from './FileUpload.module.css'

const ACCEPTED_TYPES = ['.pdf', '.docx']
const MAX_SIZE = 10 * 1024 * 1024

interface FileUploadProps {
  onFileSelected: (file: File) => void
  onFileRemoved: () => void
  selectedFile: File | null
  disabled?: boolean
  error?: string
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function FileUpload({ onFileSelected, onFileRemoved, selectedFile, disabled, error }: FileUploadProps) {
  const [dragging, setDragging] = useState(false)
  const [validationError, setValidationError] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  function validateFile(file: File): string | null {
    const ext = '.' + file.name.split('.').pop()?.toLowerCase()
    if (!ACCEPTED_TYPES.includes(ext)) {
      return 'Formato não suportado. Aceitamos apenas PDF e DOCX.'
    }
    if (file.size > MAX_SIZE) {
      return 'Arquivo muito grande. O limite é de 10MB.'
    }
    return null
  }

  function handleFile(file: File) {
    setValidationError('')
    const err = validateFile(file)
    if (err) {
      setValidationError(err)
      return
    }
    onFileSelected(file)
  }

  function handleInputChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (file) handleFile(file)
    if (inputRef.current) inputRef.current.value = ''
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault()
    if (!disabled) setDragging(true)
  }

  function handleDragLeave() {
    setDragging(false)
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault()
    setDragging(false)
    const file = e.dataTransfer.files?.[0]
    if (file) handleFile(file)
  }

  function handleClick() {
    if (!disabled) inputRef.current?.click()
  }

  const displayError = error || validationError

  return (
    <div>
      <input
        ref={inputRef}
        type="file"
        accept=".pdf,.docx"
        onChange={handleInputChange}
        style={{ display: 'none' }}
      />

      {selectedFile ? (
        <div className={styles.fileInfo}>
          <span>📄</span>
          <span className={styles.fileInfoName}>{selectedFile.name}</span>
          <span className={styles.fileInfoSize}>{formatSize(selectedFile.size)}</span>
          <button
            className={styles.fileInfoRemove}
            onClick={onFileRemoved}
            disabled={disabled}
            aria-label="Remover arquivo"
          >
            ✕
          </button>
        </div>
      ) : (
        <div
          className={`${styles.dropzone} ${dragging ? styles.dragging : ''} ${displayError ? styles.hasError : ''} ${disabled ? styles.disabled : ''}`}
          onClick={handleClick}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
        >
          <div className={styles.icon}>📄</div>
          <p className={styles.hint}>
            <strong>Clique aqui</strong> ou arraste um arquivo
          </p>
          <p className={styles.formats}>PDF ou DOCX · até 10MB</p>
        </div>
      )}

      {displayError && <p className={styles.error}>{displayError}</p>}
    </div>
  )
}

export default FileUpload
