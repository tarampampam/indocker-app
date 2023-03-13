import {ContainerInfo, ContainerInspectInfo, ContainerStats, DockerVersion} from 'dockerode'
import {Base64} from 'js-base64'

const ansiCleanupRegex = /[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g

export default class {
  private baseUrl: string

  private defaultFetchOptions: RequestInit = {
    keepalive: true,
  }

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl
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
