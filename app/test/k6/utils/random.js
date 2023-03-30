/**
 * Returns a random integer between min (inclusive) and max (inclusive).
 *
 * @param {number} min
 * @param {number} max
 * @returns {number}
 */
export function randomIntBetween(min, max) { // min and max included
  return Math.floor(Math.random() * (max - min + 1) + min)
}

/**
 * Generates a random string of the specified length.
 *
 * @param {number} length
 * @param {string} charset
 * @returns {string}
 */
export function randomString(length, charset = 'abcdefghijklmnopqrstuvwxyz0123456789') {
  let res = ''

  while (length--) {
    res += charset[(Math.random() * charset.length) | 0]
  }

  return res
}
