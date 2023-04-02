import type { Table } from 'dexie'
import Dexie from 'dexie'
import type { ContainerInspectInfo, ContainerStats } from 'dockerode'

export interface DockerState {
  [id: string]: {
    inspect: ContainerInspectInfo
    stats: ContainerStats
  }
}

export class DB extends Dexie {
  public docker_state!: Table<{ ids: string[]; state: DockerState }, number>

  public dockerStateRecordsTTLSec: number = 60 * 15 // TTL for docker state records

  private dbVersion = 1

  constructor() {
    super('app')

    this.version(this.dbVersion).stores({
      docker_state: '&,ids', // the primary key is a timestamp
    })
  }

  async putDockerState(state: DockerState): Promise<number> {
    const pkey = Date.now()

    return this.transaction<number>('rw', this.docker_state, async () => {
      // put the new record
      const id = await this.docker_state.add({ ids: Object.keys(state), state: state }, pkey)

      // delete old records
      await this.docker_state
        .where(':id')
        .below(pkey - this.dockerStateRecordsTTLSec * 1000)
        .delete()

      return id
    })
  }
}
