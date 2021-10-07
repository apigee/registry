module github.com/apigee/registry

go 1.15

require (
	cloud.google.com/go/pubsub v1.17.0
	github.com/GoogleCloudPlatform/cloudsql-proxy v1.26.0
	github.com/apex/log v1.9.0
	github.com/blevesearch/bleve v1.0.14
	github.com/getkin/kin-openapi v0.77.0
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.2
	github.com/google/cel-go v0.8.0
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/googleapis/gax-go/v2 v2.1.1
	github.com/googleapis/gnostic v0.5.5
	github.com/nsf/jsondiff v0.0.0-20210926074059-1e845ec5d249
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tufin/oasdiff v1.0.6
	github.com/yoheimuta/go-protoparser/v4 v4.4.0
	golang.org/x/net v0.0.0-20211007125505-59d4e928ea9d
	golang.org/x/oauth2 v0.0.0-20211005180243-6b3c2da341f1
	google.golang.org/api v0.58.0
	google.golang.org/genproto v0.0.0-20211007155348-82e027067bd4
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gorm.io/driver/postgres v1.1.2
	gorm.io/driver/sqlite v1.1.5
	gorm.io/gorm v1.21.15
)

require (
	github.com/RoaringBitmap/roaring v0.9.4 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20210930093333-01de314d7883 // indirect
	github.com/bits-and-blooms/bitset v1.2.1 // indirect
	github.com/blevesearch/mmap-go v1.0.3 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/willf/bitset v1.2.1 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/mod v0.5.1 // indirect
	golang.org/x/sys v0.0.0-20211007075335-d3039528d8ac // indirect
	golang.org/x/tools v0.1.7 // indirect
)

replace (
	github.com/willf/bitset => github.com/bits-and-blooms/bitset v1.1.10
)
