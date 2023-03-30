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
      const resp = http.request('TRACE', `${scheme}://${randomString(8)}.indocker.app/discover`, null, {
        headers: {'X-InDocker': 'true'},
        tags: {scheme},
      })

      check(resp, {
        'status is 200': (r) => r.status === 200,
        'content-type is json': (r) => r.headers['Content-Type'].includes('application/json'),
        'CORS Origin header': (r) => r.headers['Access-Control-Allow-Origin'] === '*',
        'CORS Methods header': (r) => r.headers['Access-Control-Allow-Methods'] === 'TRACE',
        'CORS Headers header': (r) => r.headers['Access-Control-Allow-Headers'] === 'X-InDocker',
      })

      /** @type {{base_url: string}} */
      const body = JSON.parse(resp.body.toString())

      check(body, {
        'base url is set': (b) => b.base_url !== undefined,
        'base url is correct': (b) => new RegExp(`^${scheme}:\\/\\/[a-zA-Z0-9-_]+\\.indocker\\.app$`).test(b.base_url),
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
