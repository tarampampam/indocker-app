/// <reference lib="WebWorker" />

import deadDockerSvg from '~/shared/assets/dead-docker.svg?raw'

// kinda fuse
if (self.constructor.name.toLowerCase().includes('worker')) {
  const sw = self as unknown as ServiceWorkerGlobalScope & typeof globalThis

  const faviconBase64 =
    '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' +
    'TkSuQmCC'

  const fallback = new Response(
    `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <link rel="icon" type="image/png" sizes="32x32" href="data:image/png;base64,${faviconBase64}">
  <style>
    body, html {
      background-color: #fff;
      color: #010101;
      font-family: sans-serif;
      height: 100vh;
      padding: 0;
      margin: 0;
      font-size: 15px;
    }

    @media (prefers-color-scheme: dark) {
      body, html {
        background-color: #18181c;
        color: #dcdcdc;
      }
    }

    body {
      display: flex;
      align-items: center;
      justify-content: center;
      flex-direction: column;

      h1, h3 {
        margin: 0 0 0.5em 0;
        text-align: center;
      }

      article {
        margin: 1em 0 0 0;
        text-align: center;
        font-size: 0.8em;
        opacity: 0.7;

        p {
          margin: 0 0 1em 0;
        }

        .unregister {
          text-decoration: underline;
          cursor: pointer;
        }
      }

      picture {
        padding-bottom: 30px;

        img {
          width: 300px;
          height: auto;
          user-select: none;
          -webkit-user-drag: none;
        }
      }
    }

    @media screen and (min-width: 2000px) {
      body, html {
        font-size: 19px;
      }

      body {
        picture {
          img {
            width: 380px;
          }
        }
      }
    }
  </style>
  <meta http-equiv="refresh" content="3" />
  <title>Offline</title>
</head>
<body>
  <picture>
    <img src="data:image/svg+xml;base64,${btoa(deadDockerSvg)}" alt="Dead Docker" />
  </picture>
  <h1>The app isn't running</h1>
  <h3>Please, start indocker and try again</h3>
  <article>
    <p>
      You're seeing this page because the service worker has<br />
      replaced the default browser offline page with this one<br />
      <span class="unregister">(unregister service worker)</span>
    </p>
  </article>

  <script type="module">
    const unregisterElements = document.querySelectorAll('.unregister')

    if ('serviceWorker' in navigator) {
      const unregister = async () => {
        for (const reg of (await navigator.serviceWorker.getRegistrations())) {
          await reg.unregister()
        }

        reloadPage()
      }

      unregisterElements.forEach((el) => {
        el.addEventListener('click', () => unregister().catch(console.error))
      })
    } else {
      unregisterElements.forEach((el) => {
        el.disabled = true
        el.style.display = 'none'
      })
    }

    window.addEventListener('online', () => window.location.reload(), {passive: true, once: true})
  </script>
</body>
</html>`,
    {
      status: 521,
      statusText: 'Web Server Is Down',
      headers: {
        'Content-Type': 'text/html; charset=utf-8',
        'X-From-Service-Worker': 'true',
      },
    }
  )

  sw.addEventListener('install', async () => await sw.skipWaiting())

  sw.addEventListener('fetch', (event) => {
    // only call event.respondWith() if this is a navigation request for an HTML page AND not an API request
    if (event.request.mode === 'navigate' && event.request.url.toLowerCase().indexOf('/api/') === -1) {
      event.respondWith(fetch(event.request).catch(() => Promise.resolve(fallback.clone())))
    }
  })
}
