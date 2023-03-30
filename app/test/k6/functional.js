import { check, group } from 'k6'
import http from 'k6/http'
import { randomString } from './utils/random.js'

/**
 * Default options.
 *
 * @link https://k6.io/docs/using-k6/k6-options/how-to/#examples-of-setting-options
 * @link https://k6.io/docs/using-k6/k6-options/reference/
 *
 * @type {import('k6/options').Options}
 */
export const options = {
  scenarios: {
    default: {
      executor: 'per-vu-iterations',
    },
  },
};

/**
 * VU (test) code. Called once per iteration.
 *
 * @link https://k6.io/docs/using-k6/test-lifecycle/
 * @link https://k6.io/docs/using-k6/execution-context-variables/
 */
export default () => {
  ['http', 'https'].forEach((scheme) => {
    group('Discover the API', () => {
      const domain = randomString(8)

      let resp = http.options(`${scheme}://${domain}.indocker.app/x/indocker/discover`, null, {
        tags: {scheme},
      })

      check(resp, {
        'status is 200': (r) => r.status === 204,
        'body is empty': (r) => r.body.length === 0,
        'CORS Methods header': (r) => r.headers['Access-Control-Allow-Methods'] === 'GET',
        'CORS Headers header': (r) => r.headers['Access-Control-Allow-Headers'] === '*',
        'CORS Origin header': (r) => r.headers['Access-Control-Allow-Origin'] === `${scheme}://monitor.indocker.app`,
      })

      resp = http.get(`${scheme}://${domain}.indocker.app/x/indocker/discover`, {
        headers: {'X-InDocker': 'true'},
        tags: {scheme},
      })

      check(resp, {
        'status is 200': (r) => r.status === 200,
        'is json': (r) => r.headers['Content-Type'].includes('application/json'),
        'CORS Methods header': (r) => r.headers['Access-Control-Allow-Methods'] === 'GET',
        'CORS Headers header': (r) => r.headers['Access-Control-Allow-Headers'] === '*',
        'CORS Origin header': (r) => {
          return new RegExp(`^${scheme}:\\/\\/[a-zA-Z0-9-_]+\\.indocker\\.app$`).test(
            r.headers['Access-Control-Allow-Origin'],
          )
        },
      })

      /** @type {{api: {base_url: string}}} */
      const body = JSON.parse(resp.body.toString())

      check(body, {
        'base url is set': (b) => b.api.base_url !== undefined,
        'base url is correct': (b) => {
          return new RegExp(`^${scheme}:\\/\\/[a-zA-Z0-9-_]+\\.indocker\\.app\\/api$`).test(b.api.base_url)
        },
      })
    })

    group('Healthcheck', () => {
      const resp = http.get(`${scheme}://any-domain.indocker.app/healthz`, {
        headers: {'User-Agent': 'HealthChecker/indocker'},
        tags: {scheme},
      })

      check(resp, {
        'status is 200': (r) => r.status === 200,
        'body is OK': (r) => r.body.toString() === 'OK',
      })
    })
  })
}
