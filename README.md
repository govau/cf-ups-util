# cf-ups-util

CloudFoundry CLI plugin to list all users in a CloudFoundry installation

## Install from binary

Pick as appropriate for your OS:

```bash
cf install-plugin https://github.com/govau/cf-ups-util/releases/download/v0.2.0/ups-util.linux32
cf install-plugin https://github.com/govau/cf-ups-util/releases/download/v0.2.0/ups-util.linux64
cf install-plugin https://github.com/govau/cf-ups-util/releases/download/v0.2.0/ups-util.osx
cf install-plugin https://github.com/govau/cf-ups-util/releases/download/v0.2.0/ups-util.win32
cf install-plugin https://github.com/govau/cf-ups-util/releases/download/v0.2.0/ups-util.win64
```

## Install from source

```bash
go get github.com/govau/cf-ups-util/cmd/ups-util
cf install-plugin $GOPATH/bin/ups-util -f
```

## Usage

```bash
cf target -o org -s space
cf ups-util
```

Or:

```bash
cf ups-util app1 app2 app3
```

## Development

```bash
go install ./cmd/ups-util && \
    cf install-plugin $GOPATH/bin/ups-util -f && \
    cf ups-util
```

## Building a new release

```bash
PLUGIN_PATH=$GOPATH/src/github.com/govau/cf-ups-util/cmd/ups-util
PLUGIN_NAME=$(basename $PLUGIN_PATH)
cd $PLUGIN_PATH

GOOS=linux GOARCH=amd64 go build -o ${PLUGIN_NAME}.linux64
GOOS=linux GOARCH=386 go build -o ${PLUGIN_NAME}.linux32
GOOS=windows GOARCH=amd64 go build -o ${PLUGIN_NAME}.win64
GOOS=windows GOARCH=386 go build -o ${PLUGIN_NAME}.win32
GOOS=darwin GOARCH=amd64 go build -o ${PLUGIN_NAME}.osx

shasum -a 1 ${PLUGIN_NAME}.*
```
