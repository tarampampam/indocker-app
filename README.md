<div align="center">
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://socialify.git.ci/tarampampam/indocker-app/image?description=1&font=Raleway&forks=1&issues=1&owner=1&pulls=1&pattern=Solid&stargazers=1&theme=Dark">
  <img src="https://socialify.git.ci/tarampampam/indocker-app/image?description=1&font=Raleway&forks=1&issues=1&owner=1&pulls=1&pattern=Solid&stargazers=1">
</picture>

<p>&nbsp;</p>

[![tests][badge-tests]][actions]
[![release][badge-release]][actions]
[![docker][badge-docker]][quay]
[![license][badge-license]][license]
</div>

> This project previously was called `localhost.tools`

One time you may want to run docker containers locally and interact with they're using domain names instead of
different TCP/UDP ports. And in addition - you want to use HTTPS protocol for this. This is what this project
is about.

I've registered a domain name `indocker.app` and configured it to point any subdomain to your local machine:

```bash
$ dig +noall +answer -t A foo.indocker.app # IPv4
foo.indocker.app.	7131	IN	A	127.0.0.1

$ dig +noall +answer -t AAAA foo.indocker.app # IPv6
foo.indocker.app.	86400	IN	AAAA	::1

$ dig +noall +answer foo.bar.baz.indocker.app # any depth
foo.bar.baz.indocker.app. 86400	IN	A	127.0.0.1
```

## Ok, but how to use it?

It's very simple :) To connect locally run docker containers with subdomains in zone `indocker.app`, we need a
reverse proxy server, which will be able to handle requests to `*.indocker.app` and forward them to the required
**locally run** docker containers.

So I decided to use [traefik][traefik] with a wildcard SSL certificate from [Let's Encrypt][letsencrypt], put it
together in a docker image, and publish it.

As you probably know, traefik has a built-in feature named "configuration discovery", which allows you to configure
it using docker container labels (that's why it needs a docker socket to be mounted).

All we have to do is run it, the necessary containers with the required labels, and Viola! It works! More details
you can find on the website of the project: [indocker.app](https://indocker.app/).

## Where is the docker image?

Here: `quay.io/indocker/app` _(~40 Mb download size)_.

From time to time (about once every 2 months) you need to update it. To do this, you need to run the following command:

```bash
$ docker pull quay.io/indocker/app:<used-tag>
```

## Additional features

- You can mount any [dynamic config file][dynamic-configuration] (with your middlewares, routers, and services)
  to the `/etc/traefik/dynamic` directory inside the image, and it will be automatically loaded by traefik
  (eg.: `/etc/traefik/dynamic/my-custom-config.yaml`)

### License

This is open-sourced software licensed under the [MIT License][license].

[badge-tests]:https://img.shields.io/github/actions/workflow/status/tarampampam/indocker-app/tests.yml?branch=master&maxAge=30&label=tests&logo=github&style=flat-square
[badge-release]:https://img.shields.io/github/actions/workflow/status/tarampampam/indocker-app/release.yml?maxAge=30&label=release&logo=github&style=flat-square
[badge-docker]:https://shields.io/static/v1?label=Docker%20image&message=quay.io%2Findocker%2Fapp&color=blue&style=flat-square
[badge-license]:https://img.shields.io/github/license/tarampampam/indocker-app.svg?maxAge=30&style=flat-square

[actions]:https://github.com/tarampampam/indocker-app/actions
[quay]:https://quay.io/indocker/app
[license]:https://github.com/tarampampam/indocker-app/blob/master/LICENSE
[traefik]:https://github.com/traefik/traefik
[letsencrypt]:https://letsencrypt.org/
[dynamic-configuration]:https://doc.traefik.io/traefik/getting-started/configuration-overview/#the-dynamic-configuration
