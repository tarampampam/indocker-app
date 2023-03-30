import semverParse from 'semver/functions/parse'
import type { SemVer } from 'semver'

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

export default class {
  private discovered: DiscoverResponse | null = null

  /** Returns the base URL of the API. */
  async baseUrl(): Promise<string> {
    if (!this.discovered) {
      this.discovered = await this.discover()
    }

    if (!this.discovered.api.base_url) {
      throw new Error('API base URL not discovered')
    }

    return this.discovered.api.base_url
  }

  /** Discovers the API. */
  async discover(): Promise<DiscoverResponse> {
    const rnd = (Math.random() + 1).toString(36).substring(3)
    const req = new Request(
      `${location.protocol}//x-${rnd}.indocker.app/x/indocker/discover`,
      {method: 'GET', headers: {'X-InDocker': 'true'}},
    )

    return (await fetch(req)).json()
  }

  /** Returns the current version of the app. */
  async version(): Promise<SemVer> {
    const resp = await (
      await fetch(new Request(`${await this.baseUrl()}/version/current`, {method: 'GET'}))
    ).json() as {
      version: string
    }

    // a little hack to make semver happy
    resp.version = resp.version.replace('@', '-')

    // parse version
    const version = semverParse(resp.version, {loose: true})

    // check version
    if (!version) {
      throw new Error(`Invalid version: ${resp.version}`)
    }

    return version
  }

  /** Returns the latest version of the app. */
  async latestVersion(): Promise<LatestVersion> {
    const resp = await (
      await fetch(new Request(`${await this.baseUrl()}/version/latest`, {method: 'GET'}))
    ).json() as {
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

    return {
      version: version,
      url: resp.url,
      name: resp.name,
      body: resp.body,
      created_at: Object.freeze(new Date(resp.created_at)),
    }
  }
}
