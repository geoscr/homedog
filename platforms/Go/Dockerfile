FROM golang:1.11 AS base
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 \
    && chmod +x /usr/local/bin/dep

FROM base AS dep
WORKDIR /dep-setup/src/homedog
ADD src/homedog/Gopkg.toml src/homedog/Gopkg.lock ./
ENV GOPATH=/dep-setup
RUN /usr/local/bin/dep ensure -v -vendor-only

FROM dep AS build
WORKDIR /app
ENV GOPATH /app
ENV PATH /app/bin:$PATH
COPY --from=dep /dep-setup/src/homedog/vendor /app/src/homedog/vendor
COPY --from=dep /dep-setup/pkg                /app/pkg
WORKDIR /app
#ADD src/homedog /app/src/homedog
#RUN go get github.com/derekparker/delve/cmd/dlv

FROM build AS installer
WORKDIR /app
ENV GOPATH /app
ENV PATH /app/bin:$PATH
ADD src/homedog /app/src/homedog
RUN go install homedog

CMD [ "/app/bin/homedog" ]
