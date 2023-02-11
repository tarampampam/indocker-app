import {ContainerInfo, ContainerInspectInfo, ContainerStats, DockerVersion} from 'dockerode'
import {Base64} from 'js-base64'

const ansiCleanupRegex = /[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g

export default class {
  private baseUrl: string

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl
  }

  async ping(): Promise<boolean> {
    try {
      return (await fetch(new Request(`${this.baseUrl}/ping`, {method: 'GET'}))).ok
    } catch (_) {
      return false
    }
  }

  /** @throws {Error} */
  async version(): Promise<DockerVersion> {
    return await fetch(new Request(`${this.baseUrl}/version`, {method: 'GET'}))
      .then((resp) => resp.json())
  }

  /** @throws {Error} */
  async containersList(): Promise<ContainerInfo[]> {
    return await fetch(new Request(`${this.baseUrl}/containers/list`, {method: 'GET'}))
      .then((resp) => resp.json())
  }

  /** @throws {Error} */
  async inspect(id: string): Promise<ContainerInspectInfo> {
    return await fetch(new Request(`${this.baseUrl}/inspect?id=${id}`, {method: 'GET'}))
      .then((resp) => resp.json())
  }

  /** @throws {Error} */
  async stats(id: string): Promise<ContainerStats> {
    return await fetch(new Request(`${this.baseUrl}/stats?id=${id}`, {method: 'GET'}))
      .then((resp) => resp.json())
  }

  /** @throws {Error} */
  async logs(id: string, tail = 0, cleanup = true): Promise<string[]> {
    const b64lines = await (
      await fetch(new Request(`${this.baseUrl}/logs?id=${id}&tail=${tail}`, {method: 'GET'}))
    ).json() as string[]

    return b64lines
      .map((line: string) => Base64.decode(line))
      .map((line: string) => cleanup ? line.replace(ansiCleanupRegex, '') : line)
  }
}
