export enum EventNames {
  MemoryUsageUpdated = 'memory-usage-updated',
}

export interface MemoryUsageUpdatedEvent {
  containers: {
    [containerID: string]: {
      totalMemoryUsageInBytes: number
    }
  }
}
