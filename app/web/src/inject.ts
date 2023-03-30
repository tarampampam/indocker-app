import type { InjectionKey } from 'vue'
import { inject } from 'vue'
import type { API } from './api/api'

export const APIKey: InjectionKey<API> = Symbol('API')

/** Safely inject a dependency. */
export function safeInject<T>(key: InjectionKey<T>): T {
  const resolved = inject(key)

  if (!resolved) {
    throw new Error(`Could not resolve ${key.description}`)
  }

  return resolved
}
