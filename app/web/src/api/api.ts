import type { ContainerInspectInfo, ContainerStats } from 'dockerode'
import semverParse from 'semver/functions/parse'
import type { SemVer } from 'semver'
import ReconnectingWebSocket from 'reconnecting-websocket'
import type { InjectionKey } from 'vue'
import { inject } from 'vue'

interface DiscoverResponse {
  api: {
    base_url?: string
  }
}

interface LatestVersion {
  version: SemVer
  url: string
  name: string
  body: string
  created_at: Date
}

interface ContainerState {
  inspect: ContainerInspectInfo
  stats: ContainerStats
}

export class API {
  /** Returns the base URL of the API. */
  baseUrl(type: 'http' | 'ws' = 'http'): Readonly<string> {
    const proto = type === 'http' ? location.protocol : location.protocol.replace('http', 'ws')
    const port = location.port ? `:${location.port}` : ''

    return Object.freeze(`${proto}//${location.host}${port}/api`) // without trailing slash
  }

  /** Discovers the API. */
  async discover(): Promise<Readonly<DiscoverResponse>> {
    const rnd = (Math.random() + 1).toString(36).substring(3)
    const req = new Request(`${location.protocol}//x-${rnd}.indocker.app/x/indocker/discover`, {
      method: 'GET',
      headers: { 'X-InDocker': 'true' }
    })

    return Object.freeze((await fetch(req)).json())
  }

  /** Returns the current version of the app. */
  async version(): Promise<Readonly<SemVer>> {
    const resp = (await (
      await fetch(new Request(`${this.baseUrl()}/version/current`, { method: 'GET' }))
    ).json()) as {
      version: string
    }

    // a little hack to make semver happy
    resp.version = resp.version.replace('@', '-')

    // parse version
    const version = semverParse(resp.version, { loose: true })

    // check version
    if (!version) {
      throw new Error(`Invalid version: ${resp.version}`)
    }

    return Object.freeze(version)
  }

  /** Returns the latest version of the app. */
  async latestVersion(): Promise<Readonly<LatestVersion>> {
    const resp = (await (
      await fetch(new Request(`${this.baseUrl()}/version/latest`, { method: 'GET' }))
    ).json()) as {
      version: string
      url: string
      name: string
      body: string
      created_at: string
    }

    // parse version
    const version = semverParse(resp.version)

    // check version
    if (!version) {
      throw new Error(`Invalid version: ${resp.version}`)
    }

    return Object.freeze({
      version: Object.freeze(version),
      url: resp.url,
      name: resp.name,
      body: resp.body,
      created_at: Object.freeze(new Date(resp.created_at))
    })
  }

  /** Returns a WebSocket connection to the Docker state updates. */
  watchDockerState(
    onMessage: (map: Readonly<{ [id: string]: ContainerState }>) => void
  ): ReconnectingWebSocket {
    const ws = new ReconnectingWebSocket(`${this.baseUrl('ws')}/ws/docker/state`, undefined, {
      maxReconnectionDelay: 10000
    })

    ws.addEventListener('message', (msg): void => {
      onMessage(Object.freeze(JSON.parse(msg.data)))
    })

    return ws
  }
}

export const APIKey: InjectionKey<API> = Symbol('API')

/** Resolve the API instance in VUE context. */
export function useAPI(): API {
  const resolved = inject(APIKey)

  if (!resolved) {
    throw new Error(`Could not resolve ${APIKey.description}`)
  }

  return resolved
}
