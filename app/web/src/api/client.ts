import createClient, { type Client as OpenapiClient, type ClientOptions } from 'openapi-fetch'
import { coerce as semverCoerce, parse as semverParse, type SemVer } from 'semver'
import { APIErrorUnknown } from './errors'
import { throwIfNotJSON, throwIfNotValidResponse } from './middleware'
import { components, paths } from './schema.gen'

type ContainerRoutesList = ReadonlyMap<string, ReadonlyArray<URL>>

export class Client {
  private readonly baseUrl: URL
  private readonly api: OpenapiClient<paths>
  private cache: Partial<{
    currentVersion: Readonly<SemVer>
    latestVersion: Readonly<SemVer>
  }> = {}

  constructor(opt?: ClientOptions) {
    this.baseUrl = new URL(
      opt?.baseUrl ? opt.baseUrl.replace(/\/+$/, '') : window.location.protocol + '//' + window.location.host
    )

    this.api = createClient<paths>(opt)
    this.api.use(throwIfNotJSON, throwIfNotValidResponse)
  }

  /**
   * Returns the version of the app.
   *
   * @throws {APIError}
   */
  async currentVersion(force: boolean = false): Promise<Readonly<SemVer>> {
    if (this.cache.currentVersion && !force) {
      return this.cache.currentVersion
    }

    const { data, response } = await this.api.GET('/api/version')

    if (data) {
      const version = semverParse(semverCoerce(data.version.replace('@', '-')))

      if (!version) {
        throw new APIErrorUnknown({ message: `Failed to parse the current version value: ${data.version}`, response })
      }

      this.cache.currentVersion = Object.freeze(version)

      return this.cache.currentVersion
    }

    throw new APIErrorUnknown({ message: response.statusText, response }) // will never happen
  }

  /**
   * Returns the latest available version of the app.
   *
   * @throws {APIError}
   */
  async latestVersion(force: boolean = false): Promise<Readonly<SemVer>> {
    if (this.cache.latestVersion && !force) {
      return this.cache.latestVersion
    }

    const { data, response } = await this.api.GET('/api/version/latest')

    if (data) {
      const version = semverParse(semverCoerce(data.version))

      if (!version) {
        throw new APIErrorUnknown({ message: `Failed to parse the latest version value: ${data.version}`, response })
      }

      this.cache.latestVersion = Object.freeze(version)

      return this.cache.latestVersion
    }

    throw new APIErrorUnknown({ message: response.statusText, response }) // will never happen
  }

  /**
   * Returns the list of all registered routes.
   *
   * @throws {APIError}
   */
  async routesList(): Promise<ContainerRoutesList> {
    const { data, response } = await this.api.GET('/api/routes')

    if (data) {
      const map = new Map<string, ReadonlyArray<URL>>()

      for (const route of data.routes) {
        map.set(route.hostname, Object.freeze(route.urls.map((url) => Object.freeze(new URL(url)))))
      }

      // sort the map by keys before returning it
      return Object.freeze(new Map([...map.entries()].sort()))
    }

    throw new APIErrorUnknown({ message: response.statusText, response }) // will never happen
  }

  /**
   * Subscribe to route changes via WebSocket.
   *
   * The promise resolves with a closer function that can be called to close the WebSocket connection.
   *
   *
   * */
  async routesSubscribe({
    onConnected,
    onUpdate,
    onError,
  }: {
    onConnected?: () => void
    onUpdate: (routes: ContainerRoutesList) => void
    onError?: (err: Error) => void
  }): Promise</* closer */ () => void> {
    const protocol = this.baseUrl.protocol === 'https:' ? 'wss:' : 'ws:'
    const path: keyof paths = '/api/routes/subscribe'

    return new Promise((resolve: (closer: () => void) => void, reject: (err: unknown) => void) => {
      const ws = new WebSocket(`${protocol}//${this.baseUrl.host}${path}`)

      ws.onopen = (): void => {
        onConnected?.()
        resolve((): void => ws.close())
      }

      ws.onerror = (err): void => {
        onError?.(new Error(err instanceof ErrorEvent ? err.message : String(err)))
        reject(err) // will be ignored if the promise is already resolved
      }

      ws.onmessage = (event): void => {
        if (event.data) {
          const content = JSON.parse(event.data) as components['schemas']['ContainerRoutesList']
          const map = new Map<string, ReadonlyArray<URL>>()

          for (const route of content.routes) {
            map.set(route.hostname, Object.freeze(route.urls.map((url) => Object.freeze(new URL(url)))))
          }

          // sort the map by keys before calling the callback
          onUpdate(Object.freeze(Object.freeze(new Map([...map.entries()].sort()))))
        }
      }
    })
  }
}

export default new Client() // singleton instance
