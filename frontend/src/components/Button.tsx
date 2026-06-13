import type { ButtonHTMLAttributes, ReactNode } from 'react'
import LoadingSpinner from './LoadingSpinner'
import styles from './Button.module.css'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  loading?: boolean
  children: ReactNode
}

function Button({
  variant = 'primary',
  size = 'md',
  loading = false,
  disabled,
  children,
  className,
  ...props
}: ButtonProps) {
  return (
    <button
      className={`${styles.button} ${styles[variant]} ${styles[size]} ${className ?? ''}`}
      disabled={disabled || loading}
      {...props}
    >
      {loading && <LoadingSpinner size="sm" />}
      {children}
    </button>
  )
}

export default Button
