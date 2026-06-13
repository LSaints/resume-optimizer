import styles from './LoadingSpinner.module.css'

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg'
  color?: string
  className?: string
}

function LoadingSpinner({ size = 'md', color, className }: LoadingSpinnerProps) {
  return (
    <span
      className={`${styles.spinner} ${styles[size]} ${className ?? ''}`}
      style={color ? { color } : undefined}
      aria-hidden="true"
    />
  )
}

export default LoadingSpinner
