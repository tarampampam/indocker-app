import React, { type CSSProperties } from 'react'
import styles from './wave-with-bubbles.module.scss'

export default function Component({ style }: { style?: CSSProperties }): React.JSX.Element {
  return <div className={styles.component} style={{ ...style }} />
}
