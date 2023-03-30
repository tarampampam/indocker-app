interface DiscoverResponse {
  base_url?: string
}

const defaultFetchOptions: RequestInit = {
  keepalive: true,
}

export async function discover(): Promise<DiscoverResponse> {
  const rnd = (Math.random() + 1).toString(36).substring(7)
  const req = new Request(
    `${location.protocol}//x-${rnd}.indocker.app/x/indocker/discover`,
    {...defaultFetchOptions, method: 'GET', headers: {'X-InDocker': 'true'}}
  )

  return (await fetch(req)).json()
}
