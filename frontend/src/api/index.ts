import {ContainerInfo, ContainerInspectInfo, ContainerStats, DockerVersion} from 'dockerode'
import {Base64} from 'js-base64'

const ansiCleanupRegex = /[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g

export interface DockerState {
  created_at: string
  containers: {
    [container_id: string]: {
      inspect: {
        cmd?: any
        env: string[]
        hostname: string
        labels: { [key: string]: string }
        user: string
        created: string
        id: string
        image: string
        name: string
        restart_count: number
        exit_code: number
        health_status: string
        failing_streak: number
        oom_killed: boolean
        dead: boolean
        paused: boolean
        restarting: boolean
        running: boolean
        pid: number
        status: string
      }
      stats: {
        read: string
        num_procs: number
        cpu_usage: number
        memory_usage: number
        memory_max_usage: number
        memory_limit: any
        network_rx_bytes: number
        network_rx_errors: number
        network_tx_bytes: number
        network_tx_errors: number
      }
    }
  }
}

export default class {
  private baseUrl: string

  private defaultFetchOptions: RequestInit = {
    keepalive: true,
  }

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl
  }

  /**
   * @link https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
   */
  streamDockerState(callback: (data: DockerState[]) => void, interval: string = '1000ms'): EventSource {
    const sse = new EventSource(`${this.baseUrl}/stream-docker-state?interval=${interval}`)

    sse.addEventListener('message', (event) => {
      callback((JSON.parse(event.data) as DockerState[]))
    })

    return sse
  }

  /**
   * @link https://docs.docker.com/engine/api/v1.42/#tag/System/operation/SystemPing
   */
  async ping(): Promise<boolean> {
    try {
      return (await fetch(new Request(`${this.baseUrl}/ping`, {...this.defaultFetchOptions, method: 'GET'}))).ok
    } catch (_) {
      return false
    }
  }

  /**
   * @link https://docs.docker.com/engine/api/v1.42/#tag/System/operation/SystemVersion
   * @throws {Error}
   */
  async version(): Promise<DockerVersion> {
    return await fetch(new Request(`${this.baseUrl}/version`, {...this.defaultFetchOptions, method: 'GET'}))
      .then((resp) => resp.json())
  }

  /**
   * @link https://docs.docker.com/engine/api/v1.42/#tag/Container/operation/ContainerList
   * @throws {Error}
   */
  async containersList(): Promise<ContainerInfo[]> {
    return await fetch(new Request(`${this.baseUrl}/containers/list`, {...this.defaultFetchOptions, method: 'GET'}))
      .then((resp) => resp.json())
  }

  /**
   * @link https://docs.docker.com/engine/api/v1.42/#tag/Container/operation/ContainerInspect
   * @throws {Error}
   */
  async inspect(id: string): Promise<ContainerInspectInfo> {
    return await fetch(new Request(`${this.baseUrl}/inspect?id=${id}`, {...this.defaultFetchOptions, method: 'GET'}))
      .then((resp) => resp.json())
  }

  /**
   * @link https://docs.docker.com/engine/api/v1.42/#tag/Container/operation/ContainerStats
   * @throws {Error}
   */
  async stats(id: string): Promise<ContainerStats> {
    return await fetch(new Request(`${this.baseUrl}/stats?id=${id}`, {...this.defaultFetchOptions, method: 'GET'}))
      .then((resp) => resp.json())
  }

  /**
   * @link https://docs.docker.com/engine/api/v1.42/#tag/Container/operation/ContainerLogs
   * @throws {Error}
   */
  async logs(id: string, tail = 0, cleanup = true): Promise<string[]> {
    const b64lines = await (
      await fetch(new Request(`${this.baseUrl}/logs?id=${id}&tail=${tail}`, {
        ...this.defaultFetchOptions,
        method: 'GET',
      }))
    ).json() as string[]

    return b64lines
      .map((line: string) => Base64.decode(line))
      .map((line: string) => cleanup ? line.replace(ansiCleanupRegex, '') : line)
  }
}
