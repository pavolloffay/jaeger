module github.com/jaegertracing/jaeger

go 1.13

require (
	github.com/AndreasBriese/bbloom v0.0.0-20190825152654-46b345b51c96 // indirect
	github.com/DataDog/zstd v1.4.4 // indirect
	github.com/Shopify/sarama v1.22.2-0.20190604114437-cd910a683f9f
	github.com/apache/thrift v0.0.0-20161221203622-b2a4d4ae21c7
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/bsm/sarama-cluster v2.1.13+incompatible
	github.com/crossdock/crossdock-go v0.0.0-20160816171116-049aabb0122b
	github.com/dgraph-io/badger v1.5.3
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/frankban/quicktest v1.7.3 // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-openapi/analysis v0.19.7 // indirect
	github.com/go-openapi/errors v0.19.3
	github.com/go-openapi/loads v0.19.4
	github.com/go-openapi/runtime v0.19.11
	github.com/go-openapi/spec v0.19.6
	github.com/go-openapi/strfmt v0.19.4
	github.com/go-openapi/swag v0.19.7
	github.com/go-openapi/validate v0.19.6
	github.com/gocql/gocql v0.0.0-20200226121155-e5c8c1f505c5
	github.com/gogo/googleapis v1.3.0
	github.com/gogo/protobuf v1.3.0
	github.com/golang/protobuf v1.3.3
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.13.0
	github.com/hashicorp/go-hclog v0.8.0
	github.com/hashicorp/go-plugin v1.0.1
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/yamux v0.0.0-20190923154419-df201c70410d // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/kr/pretty v0.2.0
	github.com/kr/text v0.2.0 // indirect
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/olivere/elastic v6.2.27+incompatible
	github.com/open-telemetry/opentelemetry-collector v0.2.7-0.20200226144913-d17176da0562
	github.com/opentracing-contrib/go-stdlib v0.0.0-20190519235532-cf7a6c988dc9
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pierrec/lz4 v2.4.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/common v0.9.1 // indirect
	github.com/prometheus/procfs v0.0.10 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/rs/cors v1.7.0
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.5.0
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	github.com/uber/tchannel-go v1.16.0
	go.mongodb.org/mongo-driver v1.3.0 // indirect
	go.uber.org/atomic v1.5.1
	go.uber.org/automaxprocs v1.3.0
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/zap v1.13.0
	golang.org/x/crypto v0.0.0-20200214034016-1d94cc7ab1c6 // indirect
	golang.org/x/lint v0.0.0-20200130185559-910be7a94367 // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/sys v0.0.0-20200217220822-9197077df867
	golang.org/x/tools v0.0.0-20200218205902-f8e42dc47720 // indirect
	google.golang.org/genproto v0.0.0-20200218151345-dad8c97a84f5 // indirect
	google.golang.org/grpc v1.27.1
	gopkg.in/ini.v1 v1.52.0 // indirect
	gopkg.in/jcmturner/goidentity.v3 v3.0.0 // indirect
	gopkg.in/jcmturner/gokrb5.v7 v7.5.0 // indirect
	gopkg.in/yaml.v2 v2.2.8
)

// taken from opentelemetry-collector. It imports k8s client
replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190620085101-78d2af792bab
