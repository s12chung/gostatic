# gostatic

[![GoDoc](https://godoc.org/github.com/s12chung/gostatic?status.svg)](https://godoc.org/github.com/s12chung/gostatic)
[![Build Status](https://travis-ci.com/s12chung/gostatic.svg?branch=master)](https://travis-ci.com/s12chung/gostatic)
[![Go Report Card](https://goreportcard.com/badge/github.com/s12chung/gostatic)](https://goreportcard.com/report/github.com/s12chung/gostatic)

Use Go and Webpack to generate static websites like a standard Go web application. You can even run `gostatic` apps, as a web app.

## Limitations

Meant to be simple, so defaults are given (SASS, image optimization, S3 deployment, etc.). You can rip out the defaults and use your own Go HTML templates, CSS preprocessor, etc.

## Requirements
- [Docker Desktop](https://www.docker.com) (if no Docker, see [`Dockerfile`](blueprint/Dockerfile) for system requirements)
- Optional [direnv](https://github.com/direnv/direnv) to automatically export/unexport ENV variables. You can export them yourself via `source ./.envrc`.

## New Project
You can install the `gostatic` binary via `go get` and run it:

```bash
go get github.com/s12chung/gostatic
...
gostatic new some_project_name
```

This is will create a new project in the current directory. After:

```bash
# without direnv, call: source ./.envrc
make docker-install
```

This will build docker and do all the installation inside docker. After, it’ll copy all downloaded code libraries from the docker container to the host, so that when the container and host sync filesystems, the libraries will be there.

You can run a developer instance through Docker via:
```bash
make docker
```

For production:
```bash
make docker-prod
```

By default, this will generate all static website files in the `generated` directory and host the directory in a file server at `http://localhost:3000`.

## Structure

The following is managed through a `Makefile`:

- A Go executable
- A Webpack setup
- `watchman` - A file diff watcher by Facebook to auto compile everything conveniently for development
- `aws-cli` - Handles Amazon S3 Deployment

First, Webpack compiles/optimizes all the assets (JS, CSS, images) in the `generated/assets` directory (`make build-assets`). It also generates a `Manifest.json` and a `images/responsive` folder of JSON files. These JSON files give the Go executable file paths to assets from the Webpack compilation.

After, the Go executable generates all the route files in the `generated` directory (`make build-go`) . See the [Go section](#go) below.

The Go executable also hosts the `generated` directory in a file server (`make file-server`).

You can also run your project as a web app (`make server`), which uses Go std lib `net/http` internally.

I run them through Docker, which handles all the system dependencies: Go, nodejs, image optimization, etc. See [`Dockerfile`](blueprint/Dockerfile) to see system dependencies. Your local system probably has some of them already, as Docker is running Alpine, a minimal Linux distribution.

## Go

The following packages are used in the bare bones Hello World app provided for you:

- [`cli`](https://godoc.org/github.com/s12chung/gostatic/go/cli) - Basic CLI interface for for your main.go
- [`app`](https://godoc.org/github.com/s12chung/gostatic/go/app) - Does high level commands of the [`cli.App` interface](https://godoc.org/github.com/s12chung/gostatic/go/cli#App) (generate, file-server, server) by taking your routes to generate files concurrently or serving it via http
- [`html`](https://godoc.org/github.com/s12chung/gostatic/go/lib/html) - Wrapper around Go std lib `html/template` to render templates, handle layouts, etc.
- [`webpack`](https://godoc.org/github.com/s12chung/gostatic/go/lib/webpack) - Lets Go see into the generated asset paths, `Manifest.json`, and `images/responsive` folder of JSON files from Webpack
- [`router`](https://godoc.org/github.com/s12chung/gostatic/go/lib/router) - Maps the URL paths to your functions like a http router, so that it can generate files or host a web app

It's best to start at [go/content/content.go](blueprint/go/content/content.go) and add more routes:

```go
func (content *Content) renderHTML(ctx router.Context, name string, layoutD interface{}) error {
	bytes, err := content.HTMLRenderer.Render(name, layoutD)
	if err != nil {
		return err
	}
	ctx.Respond(bytes)
	return nil
}

// SetRoutes is where you set the routes
func (content *Content) SetRoutes(r router.Router, tracker *app.Tracker) error {
	r.GetRootHTML(content.getRoot)
	r.GetHTML("/404.html", content.get404)
	r.GetHTML("/robots.txt", content.getRobots)
	return nil
}

func (content *Content) getRoot(ctx router.Context) error {
	return content.renderHTML(ctx, "root", layoutData{"", "Hello World!"})
}

func (content *Content) get404(ctx router.Context) error {
	return content.renderHTML(ctx, "404", layoutData{"404", nil})
}

func (content *Content) getRobots(ctx router.Context) error {
	// "github.com/s12chung/gostatic-packages/robots"
	// userAgents := []*robots.UserAgent {
	//	 robots.NewUserAgent(robots.EverythingUserAgent, []string { "/" }),
	// }
	//return ctx.Respond([]byte(robots.ToFileString(userAgents)))
	ctx.Respond([]byte{})
	return nil
}
```

See [`router.Router` interface](https://godoc.org/github.com/s12chung/gostatic/go/lib/router#Router) and the [`app.Setter` interface](https://godoc.org/github.com/s12chung/gostatic/go/app#Setter).
There are helpful packages in [s12chung/gostatic-packages](https://github.com/s12chung/gostatic-packages) too.

## Webpack

A [default webpack config](blueprint/webpack.config.js) is given to you, which handles assets in the `assets` directory. Below are defaults, via npm packages:

- [`gostatic-webpack`](https://github.com/s12chung/gostatic-webpack) - wraps the two libs below to configure Webpack for `gostatic` (manifest, entrypoints, Webpack optimizations, etc.)
- [`gostatic-webpack-css`](https://github.com/s12chung/gostatic-webpack-css) - handles the `.css` and `.scss` (SASS) files in the `assets/css` directory 
- [`gostatic-webpack-images`](https://github.com/s12chung/gostatic-webpack-images) - handles favicon files placement, image and compression in the `assets/favicon` and `assets/images` directory

Packages that you can easily rip out are:
- [`gostatic-webpack-babel`](https://github.com/s12chung/gostatic-webpack-babel) - handles javascript transpiling of the `js` directory

The `assets/js-extract` directory is intended for importing things you want Webpack to extract out (CSS, Images, etc.).

## Deploy

`gostatic` projects are designed to be hosted on Amazon S3. See [wiki on S3](https://github.com/s12chung/gostatic/wiki/S3-Config-Credentials) to setup S3 in the AWS Management Console, outside of the code.

Within your project, store the `Access Key ID` and `Secret Access Key` in `aws/credentials` of the project (see `aws/config` too), so `aws-cli` can use them. Set the `S3_BUCKET` in `.envrc`, so `Makefile` can see it. Then to upload everything to S3:

```bash
make docker-deploy
```

I usually use:

```bash
make push-docker-deploy
```

because it ensures my origin master is synced with my homepage. In the future, I’ll make `gostatic` projects deploy via Travis CI, so whenever origin master updates, the page updates.

## Host

You can host your project directly from Amazon S3. This is easier, but you won't have SSL and a CDN. See [S3 instructions on wiki](https://github.com/s12chung/gostatic/wiki/Hosting-via-S3-Directly).

I find it best to use Amazon CloudFront CDN with Amazon Certificate Manager to provide SSL. See [CloudFront instructions on wiki](https://github.com/s12chung/gostatic/wiki/Hosting-via-CloudFront).

Also, note about [CNAMEs on Root domains](https://serverfault.com/questions/613829/why-cant-a-cname-record-be-used-at-the-apex-aka-root-of-a-domain), which can break your emails.

## Projects
Here are my projects built with `gostatic`:

- [s12chung/go_homepage](https://github.com/s12chung/go_homepage) - my homepage, where most of this code was extracted from
- [s12chung/photoswipestory](https://github.com/s12chung/photoswipestory) - a static website as a photo book type project for a birthday present (so not my best code)

## Inspiration
Much inspiration was taken from [brandur/sorg](https://github.com/brandur/sorg).
