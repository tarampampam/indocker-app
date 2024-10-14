<div align="center">
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://socialify.git.ci/tarampampam/indocker-app/image?description=1&font=Raleway&forks=1&issues=1&owner=1&pulls=1&pattern=Solid&stargazers=1&theme=Dark">
  <img src="https://socialify.git.ci/tarampampam/indocker-app/image?description=1&font=Raleway&forks=1&issues=1&owner=1&pulls=1&pattern=Solid&stargazers=1">
</picture>

[![tests][badge-tests]][actions]
[![release][badge-release]][actions]
[![docker][badge-docker]][quay]
[![license][badge-license]][license]
</div>

[badge-tests]:https://img.shields.io/github/actions/workflow/status/tarampampam/indocker-app/tests.yml?branch=master&maxAge=30&label=tests&logo=github&style=flat-square
[badge-release]:https://img.shields.io/github/actions/workflow/status/tarampampam/indocker-app/release.yml?maxAge=30&label=release&logo=github&style=flat-square
[badge-docker]:https://shields.io/static/v1?label=Docker%20image&message=quay.io%2Findocker%2Fapp&color=blue&style=flat-square
[badge-license]:https://img.shields.io/github/license/tarampampam/indocker-app.svg?maxAge=30&style=flat-square

> [!INFO]
> This project was previously called `localhost.tools`.

One time, you may want to run Docker containers locally and interact with them using domain names instead of
`127.0.0.1:1234`, for example. Additionally, you may want to use the HTTPS protocol. That’s what this project
aims to provide.

> [!WARNING]
> This project is under development and may not work as expected. For instance, the current implementation does
> not support WebSockets, gRPC, or other connection types - only HTTP/HTTPS is supported.

## How does it work?

Technically, this project is a simple reverse proxy server. It listens to all incoming requests to domains
like `*.indocker.app` and forwards them to the corresponding Docker containers running on your local machine.
To make this work, I've registered the domain `indocker.app` and configured it so any subdomain points to your
local machine (`127.0.0.1` for IPv4 and `::1` for IPv6):

```bash
$ dig +noall +answer -t A foo.indocker.app # IPv4
foo.indocker.app.	7131	IN	A	127.0.0.1

$ dig +noall +answer -t AAAA foo.indocker.app # IPv6
foo.indocker.app.	86400	IN	AAAA	::1

$ dig +noall +answer foo.bar.baz.indocker.app # any depth
foo.bar.baz.indocker.app. 86400	IN	A	127.0.0.1
```

This eliminates the need for modifying the `hosts` file or using additional software to resolve domain names locally.
All you need to do is run the indocker app and configure your Docker containers to be accessible via domain names
using **Docker labels**.

Here’s an example of how the routing works:

- You send an HTTP request to `https://foo.indocker.app`
- `foo.indocker.app` resolves to `127.0.0.1` (your local machine)
- The indocker app on your local machine (listening on ports 443 and 80) receives the request and forwards it to
  the appropriate Docker container based on the domain name

> [!INFO]
> More examples can be found in the [examples](examples) directory.

## What about HTTPS?

To enable HTTPS, I’ve generated a wildcard SSL certificate for the `*.indocker.app` domain, signed by
[Let's Encrypt][letsencrypt]. The indocker app uses this certificate to encrypt all incoming requests.

The certificate is automatically renewed periodically, and the app downloads it each time it starts, so you don’t
need to worry about managing it.

[letsencrypt]: https://letsencrypt.org/

### License

This is open-sourced software licensed under the [MIT License][license].

[license]:https://github.com/tarampampam/indocker-app/blob/master/LICENSE

[actions]:https://github.com/tarampampam/indocker-app/actions
[quay]:https://quay.io/indocker/app
