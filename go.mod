module github.com/sikalabs/slr

go 1.24

toolchain go1.24.2

replace (
	k8s.io/api => k8s.io/api v0.32.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.32.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.32.1
	k8s.io/apiserver => k8s.io/apiserver v0.32.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.32.1
	k8s.io/client-go => k8s.io/client-go v0.32.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.32.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.32.1
	k8s.io/code-generator => k8s.io/code-generator v0.32.1
	k8s.io/component-base => k8s.io/component-base v0.32.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.32.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.32.1
	k8s.io/cri-api => k8s.io/cri-api v0.32.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.32.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.32.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.32.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.32.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.32.1
	k8s.io/kubectl => k8s.io/kubectl v0.32.1
	k8s.io/kubelet => k8s.io/kubelet v0.32.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.32.1
	k8s.io/metrics => k8s.io/metrics v0.32.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.32.1
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.32.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.32.1
)

require (
	github.com/Masterminds/semver/v3 v3.4.0
	github.com/coreos/go-oidc v2.3.0+incompatible
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/hashicorp/vault/api v1.20.0
	github.com/mlosinsky/clisso/ssoclient v1.0.0
	github.com/osteele/gojekyll v0.2.16
	github.com/prometheus/client_golang v1.23.0
	github.com/rwcarlsen/goexif v0.0.0-20190401172101-9e8deecbddbd
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/sikalabs/slu v0.85.0
	github.com/spf13/cobra v1.9.1
	golang.org/x/oauth2 v0.30.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/apimachinery v0.32.1
	k8s.io/client-go v11.0.1-0.20190816222228-6d55c1b1f1ca+incompatible
)

require (
	github.com/alecthomas/chroma v0.10.0 // indirect
	github.com/alecthomas/kingpin/v2 v2.4.0 // indirect
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/aybabtme/rgbterm v0.0.0-20170906152045-cc83f3b3ce59 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bep/godartsass/v2 v2.5.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/danog/blackfriday/v2 v2.1.6 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-github v17.0.0+incompatible // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.2.0 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.7 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-7 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jaschaephraim/lrserver v0.0.0-20240306232639-afed386b3640 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k0kubun/pp v3.0.1+incompatible // indirect
	github.com/kyokomi/emoji v2.2.4+incompatible // indirect
	github.com/lufia/plan9stats v0.0.0-20250317134145-8bc96cf8fc35 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/ondrejsika/go-dela v1.1.0 // indirect
	github.com/osteele/liquid v1.6.0 // indirect
	github.com/osteele/tuesday v1.0.3 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/pquerna/cachecontrol v0.2.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.65.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/qeesung/image2ascii v1.0.1 // indirect
	github.com/radovskyb/watcher v1.0.7 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/tdewolff/minify v2.3.6+incompatible // indirect
	github.com/tdewolff/parse v2.3.4+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/wayneashleyberry/terminal-dimensions v1.1.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.32.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241105132330-32ad38e42d3f // indirect
	k8s.io/utils v0.0.0-20241104100929-3ea5e8cea738 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.2 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
