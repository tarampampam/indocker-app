import createClient, { type ClientOptions, type Client as OpenapiClient } from 'openapi-fetch'
import { parse as semverParse, coerce as semverCoerce, type SemVer } from 'semver'
import { APIErrorUnknown } from './errors.ts'
import { throwIfNotJSON, throwIfNotValidResponse } from './middleware'
import { paths } from './schema.gen'

export class Client {
  private readonly api: OpenapiClient<paths>

  constructor(opt?: ClientOptions) {
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
  async routesList(): Promise<ReadonlyMap<string, ReadonlyArray<URL>>> {
    const { data, response } = await this.api.GET('/api/routes')

    if (data) {
      const map = new Map<string, ReadonlyArray<URL>>()

      for (const route of data.routes) {
        map.set(
          route.hostname,
          route.urls.map((url) => Object.freeze(new URL(url)))
        )
      }

      map.set('foo', []) // TODO: remove this line
      map.set('aaa.foo', []) // TODO: remove this line
      map.set('bbb.foo', []) // TODO: remove this line
      map.set('111.bbb.foo', []) // TODO: remove this line
      map.set('qqq.www.eee', [new URL('https://example.com')]) // TODO: remove this line

      return Object.freeze(map)
    }

    throw new APIErrorUnknown({ message: response.statusText, response }) // will never happen
  }
}

export default new Client() // singleton instance
