# WASM Extensions in TSB(Istio)

This project demonstrate how WASM extensions can be created and integrated into an Istio Gateway using TSB as a service mesh.


## WASM Plugin to modify Request/Response Header

[wasm-add-header](./wasm-add-header/) is a Go WASM plugin example that will add a specific header to the request and response. 

### Build the WASM Plugin

*Pre-req:* Make sure you have [TinyGo Installed](https://tinygo.org/getting-started/install/) in your env

To build the WASM plugin and create a `.wasm` executable, run:

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
