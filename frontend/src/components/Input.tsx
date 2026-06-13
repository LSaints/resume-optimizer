import type { InputHTMLAttributes } from 'react'
import styles from './Input.module.css'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

function Input({ label, error, className, id, ...props }: InputProps) {
  const inputId = id ?? label?.toLowerCase().replace(/\s+/g, '-')

  return (
    <div className={`${styles.field} ${className ?? ''}`}>
      {label && (
        <label className={styles.label} htmlFor={inputId}>
          {label}
        </label>
      )}
      <input
        id={inputId}
        className={`${styles.input} ${error ? styles.hasError : ''}`}
        {...props}
      />
      {error && <span className={styles.error}>{error}</span>}
    </div>
  )
}

export default Input
