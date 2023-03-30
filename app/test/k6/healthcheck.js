import { check, group } from 'k6'
import http from 'k6/http'

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
    group(`Discover the API (${scheme})`, () => {
      const resp = http.get(`${scheme}://any-domain.indocker.app/healthz`, {
        headers: {
          'User-Agent': 'HealthChecker/indocker',
        },
        tags: {
          scheme,
        },
      })

      check(resp, {
        'status is 200': (r) => r.status === 200,
        'body is OK': (r) => r.body.toString() === 'OK',
      })
    })
  })
}
