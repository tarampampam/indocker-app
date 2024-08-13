import React, { type CSSProperties, type ReactNode } from 'react'
import styles from './icon.module.scss'

export default function Icon({ icon, style }: { icon: ReactNode, style?: CSSProperties }): React.JSX.Element {
  return <span className={styles.icon} style={{ ...style }}>
    {icon}
  </span>
}
