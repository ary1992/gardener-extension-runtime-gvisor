// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: networking/v1alpha3/envoy_filter.proto

// `EnvoyFilter` provides a mechanism to customize the Envoy
// configuration generated by Istio Pilot. Use EnvoyFilter to modify
// values for certain fields, add specific filters, or even add
// entirely new listeners, clusters, etc. This feature must be used
// with care, as incorrect configurations could potentially
// destabilize the entire mesh. Unlike other Istio networking objects,
// EnvoyFilters are additively applied. Any number of EnvoyFilters can
// exist for a given workload in a specific namespace. The order of
// application of these EnvoyFilters is as follows: all EnvoyFilters
// in the config [root
// namespace](https://istio.io/docs/reference/config/istio.mesh.v1alpha1/#MeshConfig),
// followed by all matching EnvoyFilters in the workload's namespace.
//
// **NOTE 1**: Some aspects of this API are deeply tied to the internal
// implementation in Istio networking subsystem as well as Envoy's XDS
// API. While the EnvoyFilter API by itself will maintain backward
// compatibility, any envoy configuration provided through this
// mechanism should be carefully monitored across Istio proxy version
// upgrades, to ensure that deprecated fields are removed and replaced
// appropriately.
//
// **NOTE 2**: When multiple EnvoyFilters are bound to the same
// workload in a given namespace, all patches will be processed
// sequentially in order of creation time.  The behavior is undefined
// if multiple EnvoyFilter configurations conflict with each other.
//
// **NOTE 3**: To apply an EnvoyFilter resource to all workloads
// (sidecars and gateways) in the system, define the resource in the
// config [root
// namespace](https://istio.io/docs/reference/config/istio.mesh.v1alpha1/#MeshConfig),
// without a workloadSelector.
//
// The example below declares a global default EnvoyFilter resource in
// the root namespace called `istio-config`, that adds a custom
// protocol filter on all sidecars in the system, for outbound port
// 9307. The filter should be added before the terminating tcp_proxy
// filter to take effect. In addition, it sets a 30s idle timeout for
// all HTTP connections in both gateways and sidecars.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: EnvoyFilter
// metadata:
//   name: custom-protocol
//   namespace: istio-config # as defined in meshConfig resource.
// spec:
//   configPatches:
//   - applyTo: NETWORK_FILTER
//     match:
//       context: SIDECAR_OUTBOUND # will match outbound listeners in all sidecars
//       listener:
//         portNumber: 9307
//         filterChain:
//           filter:
//             name: "envoy.filters.network.tcp_proxy"
//     patch:
//       operation: INSERT_BEFORE
//       value:
//         # This is the full filter config including the name and typed_config section.
//         name: "envoy.config.filter.network.custom_protocol"
//         typed_config:
//          ...
//   - applyTo: NETWORK_FILTER # http connection manager is a filter in Envoy
//     match:
//       # context omitted so that this applies to both sidecars and gateways
//       listener:
//         filterChain:
//           filter:
//             name: "envoy.filters.network.http_connection_manager"
//     patch:
//       operation: MERGE
//       value:
//         name: "envoy.filters.network.http_connection_manager"
//         typed_config:
//           "@type": "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager"
//           common_http_protocol_options:
//             idle_timeout: 30s
// ```
//
// The following example enables Envoy's Lua filter for all inbound
// HTTP calls arriving at service port 8080 of the reviews service pod
// with labels "app: reviews", in the bookinfo namespace. The lua
// filter calls out to an external service internal.org.net:8888 that
// requires a special cluster definition in envoy. The cluster is also
// added to the sidecar as part of this configuration.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: EnvoyFilter
// metadata:
//   name: reviews-lua
//   namespace: bookinfo
// spec:
//   workloadSelector:
//     labels:
//       app: reviews
//   configPatches:
//     # The first patch adds the lua filter to the listener/http connection manager
//   - applyTo: HTTP_FILTER
//     match:
//       context: SIDECAR_INBOUND
//       listener:
//         portNumber: 8080
//         filterChain:
//           filter:
//             name: "envoy.filters.network.http_connection_manager"
//             subFilter:
//               name: "envoy.filters.http.router"
//     patch:
//       operation: INSERT_BEFORE
//       value: # lua filter specification
//        name: envoy.lua
//        typed_config:
//           "@type": "type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua"
//           inlineCode: |
//             function envoy_on_request(request_handle)
//               -- Make an HTTP call to an upstream host with the following headers, body, and timeout.
//               local headers, body = request_handle:httpCall(
//                "lua_cluster",
//                {
//                 [":method"] = "POST",
//                 [":path"] = "/acl",
//                 [":authority"] = "internal.org.net"
//                },
//               "authorize call",
//               5000)
//             end
//   # The second patch adds the cluster that is referenced by the lua code
//   # cds match is omitted as a new cluster is being added
//   - applyTo: CLUSTER
//     match:
//       context: SIDECAR_OUTBOUND
//     patch:
//       operation: ADD
//       value: # cluster specification
//         name: "lua_cluster"
//         type: STRICT_DNS
//         connect_timeout: 0.5s
//         lb_policy: ROUND_ROBIN
//         load_assignment:
//           cluster_name: lua_cluster
//           endpoints:
//           - lb_endpoints:
//             - endpoint:
//                 address:
//                   socket_address:
//                     protocol: TCP
//                     address: "internal.org.net"
//                     port_value: 8888
// ```
//
// The following example overwrites certain fields (HTTP idle timeout
// and X-Forward-For trusted hops) in the HTTP connection manager in a
// listener on the ingress gateway in istio-system namespace for the
// SNI host app.example.com:
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: EnvoyFilter
// metadata:
//   name: hcm-tweaks
//   namespace: istio-system
// spec:
//   workloadSelector:
//     labels:
//       istio: ingressgateway
//   configPatches:
//   - applyTo: NETWORK_FILTER # http connection manager is a filter in Envoy
//     match:
//       context: GATEWAY
//       listener:
//         filterChain:
//           sni: app.example.com
//           filter:
//             name: "envoy.filters.network.http_connection_manager"
//     patch:
//       operation: MERGE
//       value:
//         typed_config:
//           "@type": "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager"
//           xff_num_trusted_hops: 5
//           common_http_protocol_options:
//             idle_timeout: 30s
// ```
//
// The following example inserts an attributegen filter
// that produces `istio_operationId` attribute which is consumed
// by the istio.stats fiter. `filterClass: STATS` encodes this dependency.
//
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: EnvoyFilter
// metadata:
//   name: reviews-request-operation
//   namespace: myns
// spec:
//   workloadSelector:
//     labels:
//       app: reviews
//   configPatches:
//   - applyTo: HTTP_FILTER
//     match:
//       context: SIDECAR_INBOUND
//     patch:
//       operation: ADD
//       filterClass: STATS # This filter will run *before* the Istio stats filter.
//       value:
//         name: istio.request_operation
//         typed_config:
//          "@type": type.googleapis.com/udpa.type.v1.TypedStruct
//          type_url: type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
//          value:
//            config:
//              configuration: |
//                {
//                  "attributes": [
//                    {
//                      "output_attribute": "istio_operationId",
//                      "match": [
//                        {
//                          "value": "ListReviews",
//                          "condition": "request.url_path == '/reviews' && request.method == 'GET'"
//                        }]
//                    }]
//                }
//              vm_config:
//                runtime: envoy.wasm.runtime.null
//                code:
//                  local: { inline_string: "envoy.wasm.attributegen" }
// ```
//
// The following example inserts an http ext_authz filter in the `myns` namespace.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: EnvoyFilter
// metadata:
//   name: myns-ext-authz
//   namespace: myns
// spec:
//   configPatches:
//   - applyTo: HTTP_FILTER
//     match:
//       context: SIDECAR_INBOUND
//     patch:
//       operation: ADD
//       filterClass: AUTHZ # This filter will run *after* the Istio authz filter.
//       value:
//         name: envoy.filters.http.ext_authz
//         typed_config:
//           "@type": type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz
//           grpc_service:
//             envoy_grpc:
//               cluster_name: acme-ext-authz
//             initial_metadata:
//             - key: foo
//               value: myauth.acme # required by local ext auth server.
// ```
//
// A workload in the `myns` namespace needs to access a different ext_auth server
// that does not accept initial metadata. Since proto merge cannot remove fields, the
// following configuration uses the `REPLACE` operation. If you do not need to inherit
// fields, REPLACE is preferred over MERGE.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: EnvoyFilter
// metadata:
//   name: mysvc-ext-authz
//   namespace: myns
// spec:
//   workloadSelector:
//     labels:
//       app: mysvc
//   configPatches:
//   - applyTo: HTTP_FILTER
//     match:
//       context: SIDECAR_INBOUND
//     patch:
//       operation: REPLACE
//       value:
//         name: envoy.filters.http.ext_authz
//         typed_config:
//           "@type": type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz
//           grpc_service:
//             envoy_grpc:
//               cluster_name: acme-ext-authz-alt
// ```
//
// The following example deploys a Wasm extension for all inbound sidecar HTTP requests.
//
// ```yaml
// apiVersion: networking.istio.io/v1alpha3
// kind: EnvoyFilter
// metadata:
//   name: wasm-example
//   namespace: myns
// spec:
//   configPatches:
//   # The first patch defines a named Wasm extension and provides a URL to fetch Wasm binary from,
//   # and the binary configuration. It should come before the next patch that applies it.
//   # This resource is visible to all proxies in the namespace "myns". It is possible to provide
//   # multiple definitions for the same name "my-wasm-extension" in multiple namespaces. We recommend that:
//   # - if overriding is desired, then the root level definition can be overriden per namespace with REPLACE.
//   # - if overriding is not desired, then the name should be qualified with the namespace "myns/my-wasm-extension",
//   #   to avoid accidental name collisions.
//   - applyTo: EXTENSION_CONFIG
//     patch:
//       operation: ADD # REPLACE is also supported, and would override a cluster level resource with the same name.
//       value:
//         name: my-wasm-extension
//         typed_config:
//           "@type": type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
//           config:
//             root_id: my-wasm-root-id
//             vm_config:
//               vm_id: my-wasm-vm-id
//               runtime: envoy.wasm.runtime.v8
//               code:
//                 remote:
//                   http_uri:
//                     uri: http://my-wasm-binary-uri
//             configuration:
//               "@type": "type.googleapis.com/google.protobuf.StringValue"
//               value: |
//                 {}
//   # The second patch instructs to apply the above Wasm filter to the listener/http connection manager.
//   - applyTo: HTTP_FILTER
//     match:
//       context: SIDECAR_INBOUND
//     patch:
//       operation: ADD
//       filterClass: AUTHZ # This filter will run *after* the Istio authz filter.
//       value:
//         name: my-wasm-extension # This must match the name above
//         config_discovery:
//           config_source:
//             api_config_source:
//               api_type: GRPC
//               transport_api_version: V3
//               grpc_services:
//               - envoy_grpc:
//                   cluster_name: xds-grpc
//           type_urls: ["envoy.extensions.filters.http.wasm.v3.Wasm"]
// ```

package v1alpha3

import (
	bytes "bytes"
	fmt "fmt"
	github_com_gogo_protobuf_jsonpb "github.com/gogo/protobuf/jsonpb"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
	_ "istio.io/gogo-genproto/googleapis/google/api"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// MarshalJSON is a custom marshaler for EnvoyFilter
func (this *EnvoyFilter) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter
func (this *EnvoyFilter) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_ProxyMatch
func (this *EnvoyFilter_ProxyMatch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_ProxyMatch
func (this *EnvoyFilter_ProxyMatch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_ClusterMatch
func (this *EnvoyFilter_ClusterMatch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_ClusterMatch
func (this *EnvoyFilter_ClusterMatch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_RouteConfigurationMatch
func (this *EnvoyFilter_RouteConfigurationMatch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_RouteConfigurationMatch
func (this *EnvoyFilter_RouteConfigurationMatch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_RouteConfigurationMatch_RouteMatch
func (this *EnvoyFilter_RouteConfigurationMatch_RouteMatch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_RouteConfigurationMatch_RouteMatch
func (this *EnvoyFilter_RouteConfigurationMatch_RouteMatch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_RouteConfigurationMatch_VirtualHostMatch
func (this *EnvoyFilter_RouteConfigurationMatch_VirtualHostMatch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_RouteConfigurationMatch_VirtualHostMatch
func (this *EnvoyFilter_RouteConfigurationMatch_VirtualHostMatch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_ListenerMatch
func (this *EnvoyFilter_ListenerMatch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_ListenerMatch
func (this *EnvoyFilter_ListenerMatch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_ListenerMatch_FilterChainMatch
func (this *EnvoyFilter_ListenerMatch_FilterChainMatch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_ListenerMatch_FilterChainMatch
func (this *EnvoyFilter_ListenerMatch_FilterChainMatch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_ListenerMatch_FilterMatch
func (this *EnvoyFilter_ListenerMatch_FilterMatch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_ListenerMatch_FilterMatch
func (this *EnvoyFilter_ListenerMatch_FilterMatch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_ListenerMatch_SubFilterMatch
func (this *EnvoyFilter_ListenerMatch_SubFilterMatch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_ListenerMatch_SubFilterMatch
func (this *EnvoyFilter_ListenerMatch_SubFilterMatch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_Patch
func (this *EnvoyFilter_Patch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_Patch
func (this *EnvoyFilter_Patch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_EnvoyConfigObjectMatch
func (this *EnvoyFilter_EnvoyConfigObjectMatch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_EnvoyConfigObjectMatch
func (this *EnvoyFilter_EnvoyConfigObjectMatch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EnvoyFilter_EnvoyConfigObjectPatch
func (this *EnvoyFilter_EnvoyConfigObjectPatch) MarshalJSON() ([]byte, error) {
	str, err := EnvoyFilterMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EnvoyFilter_EnvoyConfigObjectPatch
func (this *EnvoyFilter_EnvoyConfigObjectPatch) UnmarshalJSON(b []byte) error {
	return EnvoyFilterUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

var (
	EnvoyFilterMarshaler   = &github_com_gogo_protobuf_jsonpb.Marshaler{}
	EnvoyFilterUnmarshaler = &github_com_gogo_protobuf_jsonpb.Unmarshaler{AllowUnknownFields: true}
)