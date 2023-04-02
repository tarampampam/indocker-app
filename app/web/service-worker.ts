const ctx: ServiceWorkerGlobalScope = self as any // eslint-disable-line @typescript-eslint/no-explicit-any
const CACHE = 'offline-fallback-v1' // https://habr.com/en/company/2gis/blog/345552/

// execute ONLY in service worker context
if (typeof ctx === 'object' && ctx.constructor.name.toLowerCase().includes('worker')) { // kinda fuse
  ctx.addEventListener('install', (event) => {
    event.waitUntil(
      caches
        .open(CACHE)
        .then((cache) => cache.addAll([
          '/favicon.ico',
          '/site.webmanifest',
          '/apple-touch-icon.png',
        ]))
        .then(() => ctx.skipWaiting()),
    )
  })

  ctx.addEventListener('activate', (event) => {
    event.waitUntil(ctx.clients.claim())
  })

  const networkOrCache = (request: RequestInfo | URL) => {
    return fetch(request)
      .then((response) => response.ok ? response : fromCache(request))
      .catch(() => fromCache(request))
  }

  const offline = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <style>
      :root{--color-bg-primary:#fff;--color-text-primary:#010101}
      @media (prefers-color-scheme:dark){:root{--color-bg-primary:#202124;--color-text-primary:#f5f6f8}}
      body,html{background-color:var(--color-bg-primary);color:var(--color-text-primary);font-family:sans-serif;
      height:100vh;padding:0;margin:0;font-size:15px;text-rendering:optimizeLegibility}
      .flex-center{display:flex;align-items:center;justify-content:center}main{flex-direction:column}
      main h1,main h3{margin:0 0 10px;text-align:center;font-family:Inter,-apple-system,BlinkMacSystemFont,'Segoe UI',
      Roboto,Oxygen,Ubuntu,Cantarell,'Fira Sans','Droid Sans','Helvetica Neue',sans-serif}
      main .icon{width:100px;height:100px;padding-bottom:30px}
    </style>
    <meta http-equiv="refresh" content="5" />
    <title>Offline</title>
  </head>
  <body class="flex-center">
    <main class="flex-center">
      <svg xmlns="http://www.w3.org/2000/svg" shape-rendering="geometricPrecision" image-rendering="optimizeQuality"
        fill-rule="evenodd" viewBox="0 0 512 512" class="icon">
        <circle transform="matrix(2.325512 2.325512 -2.325512 2.325512 256.000568 255.999957)" fill="#ef4136" r="77.84"/>
        <path
          d="M162.03 311.94l-28.73 28.73 39.61 39.62 29.11-29.12c30.83 25.41 60.09 27.74 87.12 18.45L143.65 224.13c-9.12
           26.74-6.66 55.93 18.38 87.81zm-31.46-150.11c-6.42-6.42-6.42-16.83 0-23.25s16.83-6.42 23.25 0l27.22
           27.22c9.99-11.21 20.7-21.87 30.85-32.03l21.7 21.69 49.46-49.45c8.33-8.34 21.98-8.34 30.32 0 8.33 8.33
           8.33 21.98 0 30.31l-49.47 49.47 63.1 63.09 49.46-49.46c8.34-8.34 21.98-8.34 30.32 0s8.33 21.98 0
           30.31l-49.47 49.46 22.5 22.5c-10 10-20.74 20.7-32.14 30.74L375 359.76c6.42 6.41 6.42 16.83 0
           23.25-6.42 6.41-16.84 6.41-23.25 0L130.57 161.83z"
          fill="#fff"
        />
      </svg>
      <h1>The app isn't running</h1>
      <h3>Please, start indocker and try again</h3>
    </main>
  </body>
</html>`

  const useFallback = () => {
    return Promise.resolve(new Response(offline, {
      headers: {
        'Content-Type': 'text/html; charset=utf-8',
      },
    }))
  }

  const fromCache = (request: RequestInfo | URL) => {
    return caches
      .open(CACHE)
      .then((cache) =>
        cache.match(request).then((matching) =>
          matching || Promise.reject('no-match'),
        ),
      )
  }

  ctx.addEventListener('fetch', (event) => {
    event.respondWith(networkOrCache(event.request).catch(() => useFallback()))
  })
}

export {}
