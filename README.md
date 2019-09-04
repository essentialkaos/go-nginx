<p align="center"><a href="#readme"><img src="https://gh.kaos.st/go-nginx.svg"/></a></p>

<p align="center">
  <a href="https://godoc.org/pkg.re/essentialkaos/nginx.v0"><img src="https://godoc.org/pkg.re/essentialkaos/nginx.v0?status.svg"></a>
  <a href="https://goreportcard.com/report/github.com/essentialkaos/nginx"><img src="https://goreportcard.com/badge/github.com/essentialkaos/nginx"></a>
  <a href="https://travis-ci.org/essentialkaos/nginx"><img src="https://travis-ci.org/essentialkaos/nginx.svg"></a>
  <a href="https://coveralls.io/github/essentialkaos/nginx?branch=master"><img src="https://coveralls.io/repos/github/essentialkaos/nginx/badge.svg?branch=master" alt="Coverage Status" /></a>
  <a href="https://essentialkaos.com/ekol"><img src="https://gh.kaos.st/ekol.svg" alt="License" />
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#build-status">Build Status</a> • <a href="#license">License</a></p>

<br/>

`nginx` is a Go package for reading Nginx configuration files.

**Note, that this is beta software, so it's entirely possible that there will be some significant bugs. Please report bugs so that we are aware of the issues.**

### Installation

Before the initial install allows git to use redirects for [pkg.re](https://github.com/essentialkaos/pkgre) service (_reason why you should do this described [here](https://github.com/essentialkaos/pkgre#git-support)_):

```
git config --global http.https://pkg.re.followRedirects true
```

Make sure you have a working Go 1.11+ workspace (_[instructions](https://golang.org/doc/install)_), then:

```
go get pkg.re/essentialkaos/nginx.v0
```

For update to the latest stable release, do:

```
go get -u pkg.re/essentialkaos/nginx.v0
```

### Build Status

| Branch | Status |
|--------|--------|
| `master` | [![Build Status](https://travis-ci.org/essentialkaos/nginx.svg?branch=master)](https://travis-ci.org/essentialkaos/nginx) |
| `develop` | [![Build Status](https://travis-ci.org/essentialkaos/nginx.svg?branch=develop)](https://travis-ci.org/essentialkaos/nginx) |

### License

[EKOL](https://essentialkaos.com/ekol)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
