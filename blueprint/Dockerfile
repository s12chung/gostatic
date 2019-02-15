FROM golang:1.11-alpine3.9

ARG DOCKER_WORKDIR

RUN echo '@testing http://nl.alpinelinux.org/alpine/edge/testing' >> /etc/apk/repositories &\
    apk --no-cache add\
    # go
    git make dep\
    # webpack
    nodejs nodejs-npm yarn\
    # aws
    aws-cli@testing\
    # imagemin-webpack-plugin
    optipng gifsicle\
    # https://github.com/imagemin/gifsicle-bin/issues/79
    autoconf automake\
    # lovell/sharp image resizing (runtime)
    fftw-dev vips-dev@testing\
    # lovell/sharp image resizing (install/when yarn runs)
    g++ python2

RUN yarn global add webpack webpack-cli

RUN mkdir -p /var/run/watchman/root-state

EXPOSE 3000
EXPOSE 8080

WORKDIR $DOCKER_WORKDIR
COPY . .

# install watchman from custom build because of: https://github.com/facebook/watchman/issues/602
RUN apk add ./watchman/watchman-4.7.0-r0.apk --allow-untrusted

RUN make install
