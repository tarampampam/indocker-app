import type { Table } from 'dexie'
import Dexie from 'dexie'
import type { ContainerInspectInfo, ContainerStats } from 'dockerode'
import type { InjectionKey } from 'vue'
import { inject } from 'vue'

interface DockerState {
  [id: string]: {
    inspect: Readonly<ContainerInspectInfo>
    stats: Readonly<ContainerStats>
  }
}

// interface ContainerMemoryUsage {
//   ts: Readonly<Date> // timestamp
//   ids: Readonly<{
//     [id: string]: { // container id
//       usage: Readonly<number> // memory usage in bytes
//     }
//   }[]>
// }

export class DB extends Dexie {
  protected readonly docker_state!: Table<{ ids: string[]; state: DockerState }, /* unix timestamp */number>
  protected readonly timeline_labels!: Table</* human-readable time */string, /* unix timestamp */number>
  protected readonly timeline_overall_memory_usage!: Table<{ ts: number; memory: number }[], /* container id */string>

  public ttlSec: number = 60 * 15 // TTL for the collected records

  private dbVersion = 1

  constructor() {
    super('app')

    this.version(this.dbVersion).stores({
      docker_state: '&,ids', // the primary key is a timestamp
      timeline_labels: '&', // timestamp (primary key) <=> formatted timestamp map (used for the chart labels)
      timeline_overall_memory_usage: '&,ts', // container id <=> memory usage array mapping
    })
  }

  /** Cleans up old records. */
  async cleanup(): Promise<void> {
    const now = Date.now()

    await this.transaction('rw', this.docker_state, this.timeline_labels, async () => {
      await this.docker_state
        .where(':id')
        .below(now - this.ttlSec * 1000)
        .delete()

      await this.timeline_labels
        .where(':id')
        .below(now - this.ttlSec * 1000)
        .delete()
    })
  }

  async putDockerState(state: DockerState): Promise<number> {
    const now = new Date

    const result = this.transaction<number>('rw', this.docker_state, this.timeline_labels, async () => {
      await this.timeline_labels.add(`${
        ("0" + now.getHours()).slice(-2)
      }:${
        ("0" + now.getMinutes()).slice(-2)
      }:${
        ("0" + now.getSeconds()).slice(-2)
      }`, now.valueOf())

      // put the new record
      return this.docker_state.add({ids: Object.keys(state), state: state}, now.valueOf())
    })

    // delete old records
    await this.cleanup()

    return result
  }

  // async getContainersMemoryUsage(): Promise<Readonly<ContainerMemoryUsage[]>> {
  //   const records: ContainerMemoryUsage[] = []
  //
  //   await this.docker_state
  //     .orderBy(':id')
  //     .each((record, ts) => {
  //       const ids = Object.keys(record.state)
  //       const item = {ts: Object.freeze(new Date(ts.primaryKey)), ids: new Array(ids.length)}
  //
  //       for (let i = 0; i < ids.length; i++) {
  //         const id = ids[i]
  //
  //         item.ids[i] = {
  //           [id]: {
  //             usage: record.state[id].stats.memory_stats.usage,
  //           },
  //         }
  //       }
  //
  //       Object.seal(item.ids)
  //
  //       records.push(item)
  //     })
  //
  //   return records
  // }
}

export const DBKey: InjectionKey<DB> = Symbol('DB')

/** Resolve the DB instance in VUE context. */
export function useDB(): DB {
  const resolved = inject(DBKey)

  if (!resolved) {
    throw new Error(`Could not resolve ${DBKey.description}`)
  }

  return resolved
}
