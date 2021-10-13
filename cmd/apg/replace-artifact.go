// Code generated. DO NOT EDIT.

package main

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var ReplaceArtifactInput rpcpb.ReplaceArtifactRequest

var ReplaceArtifactFromFile string

func init() {
	RegistryServiceCmd.AddCommand(ReplaceArtifactCmd)

	ReplaceArtifactInput.Artifact = new(rpcpb.Artifact)

	ReplaceArtifactCmd.Flags().StringVar(&ReplaceArtifactInput.Artifact.Name, "artifact.name", "", "Resource name.")

	ReplaceArtifactCmd.Flags().StringVar(&ReplaceArtifactInput.Artifact.MimeType, "artifact.mime_type", "", "A content type specifier for the artifact. ...")

	ReplaceArtifactCmd.Flags().BytesHexVar(&ReplaceArtifactInput.Artifact.Contents, "artifact.contents", []byte{}, "The contents of the artifact.  Provided by API...")

	ReplaceArtifactCmd.Flags().StringVar(&ReplaceArtifactFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var ReplaceArtifactCmd = &cobra.Command{
	Use:   "replace-artifact",
	Short: "ReplaceArtifact can be used to replace a...",
	Long:  "ReplaceArtifact can be used to replace a specified artifact.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if ReplaceArtifactFromFile == "" {

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if ReplaceArtifactFromFile != "" {
			in, err = os.Open(ReplaceArtifactFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &ReplaceArtifactInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "ReplaceArtifact", &ReplaceArtifactInput)
		}
		resp, err := RegistryClient.ReplaceArtifact(ctx, &ReplaceArtifactInput)

		if Verbose {
			fmt.Print("Output: ")
		}
		printMessage(resp)

		return err
	},
}
