import React, { useEffect, useState } from 'react'
import { DockerRoundIcon, ShieldIcon } from '~/assets/icons'
import { Icon } from '~/shared/components'
import type { Client } from '~/api'
import styles from './containers-list.module.scss'

type TableItem = {
  name: string
  url: URL
  routes?: ReadonlyMap<string, URL> // map[containerID]url
}

const Row = ({ apiClient, item }: { apiClient: Client; item: TableItem }): React.JSX.Element => {
  const [icon, setIcon] = useState<string>(DockerRoundIcon)

  // on component mount, get the favicon
  useEffect(() => {
    apiClient
      .getFaviconFor(item.url.toString())
      .then((base64) => setIcon(base64))
      .catch(() => {
        /* do nothing */
      })
  }, [apiClient, item.url])

  return (
    <div className={styles.row}>
      <div className={styles.col}>
        <div className={styles.segment}>
          <Icon src={icon} />
        </div>
        <div className={styles.segment}>{item.name}</div>
        {item.routes && item.routes.size > 1 && (
          <div className={styles.segment}>(containers count: {item.routes.size})</div>
        )}
      </div>
      <div className={styles.col}>
        {item.url && (
          <>
            {item.url.protocol === 'https:' && (
              <div className={styles.segment}>
                <Icon src={ShieldIcon} />
              </div>
            )}
            <div className={styles.segment}>
              {item.url.port ? item.url.port.replace(/^:/, '') : item.url.protocol === 'https:' ? '443' : '80'}
            </div>
            <div className={styles.segment}>
              <a href={item.url.toString()} target="_blank" rel="noreferrer">
                {item.url.toString()}
              </a>
            </div>
          </>
        )}
      </div>
    </div>
  )
}

export default function Component({ apiClient }: { apiClient: Client }): React.JSX.Element {
  return (
    <div className={styles.list}>
      <Row
        apiClient={apiClient}
        item={{
          name: 'name1',
          url: new URL('http://status1.com:123'),
          routes: new Map([
            ['769c041f8685e91cee965832d46e9bdd5dccd98e759fe8b8691440a714a4972f', new URL('http://172.19.0.2:8080')],
            ['f16c09e38a8a4d63669ac5638708691865d9ef6a56f2e20f95a21f86c2cfc442', new URL('http://172.19.0.3:8080')],
          ]),
        }}
      />
      <Row
        apiClient={apiClient}
        item={{
          name: 'Zame2',
          url: new URL('https://google.com'),
          routes: new Map([
            ['f16c09e38a8a4d63669ac5638708691865d9ef6a56f2e20f95a21f86c2cfc442', new URL('http://172.19.0.3:8080')],
          ]),
        }}
      />
      <Row
        apiClient={apiClient}
        item={{
          name: 'name3',
          url: new URL('https://albinakoch.com'),
          routes: new Map([['id3', new URL('http://172.19.0.4:8080')]]),
        }}
      />
    </div>
  )
}
