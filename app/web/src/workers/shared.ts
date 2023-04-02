import { DB } from '@/db/db'
import { API } from '@/api/api'

export const DockerStateChannelName: string = 'docker-state'

const ctx: SharedWorkerGlobalScope = self as any

// execute only in shared worker context
if (typeof ctx === 'object' && ctx.constructor.name.toLowerCase().includes('worker')) { // kinda fuse
  const db = new DB()
  const api = new API()
  const br = new BroadcastChannel(DockerStateChannelName)

  api.watchDockerState((map) => {
    db.putDockerState(map)
      .then((dbKey) => br.postMessage(dbKey))
      .catch(console.error)
  })

  ctx.addEventListener('error', console.error)
  ctx.addEventListener('connect', console.debug)
}

// ctx.addEventListener('connect', (event) => {
//   const port = event.ports[0]
//
//   port.addEventListener('message', (e) => {
//     port.postMessage(e.data)
//   })
//
//   port.start()
// })
