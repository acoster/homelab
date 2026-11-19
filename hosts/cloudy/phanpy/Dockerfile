# Phanpy (https://github.com/cheeaun/phanpy), a static Mastodon-API web client,
# from the upstream release tarball, served by nginx. Built on cloudy by
# doco-cd (`compose build`); nothing to compile, so it's cheap on the VPS.
#
# No third-party image on purpose: this JavaScript holds the users' tokens.
#
# To update: pick a tag from https://github.com/cheeaun/phanpy/releases, then
#   curl -sL https://github.com/cheeaun/phanpy/releases/download/<tag>/phanpy-dist.tar.gz | sha256sum
# and change both ARGs, plus the image tag in compose.yaml.
FROM nginx:stable-alpine

ARG PHANPY_VERSION=2026.10.09.febb38d
ARG PHANPY_SHA256=42a45f0df92ac10840f3064c46d60e34b2ab2029574ee4593062cfcc1663a9e0

ADD https://github.com/cheeaun/phanpy/releases/download/${PHANPY_VERSION}/phanpy-dist.tar.gz /tmp/phanpy.tar.gz
RUN echo "${PHANPY_SHA256}  /tmp/phanpy.tar.gz" | sha256sum -c - \
 && rm -rf /usr/share/nginx/html/* \
 && tar -xzf /tmp/phanpy.tar.gz -C /usr/share/nginx/html \
 && rm /tmp/phanpy.tar.gz \
 && chmod -R a+rX /usr/share/nginx/html
