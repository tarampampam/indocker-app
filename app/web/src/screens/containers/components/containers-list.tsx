import React, { useEffect, useState } from 'react'
import { DockerRoundIcon, ShieldIcon } from '~/assets/icons'
import { Icon } from '~/shared/components'
import type { Client } from '~/api'
import styles from './containers-list.module.scss'

export type ContainerListItem = {
  name: string
  url: URL
  routes?: ReadonlyMap<string, URL> // map[containerID]url
}

const Row = ({ apiClient, item }: { apiClient: Client; item: ContainerListItem }): React.JSX.Element => {
  const [icon, setIcon] = useState<string>(DockerRoundIcon)

  // on component mount, get the favicon
  useEffect(() => {
    if (item.routes && item.routes.size > 0) {
      apiClient
        // get the first route's favicon (all routes should have the same favicon)
        .getFaviconFor(item.routes.values().next().value.toString())
        .then((base64) => setIcon(base64))
        .catch(() => {
          /* do nothing */
        })
    }
  }, [apiClient, item.routes])

  return (
    <div className={styles.row}>
      <div className={styles.col}>
        <div className={styles.segment}>
          <Icon src={icon} />
        </div>
        <div className={styles.segment}>{item.name}</div>
        {item.routes && item.routes.size > 1 && (
          <div className={styles.segment} style={{ opacity: 0.5 }}>
            {'//'} {item.routes.size} containers
          </div>
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
              <a href={item.url.toString()} target="_blank" rel="noreferrer">
                {item.url
                  .toString()
                  .replace(/\/+$/, '')
                  .replace(/https?:\/\//, '')}
              </a>
            </div>
          </>
        )}
      </div>
    </div>
  )
}

export default function Component({
  apiClient,
  items,
}: {
  apiClient: Client
  items: ReadonlyArray<ContainerListItem>
}): React.JSX.Element {
  return (
    <div className={styles.list}>
      {items.map((item, i) => (
        <Row key={i} apiClient={apiClient} item={item} />
      ))}
    </div>
  )
}
