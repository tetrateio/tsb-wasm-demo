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

## How to use WASM Plugin in Istio Ingress Gateway

Since we will be using [Tetrate Service Bridge](https://docs.tetrate.io/service-bridge/next) service mesh to create Istio IngressGateway, you need to configure the following TSB resources to integrate WASM extensions within TSB.

### Create WasmExtension CR

[WasmExtension](./tsb/wasm/wasm-extension.yaml) configures the CR which can then be used inside a TSB Istio IngressGateway resource 

```yaml
apiVersion: extension.tsb.tetrate.io/v2
kind: WasmExtension
metadata:
  name: wasm-add-header
  annotations:
    tsb.tetrate.io/organization: tetrate
spec:
  image: oci://docker.io/sreeharikmarar/wasm-add-header:latest
  source: https://github.com/sreeharikmarar/tsb-wasm-demo
  config:
    header: x-ingress-header
    value: "powered by TSB"
  description: |
    This WASM plugin will add specified header in response
    To use this add following into IngressGateway, Tier1Gateway or SecuritySettings

    ```
      extension:
        - fqn: "organizations/tetrate/extensions/wasm-add-header"
          config:
            path: response
            header: x-ingress-header
            value: from tsb ingress gateway
    ````
    You must set a header and a value.
```

### Create IngressGateway CR

[IngressGateway](./tsb/ingress-gateway/tsb-config.yaml) configures the TSB Istio IngressGateway CR and other related configuration object for it to function. 

Gateway expose the `httpbin` app deployed on `httpbin` namespace on host `httpbin.tetrate.io`.

As you can see, previously applied `WasmExtension` has been used as `wasmPlugins` in `IngressGateway`. You can configure path as `request` or `response` in the plugin configration to add the configured header to the request or response when Istio ingress gateway proxy (Envoy) intercept the request.

```yaml
    apiVersion: gateway.tsb.tetrate.io/v2
    kind: Gateway
    metadata:
      name: tier2-gateway
      annotations:
        tsb.tetrate.io/organization: tetrate
        tsb.tetrate.io/tenant: tier2
        tsb.tetrate.io/workspace: tier2-ws
        tsb.tetrate.io/gatewayGroup: tier2-gg
    spec:
      displayName: Tier2 Gateway
      workloadSelector:
        namespace: tier2
        labels:
          app: tier2-gateway
      http:
        - hostname: httpbin.tetrate.io
          name: httpbin
          port: 80
          routing:
            rules:
              - route:
                  serviceDestination:
                    host: "httpbin/httpbin.httpbin.svc.cluster.local"
                    port: 8000
      wasmPlugins:
        - fqn: "organizations/tetrate/extensions/wasm-add-header"
          config:
            path: response
            header: x-ingress-header
            value: from tsb ingress gateway
```

### Apply All Configurations as K8s CRs

When you enable `gitops` in your application cluster, all these configurations can be applied directly on the application cluster as k8s CR. 

Following command will deploy `httpbin` app, create `WasmExtension` and installs Istio Ingress Gateway

```sh
cd tsb
kustomize build --reorder none | k apply -f -

wasmextension.extension.tsb.tetrate.io/wasm-add-header created
namespace/httpbin created
serviceaccount/httpbin created
service/httpbin created
deployment.apps/httpbin created
tenant.tsb.tetrate.io/httpbin created
workspace.tsb.tetrate.io/httpbin-ws created
namespace/tier2 created
tenant.tsb.tetrate.io/tier2 created
workspace.tsb.tetrate.io/tier2-ws created
group.gateway.tsb.tetrate.io/tier2-gg created
tier1gateway.install.tetrate.io/tier2-gateway created
gateway.gateway.tsb.tetrate.io/tier2-gateway created
```

## Test WASM Extensions

Retrieve Istio IngressGateway Loadbalancer service IP

```sh
export GATEWAY_IP=$(kubectl -n tier2 get service tier2-gateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

Request

```sh
curl -s "http://httpbin.tetrate.io/get" --resolve "httpbin.tetrate.io:80:${GATEWAY_IP}" -v --header "x-request-header: some value"
```

Response

```
* Added httpbin.tetrate.io:80:34.102.111.61 to DNS cache
* Hostname httpbin.tetrate.io was found in DNS cache
*   Trying 34.102.111.61:80...
* Connected to httpbin.tetrate.io (34.102.111.61) port 80 (#0)
> GET /get HTTP/1.1
> Host: httpbin.tetrate.io
> User-Agent: curl/7.79.1
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< server: istio-envoy
< date: Thu, 26 Oct 2023 14:48:53 GMT
< content-type: application/json
< content-length: 763
< access-control-allow-origin: *
< access-control-allow-credentials: true
< x-envoy-upstream-service-time: 16
< x-proxy-wasm-go-sdk-example: http_headers
< x-ingress-header: from tsb ingress gateway
<
{
  "args": {},
  "headers": {
    "Accept": "*/*",
    "Host": "httpbin.tetrate.io",
    "User-Agent": "curl/7.79.1",
    "X-B3-Parentspanid": "ad64b0102c00bf0c",
    "X-B3-Sampled": "0",
    "X-B3-Spanid": "b1f48c960d2b2735",
    "X-B3-Traceid": "b7592cad36ecc7aaad64b0102c00bf0c",
    "X-Envoy-Attempt-Count": "1",
    "X-Envoy-Internal": "true",
    "X-Forwarded-Client-Cert": "By=spiffe://gke-sreehari-us-west2-3.tsb.local/ns/httpbin/sa/httpbin;Hash=de8c9c7bbbbb957e5da0abd48acf998915bcec84f41cbd18f7d57e36147b6675;Subject=\"\";URI=spiffe://gke-sreehari-us-west2-3.tsb.local/ns/tier2/sa/tier2-gateway-service-account",
    "X-Request-Header": "changed/created by wasm"
  },
  "origin": "172.20.176.7",
  "url": "http://httpbin.tetrate.io/get"
}
```

As you can see, following headers has been added to the request/response

- `x-ingress-header` : `from tsb ingress gateway` added to the response header
- `X-Request-Header` : `changed/created by wasm` added to the request header, as you can see the value has been modified from its original value 


