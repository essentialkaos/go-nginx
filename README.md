<a href="#readme"><img src="https://gh.kaos.st/beta-alert.svg"/></a>

<p align="center"><a href="#readme"><img src="https://gh.kaos.st/go-nginx.svg"/></a></p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/essentialkaos/go-nginx"><img src="https://pkg.go.dev/badge/github.com/essentialkaos/go-nginx"></a>
  <a href="https://goreportcard.com/report/github.com/essentialkaos/go-nginx"><img src="https://goreportcard.com/badge/github.com/essentialkaos/go-nginx"></a>
  <a href="https://github.com/essentialkaos/go-nginx/actions"><img src="https://github.com/essentialkaos/go-nginx/workflows/CI/badge.svg" alt="GitHub Actions Status" /></a>
  <a href="https://coveralls.io/github/essentialkaos/go-nginx?branch=master"><img src="https://coveralls.io/repos/github/essentialkaos/go-nginx/badge.svg?branch=master" alt="Coverage Status" /></a>
  <a href="https://codebeat.co/projects/github-com-essentialkaos-go-nginx-master"><img alt="codebeat badge" src="https://codebeat.co/badges/7b8cb5a7-2b9d-426f-8637-4f2bd5644a4d" /></a>
  <a href="#license"><img src="https://gh.kaos.st/apache2.svg"></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#build-status">Build Status</a> • <a href="#license">License</a></p>

<br/>

`nginx` is a Go package for reading Nginx configuration files.

### Installation

Before the initial install allows git to use redirects for [pkg.re](https://github.com/essentialkaos/pkgre) service (_reason why you should do this described [here](https://github.com/essentialkaos/pkgre#git-support)_):

```
git config --global http.https://pkg.re.followRedirects true
```

Make sure you have a working Go 1.12+ workspace (_[instructions](https://golang.org/doc/install)_), then:

```
go get pkg.re/essentialkaos/go-nginx.v0
```

For update to the latest stable release, do:

```
go get -u pkg.re/essentialkaos/go-nginx.v0
```

### Build Status

| Branch | Status |
|--------|--------|
| `master` | [![CI](https://github.com/essentialkaos/go-nginx/workflows/CI/badge.svg?branch=master)](https://github.com/essentialkaos/go-nginx/actions) |
| `develop` | [![CI](https://github.com/essentialkaos/go-nginx/workflows/CI/badge.svg?branch=develop)](https://github.com/essentialkaos/go-nginx/actions) |

### License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
