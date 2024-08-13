import { IconBrandDocker, IconInfoHexagonFilled, IconArrowRight } from '@tabler/icons-react' // https://tabler.io/icons
import React, { useEffect } from 'react'
import { Link, Outlet } from 'react-router-dom'
import { type SemVer } from 'semver'
import { type Client } from '~/api'
import { pathTo, RouteIDs } from '~/router'
import {Icon} from '~/shared/components'
import styles from './layout.module.scss'

const Header = (): React.JSX.Element => {
  return (
    <header className={styles.header}>
      <nav className={styles.headerNavigation}>
        <Link to={pathTo(RouteIDs.Containers)}>
          <Icon icon={<IconBrandDocker stroke={1} />} />
          Containers
        </Link>
        <Link to={pathTo(RouteIDs.About)}>
          <Icon icon={<IconInfoHexagonFilled />} />
          About
        </Link>
      </nav>
    </header>
  )
}

const Main = (): React.JSX.Element => {
  return (
    <main className={styles.main}>
      <Outlet />
    </main>
  )
}

const Footer = ({
  current = null,
  latest = null,
}: {
  current: SemVer | null
  latest: SemVer | null
}): React.JSX.Element => {
  const isUpdateAvailable = current && latest && current.compare(latest) === -1
  const latestReleaseLink = 'https://github.com/tarampampam/indocker-app/releases/latest'

  return (
    <footer className={styles.footer}>
      <span className={isUpdateAvailable ? 'text-light' : undefined}>
        Version: {current ? current.toString() : '...'}
      </span>
      {isUpdateAvailable && <a href={latestReleaseLink} target="_blank" rel="noreferrer">
        an update {latest.toString()} is available
      </a>}
      {/*{isUpdateAvailable && <p>A new version is available: <strong>{latest.toString()}</strong></p>}*/}
    </footer>
  )
}

export default function Layout({ apiClient }: { apiClient: Client }): React.JSX.Element {
  const [currentVersion, setCurrentVersion] = React.useState<SemVer | null>(null)
  const [latestVersion, setLatestVersion] = React.useState<SemVer | null>(null)

  useEffect(() => {
    apiClient
      .currentVersion()
      .then((version) => setCurrentVersion(version))
      .catch(console.error)

    apiClient
      .latestVersion()
      .then((version) => setLatestVersion(version))
      .catch(console.error)
  }, [apiClient])

  return (
    <>
      <Header />
      <Main />
      <Footer current={currentVersion} latest={latestVersion} />
    </>
  )
}
