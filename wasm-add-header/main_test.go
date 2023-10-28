package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestHttpHeaders_OnHttpRequestHeaders(t *testing.T) {
	vmTest(t, func(t *testing.T, vm types.VMContext) {
		opt := proxytest.NewEmulatorOption().WithVMContext(vm)
		host, reset := proxytest.NewHostEmulator(opt)
		defer reset()

		t.Run("add a default request header", func(t *testing.T) {
			// Initialize http context.
			id := host.InitializeHttpContext()

			// Call OnHttpRequesteHeaders.
			hs := [][2]string{}
			action := host.CallOnRequestHeaders(id,
				hs, false)
			require.Equal(t, types.ActionContinue, action)

			resultHeaders := host.GetCurrentRequestHeaders(id)
			var found bool
			for _, val := range resultHeaders {
				if val[0] == "x-request-header" {
					require.Equal(t, "changed/created by wasm", val[1])
					found = true
				}
			}
			require.True(t, found)

			// Call OnHttpStreamDone.
			host.CompleteHttpContext(id)

			// Check Envoy logs.
			logs := host.GetInfoLogs()
			require.Contains(t, logs, fmt.Sprintf("%d finished", id))
			require.Contains(t, logs, "request header --> x-request-header: changed/created by wasm")
		})

		t.Run("add a new request header", func(t *testing.T) {
			// Initialize http context.
			id := host.InitializeHttpContext()

			// Call OnHttpRequesteHeaders.
			hs := [][2]string{{"key", "value"}}
			action := host.CallOnRequestHeaders(id,
				hs, false)
			require.Equal(t, types.ActionContinue, action)

			resultHeaders := host.GetCurrentRequestHeaders(id)
			var found bool
			for _, val := range resultHeaders {
				if val[0] == "key" {
					require.Equal(t, "value", val[1])
					found = true
				}
			}
			require.True(t, found)

			// Call OnHttpStreamDone.
			host.CompleteHttpContext(id)

			// Check Envoy logs.
			logs := host.GetInfoLogs()
			require.Contains(t, logs, fmt.Sprintf("%d finished", id))
			require.Contains(t, logs, "request header --> key: value")
			require.Contains(t, logs, "request header --> x-request-header: changed/created by wasm")
		})

		t.Run("modify the request header", func(t *testing.T) {
			// Initialize http context.
			id := host.InitializeHttpContext()

			// Call OnHttpRequesteHeaders.
			hs := [][2]string{{"x-request-header", "test"}}
			action := host.CallOnRequestHeaders(id,
				hs, false)
			require.Equal(t, types.ActionContinue, action)

			// Check headers.
			resultHeaders := host.GetCurrentRequestHeaders(id)
			var found bool
			for _, val := range resultHeaders {
				if val[0] == "x-request-header" {
					require.Equal(t, "changed/created by wasm", val[1])
					found = true
				}
			}
			require.True(t, found)

			// Call OnHttpStreamDone.
			host.CompleteHttpContext(id)

			// Check Envoy logs.
			logs := host.GetInfoLogs()
			require.Contains(t, logs, fmt.Sprintf("%d finished", id))
			require.Contains(t, logs, "request header --> x-request-header: changed/created by wasm")
		})
	})
}

func TestHttpHeaders_OnHttpRequestHeadersWithOnPluginStart(t *testing.T) {
	vmTest(t, func(t *testing.T, vm types.VMContext) {
		proxyConfig, err := proxyConfiguration("x-user-id", "123", "request")
		if err != nil {
			require.NoError(t, err)
		}

		opt := proxytest.NewEmulatorOption().
			WithVMContext(vm).
			WithPluginConfiguration(proxyConfig)

		host, reset := proxytest.NewHostEmulator(opt)
		defer reset()

		t.Run("add a new request header based on plugin configuration", func(t *testing.T) {
			// call start Plugin to configure plugin configuration in context
			require.Equal(t, host.StartPlugin(), types.OnPluginStartStatusOK)

			// Initialize http context.
			id := host.InitializeHttpContext()

			hs := [][2]string{}
			action := host.CallOnRequestHeaders(id,
				hs, false)
			require.Equal(t, types.ActionContinue, action)

			// Check headers.
			resultHeaders := host.GetCurrentRequestHeaders(id)
			var found bool
			for _, val := range resultHeaders {
				if val[0] == "x-user-id" {
					require.Equal(t, "123", val[1])
					found = true
				}
			}
			require.True(t, found)

			// Call OnHttpStreamDone.
			host.CompleteHttpContext(id)

			// Check Envoy logs.
			logs := host.GetInfoLogs()
			require.Contains(t, logs, fmt.Sprintf("%d finished", id))
			require.Contains(t, logs, "request header --> x-request-header: changed/created by wasm")
			require.Contains(t, logs, "request header --> x-user-id: 123")
		})
	})
}

func TestHttpHeaders_OnHttpResponseHeaders(t *testing.T) {
	vmTest(t, func(t *testing.T, vm types.VMContext) {
		opt := proxytest.NewEmulatorOption().WithVMContext(vm)
		host, reset := proxytest.NewHostEmulator(opt)
		defer reset()

		t.Run("add a default response header", func(t *testing.T) {
			// Initialize http context.
			id := host.InitializeHttpContext()

			// Call OnHttpRequesteHeaders.
			hs := [][2]string{}
			action := host.CallOnResponseHeaders(id,
				hs, false)
			require.Equal(t, types.ActionContinue, action)

			resultHeaders := host.GetCurrentResponseHeaders(id)
			var found bool
			for _, val := range resultHeaders {
				if val[0] == "x-proxy-wasm-go-sdk-example" {
					require.Equal(t, "http_headers", val[1])
					found = true
				}
			}
			require.True(t, found)

			// Call OnHttpStreamDone.
			host.CompleteHttpContext(id)

			// Check Envoy logs.
			logs := host.GetInfoLogs()
			require.Contains(t, logs, fmt.Sprintf("%d finished", id))
			require.Contains(t, logs, "response header <-- x-proxy-wasm-go-sdk-example: http_headers")
		})

		t.Run("add a new response header", func(t *testing.T) {
			// Initialize http context.
			id := host.InitializeHttpContext()

			// Call OnHttpRequesteHeaders.
			hs := [][2]string{{"x-user-id", "123"}}
			action := host.CallOnResponseHeaders(id,
				hs, false)
			require.Equal(t, types.ActionContinue, action)

			// Check headers.
			resultHeaders := host.GetCurrentResponseHeaders(id)
			var found bool
			for _, val := range resultHeaders {
				if val[0] == "x-user-id" {
					require.Equal(t, "123", val[1])
					found = true
				}
			}
			require.True(t, found)

			// Call OnHttpStreamDone.
			host.CompleteHttpContext(id)

			// Check Envoy logs.
			logs := host.GetInfoLogs()
			require.Contains(t, logs, fmt.Sprintf("%d finished", id))
			require.Contains(t, logs, "response header <-- x-proxy-wasm-go-sdk-example: http_headers")
			require.Contains(t, logs, "response header <-- x-user-id: 123")
		})
	})
}

func TestHttpHeaders_OnHttpResponseHeadersWithOnPluginStart(t *testing.T) {
	vmTest(t, func(t *testing.T, vm types.VMContext) {
		proxyConfig, err := proxyConfiguration("x-ingress-header", "powered by TSB", "response")
		if err != nil {
			require.NoError(t, err)
		}

		opt := proxytest.NewEmulatorOption().
			WithVMContext(vm).
			WithPluginConfiguration(proxyConfig)

		host, reset := proxytest.NewHostEmulator(opt)
		defer reset()

		t.Run("add a new response header based on plugin configuration", func(t *testing.T) {
			// call start Plugin to configure plugin configuration in context
			require.Equal(t, host.StartPlugin(), types.OnPluginStartStatusOK)

			// Initialize http context.
			id := host.InitializeHttpContext()

			hs := [][2]string{}
			action := host.CallOnResponseHeaders(id,
				hs, false)
			require.Equal(t, types.ActionContinue, action)

			// Check headers.
			resultHeaders := host.GetCurrentResponseHeaders(id)
			var found bool
			for _, val := range resultHeaders {
				if val[0] == "x-ingress-header" {
					require.Equal(t, "powered by TSB", val[1])
					found = true
				}
			}
			require.True(t, found)

			// Call OnHttpStreamDone.
			host.CompleteHttpContext(id)

			// Check Envoy logs.
			logs := host.GetInfoLogs()
			require.Contains(t, logs, fmt.Sprintf("%d finished", id))
			require.Contains(t, logs, "response header <-- x-proxy-wasm-go-sdk-example: http_headers")
			require.Contains(t, logs, "response header <-- x-ingress-header: powered by TSB")
		})
	})
}

// vmTest executes f twice, once with a types.VMContext that executes plugin code directly
// in the host, and again by executing the plugin code within the compiled main.wasm binary.
// Execution with main.wasm will be skipped if the file cannot be found.
func vmTest(t *testing.T, f func(*testing.T, types.VMContext)) {
	t.Helper()

	t.Run("go", func(t *testing.T) {
		f(t, &vmContext{})
	})

	t.Run("wasm", func(t *testing.T) {
		wasm, err := os.ReadFile("main.wasm")
		if err != nil {
			t.Skip("wasm not found")
		}
		v, err := proxytest.NewWasmVMContext(wasm)
		require.NoError(t, err)
		defer v.Close()
		f(t, v)
	})
}

func proxyConfiguration(header string, value string, path string) ([]byte, error) {
	rawJson, err := json.Marshal(&struct {
		Header string `json:"header"`
		Value  string `json:"value"`
		Path   string `json:"path"`
	}{
		Header: header,
		Value:  value,
		Path:   path,
	})

	if err != nil {
		return nil, err
	}

	return rawJson, nil
}
