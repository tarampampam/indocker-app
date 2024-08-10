import createClient, { type Client as OpenapiClient, type ClientOptions } from 'openapi-fetch'
import { coerce as semverCoerce, parse as semverParse, type SemVer } from 'semver'
import { APIErrorUnknown } from './errors'
import { throwIfNotJSON, throwIfNotValidResponse } from './middleware'
import { components, paths } from './schema.gen'

type ContainerRoutesList = ReadonlyMap<string, ReadonlyArray<URL>>

export class Client {
  private readonly baseUrl: URL
  private readonly api: OpenapiClient<paths>

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
  async currentVersion(): Promise<Readonly<SemVer>> {
    const { data, response } = await this.api.GET('/api/version')

    if (data) {
      const version = semverParse(semverCoerce(data.version.replace('@', '-')))

      if (!version) {
        throw new APIErrorUnknown({ message: 'Failed to parse the version', response })
      }

      return Object.freeze(version)
    }

    throw new APIErrorUnknown({ message: response.statusText, response }) // will never happen
  }

  /**
   * Returns the latest available version of the app.
   *
   * @throws {APIError}
   */
  async latestVersion(): Promise<Readonly<SemVer>> {
    const { data, response } = await this.api.GET('/api/version/latest')

    if (data) {
      const version = semverParse(semverCoerce(data.version))

      if (!version) {
        throw new APIErrorUnknown({ message: 'Failed to parse the version', response })
      }

      return Object.freeze(version)
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
  async routesSubscribe(onUpdate: (routes: ContainerRoutesList) => void): Promise</* closer */ () => void> {
    const protocol = this.baseUrl.protocol === 'https:' ? 'wss:' : 'ws:'
    const path: keyof paths = '/api/routes/subscribe'

    const ws = new WebSocket(`${protocol}//${this.baseUrl.host}${path}`)

    return new Promise((resolve: (closer: () => void) => void, reject: (err: unknown) => void) => {
      ws.onopen = (): void => resolve((): void => ws.close())
      ws.onerror = (err): void => reject(err)
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
