import { IconBrandDocker, IconInfoHexagonFilled, IconDownload } from '@tabler/icons-react' // https://tabler.io/icons
import React, { type ReactNode, useEffect } from 'react'
import { Link, Outlet } from 'react-router-dom'
import { type SemVer } from 'semver'
import { type Client } from '~/api'
import { pathTo, RouteIDs, useCurrentRouteID } from '~/routing'
import { Icon } from '~/shared/components'
import styles from './layout.module.scss'

const Header = ({
  current = null,
  latest = null,
}: {
  current: SemVer | null
  latest: SemVer | null
}): React.JSX.Element => {
  const currentRouteID = useCurrentRouteID()
  const isUpdateAvailable = current && latest && current.compare(latest) === -1

  const links: ReadonlyMap<RouteIDs, Readonly<{component: ReactNode}>> = new Map([
    [RouteIDs.Containers, {component: <><Icon icon={<IconBrandDocker stroke={1} />} /> Containers</>}],
    [RouteIDs.About, {component: <><Icon icon={<IconInfoHexagonFilled />} /> About</>}],
  ])

  return (
    <header className={styles.header}>
      <nav>
        {[...links.keys()].map(id => (
          <Link key={id} to={pathTo(id)} className={currentRouteID == id ? styles.activeLink : undefined}>
            {links.get(id)?.component}
          </Link>
        ))}
        {isUpdateAvailable && <Link to={__LATEST_RELEASE_LINK__} target="_blank">
          <Icon icon={<IconDownload />} /> Update available {latest && <>({latest.toString()})</>}
        </Link>}
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

const Footer = (): React.JSX.Element => {
  return <footer className={styles.footer} />
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
      <Header current={currentVersion} latest={latestVersion} />
      <Main />
      <Footer />
    </>
  )
}
