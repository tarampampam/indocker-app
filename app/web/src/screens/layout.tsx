import { Link, Outlet } from 'react-router-dom'
import type { SemVer } from 'semver'
import { type Client } from '~/api'
import React, { useEffect } from 'react'

export default function Layout({ apiClient }: { apiClient: Client }): React.JSX.Element {
  const [appVersion, setAppVersion] = React.useState<SemVer | null>(null)
  const [latestVersion, setLatestVersion] = React.useState<SemVer | null>(null)

  useEffect(() => {
    apiClient
      .currentVersion()
      .then((version) => setAppVersion(version))
      .catch(console.error)

    apiClient
      .latestVersion()
      .then((version) => setLatestVersion(version))
      .catch(console.error)
  }, [apiClient])

  return (
    <>
      <header>
        <Link to="/404notfound">404</Link>
      </header>
      <main>
        <Outlet />
      </main>
      <footer>
        <p>Version: {appVersion ? appVersion.toString() : '...'}</p>
        {appVersion && latestVersion && appVersion.compare(latestVersion) === -1 && (
          <p>A new version is available: {latestVersion.toString()}</p>
        )}
      </footer>
    </>
  )
}
