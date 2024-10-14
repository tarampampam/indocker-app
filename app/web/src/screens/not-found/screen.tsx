import type React from 'react'
import deadDockerSvg from '~/assets/dead-docker.svg'
import styles from './screen.module.scss'

export default function Screen(): React.JSX.Element {
  return (
    <main className={styles.main}>
      <img src={deadDockerSvg} alt="404" width={500} />
      <h1>Unfortunately, requested page was not found</h1>
    </main>
  )
}
