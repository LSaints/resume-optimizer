import type { SelectHTMLAttributes } from 'react'
import styles from './Select.module.css'

interface SelectOption {
  value: string
  label: string
}

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string
  error?: string
  options: SelectOption[]
  placeholder?: string
}

function Select({ label, error, options, placeholder, className, id, ...props }: SelectProps) {
  const selectId = id ?? label?.toLowerCase().replace(/\s+/g, '-')

  return (
    <div className={`${styles.field} ${className ?? ''}`}>
      {label && (
        <label className={styles.label} htmlFor={selectId}>
          {label}
        </label>
      )}
      <div className={styles.wrapper}>
        <select
          id={selectId}
          className={`${styles.select} ${error ? styles.hasError : ''}`}
          {...props}
        >
          {placeholder && (
            <option value="" disabled>
              {placeholder}
            </option>
          )}
          {options.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        <span className={styles.arrow} aria-hidden="true">
          ▾
        </span>
      </div>
      {error && <span className={styles.error}>{error}</span>}
    </div>
  )
}

export default Select
