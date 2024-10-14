import createClient, { type Client as OpenapiClient, type ClientOptions } from 'openapi-fetch'
import { coerce as semverCoerce, parse as semverParse, type SemVer } from 'semver'
import { APIErrorUnknown } from './errors'
import { throwIfNotJSON, throwIfNotValidResponse } from './middleware'
import { components, paths } from './schema.gen'

type ContainerRoutesList = ReadonlyMap<string, ReadonlyMap<string, URL>> // map<hostname, map<container_id, url>>

export class Client {
  private readonly baseUrl: URL
  private readonly api: OpenapiClient<paths>
  private cache: Partial<{
    currentVersion: Readonly<SemVer>
    latestVersion: Readonly<SemVer>
    favicons: Map<string, Readonly<string>> // map[base_url]favicon_base64
  }> = {
    favicons: new Map(),
  }

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

    throw new APIErrorUnknown({ message: response.statusText, response }) // will never happen due to the middleware
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

    throw new APIErrorUnknown({ message: response.statusText, response }) // will never happen due to the middleware
  }

  /**
   * Returns the list of all registered routes.
   *
   * @throws {APIError}
   */
  async routesList(): Promise<ContainerRoutesList> {
    const { data, response } = await this.api.GET('/api/routes')

    if (data) {
      const map = new Map<string, Map<string, URL>>()

      for (const route of data.routes) {
        map.set(
          route.hostname,
          Object.freeze(
            Object.entries(route.urls).reduce(
              (map, [containerID, url]) => map.set(containerID, Object.freeze(new URL(url))),
              new Map<string, URL>()
            )
          )
        )
      }

      // sort the map by keys before returning it
      return Object.freeze(new Map([...map.entries()].sort()))
    }

    throw new APIErrorUnknown({ message: response.statusText, response }) // will never happen due to the middleware
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
    onConnected?: () => void // called when the WebSocket connection is established
    onUpdate: (routes: ContainerRoutesList) => void // called when the routes are updated
    onError?: (err: Error) => void // called when an error occurs on alive connection
  }): Promise</* closer */ () => void> {
    const protocol = this.baseUrl.protocol === 'https:' ? 'wss:' : 'ws:'
    const path: keyof paths = '/api/routes/subscribe'

    return new Promise((resolve: (closer: () => void) => void, reject: (err: Error) => void) => {
      let connected: boolean = false

      try {
        const ws = new WebSocket(`${protocol}//${this.baseUrl.host}${path}`)

        ws.onopen = (): void => {
          connected = true
          onConnected?.()
          resolve((): void => ws.close())
        }

        ws.onerror = (event: Event): void => {
          // convert Event to Error
          const err = new Error(event instanceof ErrorEvent ? String(event.error) : 'WebSocket error')

          if (connected) {
            onError?.(err)
          }

          reject(err) // will be ignored if the promise is already resolved
        }

        ws.onmessage = (event): void => {
          if (event.data) {
            const content = JSON.parse(event.data) as components['schemas']['ContainerRoutesList']
            const map = new Map<string, Map<string, URL>>()

            for (const route of content.routes) {
              map.set(
                route.hostname,
                Object.freeze(
                  Object.entries(route.urls).reduce(
                    (map, [containerID, url]) => map.set(containerID, Object.freeze(new URL(url))),
                    new Map<string, URL>()
                  )
                )
              )
            }

            // sort the map by keys before calling the callback
            onUpdate(Object.freeze(Object.freeze(new Map([...map.entries()].sort()))))
          }
        }
      } catch (e) {
        // convert any exception to Error
        const err = e instanceof Error ? e : new Error(String(e))

        if (connected) {
          onError?.(err)
        }

        reject(err)
      }
    })
  }

  /** Returns the favicon (in base64) for the given base URL. */
  async getFaviconFor(hostname: string, force: boolean = false): Promise<string | null> {
    if (this.cache.favicons && this.cache.favicons.has(hostname) && !force) {
      const cached = this.cache.favicons.get(hostname)

      if (cached) {
        return cached
      }
    }

    const { response, data } = await this.api.GET('/api/favicon/{hostname}', {
      params: { path: { hostname: hostname.replace(/\/$/, '') } }, // remove trailing slash
      parseAs: 'blob',
      priority: 'low',
      signal: AbortSignal.timeout(10000), // 10 seconds request timeout
    })

    if (response.status === 204) {
      return null
    }

    if (data && typeof data === 'object' && data instanceof Blob) {
      const reader = new FileReader()

      const promise = new Promise<string>((resolve, reject) => {
        reader.onloadend = () => {
          if (typeof reader.result === 'string') {
            resolve(reader.result)
          } else {
            reject(new Error('Failed to read the favicon blob'))
          }
        }
      })

      reader.readAsDataURL(data)

      return promise.then((base64) => {
        const frozen = Object.freeze(base64)

        this.cache.favicons?.set(hostname, frozen)

        return frozen
      })
    }

    throw new APIErrorUnknown({ message: response.statusText, response }) // will never happen due to the middleware
  }
}

export default new Client() // singleton instance
