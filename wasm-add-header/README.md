# Wasm Add Header

Wasm Add Header is a Go WASM extension example that will add a specific header to the request and response. 

## Pre-requisite

Make sure you have [TinyGo Installed](https://tinygo.org/getting-started/install/) in your env

## Build WASM Extension

To build the WASM extension and create a `.wasm` executable, run:

```sh
make build
```

### Clean Build Artifacts

To clean build artifacts, run:

```sh
make clean
```

### Run Tests

```sh
make test
```

### Install Dependencies

```sh
make get
```

### Build Docker Image

To build a Docker image from the project, run:

```sh
make docker-build
```

### Push Docker Image

To tag the Docker image and push it to a Docker regitry, run:

```sh
make docker-push
```