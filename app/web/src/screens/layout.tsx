import { AnimatePresence } from 'framer-motion'
import React, { type ReactNode, useEffect } from 'react'
import { Link, useLocation, useOutlet } from 'react-router-dom'
import { type SemVer } from 'semver'
import { type Client } from '~/api'
import { pathTo, RouteIDs, useCurrentRouteID } from '~/routing'
import { Icon, type IconProps } from '~/shared/components'
import { DockerRoundIcon, SoftwareUpdateIcon } from '~/assets/icons'
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
  const commonIconProps: Omit<IconProps, 'src'> = { style: { outer: { paddingRight: '.3em' } } }

  const links: ReadonlyMap<RouteIDs, Readonly<{ component: ReactNode }>> = new Map([
    [
      RouteIDs.Containers,
      {
        component: (
          <>
            <Icon src={DockerRoundIcon} {...commonIconProps} /> Containers
          </>
        ),
      },
    ],
    // [
    //   RouteIDs.About,
    //   {
    //     component: (
    //       <>
    //         <Icon src={AboutIcon} {...commonIconProps} />
    //         About
    //       </>
    //     ),
    //   },
    // ],
  ])

  return (
    <header className={styles.header}>
      <nav>
        {[...links.keys()].map((id) => (
          <Link key={id} to={pathTo(id)} className={currentRouteID == id ? styles.activeLink : undefined}>
            {links.get(id)?.component}
          </Link>
        ))}
        {isUpdateAvailable && (
          <Link to={__LATEST_RELEASE_LINK__} target="_blank">
            <Icon src={SoftwareUpdateIcon} {...commonIconProps} />
            Update available {latest && <>({latest.toString()})</>}
          </Link>
        )}
      </nav>
    </header>
  )
}

const Main = (): React.JSX.Element => {
  const [location, outlet] = [useLocation(), useOutlet()]

  return (
    <main className={styles.main}>
      <AnimatePresence mode="wait" initial={true}>
        {outlet && React.cloneElement(outlet, { key: location.pathname })}
      </AnimatePresence>
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
