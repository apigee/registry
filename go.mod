module apigov.dev/flame

go 1.13

require (
	cloud.google.com/go v0.38.0
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/golang/protobuf v1.3.5
	github.com/googleapis/gapic-generator-go v0.12.0 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/googleapis/gnostic v0.1.1-0.20200308034506-2af3d8e5d92a
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.6.2
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200320181252-af34d8274f85 // indirect
	google.golang.org/api v0.19.0
	google.golang.org/genproto v0.0.0-20200319113533-08878b785e9c
	google.golang.org/grpc v1.28.0
)

replace github.com/googleapis/gnostic v0.1.1-0.20200308034506-2af3d8e5d92a => github.com/timburks/gnostic v0.1.1-0.20200308034506-2af3d8e5d92a
