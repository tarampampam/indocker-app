import type { Emitter } from 'mitt'
import type { InjectionKey } from 'vue'
import { inject } from 'vue'

// key is the event name, value is the event payload
export type AppEvents = {
  dockerStateUpdated: undefined // the docker state has been updated, watch the changes in the IndexedDB
}

export const EmitterKey: InjectionKey<Emitter<AppEvents>> = Symbol('Emitter')

/** Resolve the emitter instance in VUE context. */
export function useEmitter(): Emitter<AppEvents> {
  const resolved = inject(EmitterKey)

  if (!resolved) {
    throw new Error(`Could not resolve ${EmitterKey.description}`)
  }

  return resolved
}
