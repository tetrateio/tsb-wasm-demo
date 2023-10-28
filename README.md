# WASM Extensions in TSB(Istio)

This project demonstrate how WASM extensions can be created and integrated into an Istio Gateway using TSB as a service mesh.

## What is a WASM extension?

A WASM extension is a software addon of [WebAssembly](https://istio.io/latest/docs/concepts/wasm) which can be used to extend the Istio proxy (Envoy). You can read more about it from Tetrate docs [here](https://docs.tetrate.io/service-bridge/howto/wasm/wasm-overview)

To write WASM extensions, you can use Envoy WASM SDK that can compile the code into Envoy WASM compatible binary. There are [serveral SDKs available](https://github.com/proxy-wasm/spec). [TinyGo](https://github.com/tetratelabs/proxy-wasm-go-sdk) and [Rust](https://github.com/proxy-wasm/proxy-wasm-rust-sdk) are the most widely used among all. You can also refer [proxy-wasm-go-sdk/examples](https://github.com/tetratelabs/proxy-wasm-go-sdk/tree/main/examples) from Tetrate to get started with writing WASM extensions.

## Write a WASM Extension to modify Request/Response Header

[wasm-add-header](./wasm-add-header/) is a Go WASM extension example that will add a specific header to the request and response. 

Writing a WASM extension involves adding a custom logic at one or more callback functions that defined in [Envoy WASM ABI Spec](https://github.com/proxy-wasm/spec). 

In this example, we have used [TinyGo](https://github.com/tetratelabs/proxy-wasm-go-sdk) to add or modify request/response header. To do this, you can add your custom logic to `OnHttpRequestHeaders` or `OnHttpResponseHeaders`. See [wasm-add-header/main.go](./wasm-add-header/main.go)

Modify request header

```golang
func (ctx *httpHeaders) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	proxywasm.LogInfof("adding header to request: %s=%s", ctx.headerName, ctx.headerValue)
	if err := proxywasm.AddHttpRequestHeader(ctx.headerName, ctx.headerValue); err != nil {
		proxywasm.LogCriticalf("failed to set request headers: %v", err)
	}
	return types.ActionContinue
}
```

Modify response header

```golang
func (ctx *httpHeaders) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action { 
    proxywasm.LogInfof("adding header to response: %s=%s", ctx.headerName, ctx.headerValue)
    if err := proxywasm.AddHttpResponseHeader(ctx.headerName, ctx.headerValue); err != nil {
		proxywasm.LogCriticalf("failed to set response headers: %v", err)
	}
    return types.ActionContinue
}
```

Similarly you can find other supported handlers for modifying request/response headers, request/response body, dispatching a HTTP call to a remote cluster etc [here](https://github.com/tetratelabs/proxy-wasm-go-sdk/blob/main/proxywasm/hostcall.go)


### Build the WASM Extension

**Pre-req:** Make sure you have [TinyGo Installed](https://tinygo.org/getting-started/install/) in your env

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

## How to use WASM Extension in Istio Ingress Gateway

Since we will be using [Tetrate Service Bridge](https://docs.tetrate.io/service-bridge/next) service mesh to create Istio IngressGateway, you need to configure the following TSB resources to integrate WASM extensions within TSB.

### Create WasmExtension CR

[WasmExtension](./tsb/wasm/wasm-extension.yaml) configures the CR which can then be used inside a TSB Istio IngressGateway resource. 

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


### How does it work?

When you enable WASM proxy in the ManagementPlane, `wasmfetcher` component running in the `istio-system` namespace in the Controlplane cluster will fetch WASM image from the private registry and then Gateway will fetch WASM image from the `wasmfetcher` running as an http server. 

When you observe the logs, you will notice this

```
kubectl logs -l app=wasmfetcher -n istio-system -f

2023/10/12 07:39:34  info	http	starting HTTP server at: ":8080" [scope="http"]
2023/10/27 07:05:05  info	cache	fetching image sreeharikmarar/wasm-add-header from registry index.docker.io with tag latest [scope="cache"]
```

### Istio Configuration

Once TSB translate `WasmExtension` configured as part of `Gateway` into Istio's `WasmPlugin` definition, the image url will be updated to point to `wasmfetcher` url

```yaml
kubectl get wasmplugins.extensions.istio.io tier2-gateway-wasm-add-header0 -n tier2 -o yaml
```

```yaml
apiVersion: extensions.istio.io/v1alpha1
kind: WasmPlugin
metadata:
  annotations:
    tsb.tetrate.io/config-mode: bridged
    tsb.tetrate.io/etag: '"9m3TwL8TafI="'
    tsb.tetrate.io/fqn: organizations/tetrate/tenants/tier2/workspaces/tier2-ws/gatewaygroups/tier2-gg/unifiedgateways/tier2-gateway
    tsb.tetrate.io/runtime-etag: '"ot2xM0wdmKk="'
    xcp.tetrate.io/contentHash: 0001eb75bcd4685e528750c5d47706ff
  creationTimestamp: "2023-10-27T07:05:04Z"
  generation: 1
  labels:
    app.kubernetes.io/managed-by: tsb
    istio.io/rev: default
    xcp.tetrate.io/gatewayGroup: tier2-gg
    xcp.tetrate.io/workspace: tier2-ws-54ab19ffac3a761b
  name: tier2-gateway-wasm-add-header0
  namespace: tier2
  resourceVersion: "16421353"
  uid: e7ad9b18-3b31-4a22-b406-b042193096ea
spec:
  pluginConfig:
    header: x-ingress-header
    path: response
    value: from tsb ingress gateway
  pluginName: tier2-gateway-wasm-add-header0
  priority: 0
  selector:
    matchLabels:
      app: tier2-gateway
  url: http://wasmfetcher.istio-system.svc.cluster.local/fetch/b2NpOi8vZG9ja2VyLmlvL3NyZWVoYXJpa21hcmFyL3dhc20tYWRkLWhlYWRlcjpsYXRlc3Q=
```

### Envoy Configuration

Retrieve Envoy Config dump

```sh
istioctl dashboard envoy tier2-gateway-864588946d-6mhsn.tier2
```

Envoy Listener and HTTP Filter Config

```json
"dynamic_listeners": [
    {
     "name": "0.0.0.0_15443",
     "active_state": {
      "version_info": "2023-10-26T11:16:01Z/60",
      "listener": {
        ...
        ...
       },
    "filter_chains": [
      {
         "filter_chain_match": {
          "server_names": [
           "httpbin.tetrate.io"
          ]
         },
         "filters": [
          {
           "name": "envoy.filters.network.http_connection_manager",
            ...
            ...
            },
            "http_filters": [
            ...
            ...
            {
              "name": "tier2.tier2-gateway-wasm-add-header0",
              "config_discovery": {
               "config_source": {
                "ads": {},
                "initial_fetch_timeout": "0s",
                "resource_api_version": "V3"
               },
               "type_urls": [
                "type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm",
                "type.googleapis.com/envoy.extensions.filters.http.rbac.v3.RBAC"
               ]
              }
            }
            ...
      }  
```

Envoy Extension Config - Ecds Filter and [Wasm Filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/wasm_filter)



```json
  {
   "@type": "type.googleapis.com/envoy.admin.v3.EcdsConfigDump",
   "ecds_filters": [
    {
     "version_info": "2023-10-27T07:07:51Z/4825",
     "ecds_filter": {
      "@type": "type.googleapis.com/envoy.config.core.v3.TypedExtensionConfig",
      "name": "tier2.tier2-gateway-wasm-add-header0",
      "typed_config": {
       "@type": "type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm",
       "config": {
        "name": "tier2.tier2-gateway-wasm-add-header0",
        "root_id": "tier2-gateway-wasm-add-header0",
        "vm_config": {
         "runtime": "envoy.wasm.runtime.v8",
         "code": {
          "local": {
           "filename": "/var/lib/istio/data/2879432666c69c71230718d9eb2d594192aab5737d81de80d28feb705eb7f2b3/ba0fb8931399e8d0ad77756e8cc92de941e4667f5ff19405c045f32ab5c5efa9.wasm"
          }
         },
         "environment_variables": {
          "key_values": {
           "ISTIO_META_WASM_PLUGIN_RESOURCE_VERSION": "16399800"
          }
         }
        },
        "configuration": {
         "@type": "type.googleapis.com/google.protobuf.StringValue",
         "value": "{\"header\":\"x-ingress-header\",\"path\":\"response\",\"value\":\"from tsb ingress gateway\"}"
        }
       }
      }
     },
     "last_updated": "2023-10-27T07:07:51.546Z"
    }
   ]
  }
```