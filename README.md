# gostatic [![Build Status](https://travis-ci.com/s12chung/gostatic.svg?branch=master)](https://travis-ci.com/s12chung/gostatic)

Use Go and Webpack to generate static websites like a standard Go web application. You can even run `gostatic` apps, as a web app.

## Limitations

Meant to be simple, so defaults are given (SASS, image optimization, S3 deployment, etc.). You can rip the defaults out and use your own Go HTML templates, CSS preprocessor, etc. Router is only 1 level deep (`page.com` and `page.com/about` are ok, but you can't `page.com/projects/about`). Feel free to make a PR to add more to the [router](https://godoc.org/github.com/s12chung/gostatic/go/lib/router). I just never needed the feature yet.

## Requirements
- [Docker Desktop](https://www.docker.com) (if no Docker, see [`Dockerfile`](Dockerfile) for system requirements)
- Optional [direnv](https://github.com/direnv/direnv) to automatically export/unexport ENV variables or export them yourself via (`source ./.envrc`)

## New Project
First, install the gostatic binary via:

```bash
go get github.com/s12chung/gostatic
go install github.com/s12chung/gostatic
```

Then run:
```bash
${GOPATH}/bin/gostatic init some_project_name
```

This is will create a new project in the current directory. After:

```bash
# if you have direnv, allow setting ENV variables, otherwise call:
# source ./.envrc
make docker-install
```

This will build docker and do all the installation inside docker. After, it’ll copy all downloaded code libraries from the docker container to the host, so that when the container and host sync filesystems, the libraries will be there.

You can run a developer instance through Docker via:
```sh
make docker
```

For production:
```sh
make docker-prod
```

By default, this will generate all the files of the static website in `generated` and host the directory in a file server via `http://localhost:3000`.

## Structure

The following is managed through a `Makefile`:

- A Go executable
- A Webpack setup
- `watchman` - A file diff watcher by Facebook to auto compile everything conveniently for development
- `aws-cli` - Handles Amazon S3 Deployment

First, Webpack compiles/optimizes all the assets (JS, CSS, images) in the `assets` folder (`make build-assets`). It also generates a `Manifest.json` and a `images/responsive` folder of JSON files. These files give the Go executable file paths from the Webpack compilation.

With the Webpack JSON files, the Go executable generates all the web page files by defining routes (`make build-go`). See the [Go section](#go) below.

By default, all the generated results from Webpack and Go are put in the `generated` folder. The Go executable can host these files with in file server (`make file-server`).

You can also run your project as a web app (`make server`), which uses Go std lib `net/http` internally.

I run them through Docker, which handles all the system dependencies: Go, nodejs, image optimization, etc. See [`Dockerfile`](blueprint/Dockerfile) to see system dependencies. Your local system probably has some of them already, as Docker is running Alpine, a minimal Linux distribution.

## Go

The following packages are used in the bare bones Hello World app provided for you:

- [`cli`](https://godoc.org/github.com/s12chung/gostatic/go/cli) - Basic CLI interface for for your main.go
- [`app`](https://godoc.org/github.com/s12chung/gostatic/go/app) - Does high level commands of the [`cli.App` interface](https://godoc.org/github.com/s12chung/gostatic/go/cli#App) (generate, file-server, server) by taking your routes to generate files concurrently or serving it via http
- [`html`](https://godoc.org/github.com/s12chung/gostatic/go/lib/html) - Wrapper around Go std lib `html/template` to render templates, handle layouts, etc.
- [`webpack`](https://godoc.org/github.com/s12chung/gostatic/go/lib/webpack) - Maps paths generated asset paths, `Manifest.json`, and `images/responsive` folder of JSON files
- [`router`](https://godoc.org/github.com/s12chung/gostatic/go/lib/router) - Maps the URL paths to your functions like a http router, so that it can generate files or host a web app

It's best to start at [go/content/content.go](blueprint/go/content/content.go) and add more routes:

```go
func (content *Content) RenderHtml(ctx router.Context, name, defaultTitle string, data interface{}) error {
	bytes, err := content.HtmlRenderer.Render(name, defaultTitle, data)
	if err != nil {
		return err
	}
	return ctx.Respond(bytes)
}

func (content *Content) SetRoutes(r router.Router, tracker *app.Tracker) {
	r.GetRootHTML(content.getRoot)
	r.GetHTML("/404.html", content.get404)
	r.GetHTML("/robots.txt", content.getRobots)
}

func (content *Content) getRoot(ctx router.Context) error {
	return content.RenderHtml(ctx, "root", "", "Hello World!")
}

func (content *Content) get404(ctx router.Context) error {
	return content.RenderHtml(ctx, "404", "404", nil)
}

func (content *Content) getRobots(ctx router.Context) error {
	return ctx.Respond([]byte{})
}
```

See [`router.Router` interface](https://godoc.org/github.com/s12chung/gostatic/go/lib/router#Router) and the [`app.Setter` interface](https://godoc.org/github.com/s12chung/gostatic/go/app#Setter). There are other [helpful packages too](go/lib).

## Webpack

A [default webpack config](blueprint/webpack.config.js) is given to you. Below are defaults, via npm packages:

- [`gostatic-webpack`](https://github.com/s12chung/gostatic-webpack) - wraps the two libs below to configure Webpack for `gostatic` (manifest, entrypoints, Webpack optimizations, etc.)
- [`gostatic-webpack-css`](https://github.com/s12chung/gostatic-webpack-css) - handles the `.css` and `.scss` (SASS) files in the `assets/css` directory 
- [`gostatic-webpack-images`](https://github.com/s12chung/gostatic-webpack-images) - handles favicon files placement, image and compression in the `assets/favicon` and `assets/images` directory

Packages that you can easily rip out are:
- [`gostatic-webpack-babel`](https://github.com/s12chung/gostatic-webpack-babel) - handles javascript transpiling of the `js` directory

The `js-extract` directory is intended for importing things you want Webpack to extract out (CSS, Images, etc.)

## Deploy

`gostatic` projects are designed to be hosted on Amazon S3. See [wiki on S3](https://github.com/s12chung/gostatic/wiki/S3-Config-Credentials) to setup S3 in the AWS Management Console, outside of the code.

Within your project, store the `Access Key ID` and `Secret Access Key` in `aws/credentials` of the project (see `aws/config` too), so `aws-cli` can use them. Set the `S3_BUCKET` in `.envrc`, so `Makefile` can see it. Then to upload everything to S3:

```sh
make docker-deploy
```

I usually use:

```sh
make push-docker-deploy
```

because it ensures my origin master is synced with my homepage. In the future, I’ll make `gostatic` projects deploy via Travis CI, so whenever origin master updates, the page updates.

## Host

You can host your project directly from Amazon S3. This is easier, but you won't have SSL and a CDN. See [S3 instructions on wiki](https://github.com/s12chung/gostatic/wiki/Hosting-via-S3-Directly).

I find it best to use Amazon CloudFront CDN with Amazon Certificate Manager to provide SSL. See [CloudFront instructions on wiki](https://github.com/s12chung/gostatic/wiki/Hosting-via-CloudFront).

Also, note about [CNAMEs on Root domains](https://serverfault.com/questions/613829/why-cant-a-cname-record-be-used-at-the-apex-aka-root-of-a-domain), which can break your emails.

## Inspiration
Much inspiration was taken from [brandur/sorg](https://github.com/brandur/sorg).
