module apigov.dev/flame

go 1.13

require (
	cloud.google.com/go v0.38.0
	github.com/docopt/docopt-go v0.0.0-20180111231733-ee0de3bc6815
	github.com/golang/protobuf v1.3.4
	github.com/googleapis/gax-go/v2 v2.0.5
	github.com/googleapis/gnostic v0.1.1-0.20200308034506-2af3d8e5d92a
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.6.2
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	google.golang.org/api v0.19.0
	google.golang.org/genproto v0.0.0-20200310143817-43be25429f5a
	google.golang.org/grpc v1.27.1
)

replace github.com/googleapis/gnostic v0.1.1-0.20200308034506-2af3d8e5d92a => github.com/timburks/gnostic v0.1.1-0.20200308034506-2af3d8e5d92a
