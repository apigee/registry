// Code generated. DO NOT EDIT.

package main

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"
)

var CreateArtifactInput rpcpb.CreateArtifactRequest

var CreateArtifactFromFile string

func init() {
	RegistryServiceCmd.AddCommand(CreateArtifactCmd)

	CreateArtifactInput.Artifact = new(rpcpb.Artifact)

	CreateArtifactCmd.Flags().StringVar(&CreateArtifactInput.Parent, "parent", "", "Required. The parent, which owns this collection of...")

	CreateArtifactCmd.Flags().StringVar(&CreateArtifactInput.Artifact.Name, "artifact.name", "", "Resource name.")

	CreateArtifactCmd.Flags().StringVar(&CreateArtifactInput.Artifact.MimeType, "artifact.mime_type", "", "A content type specifier for the artifact. ...")

	CreateArtifactCmd.Flags().BytesHexVar(&CreateArtifactInput.Artifact.Contents, "artifact.contents", []byte{}, "The contents of the artifact.  Provided by API...")

	CreateArtifactCmd.Flags().StringVar(&CreateArtifactInput.ArtifactId, "artifact_id", "", "The ID to use for the artifact, which will become...")

	CreateArtifactCmd.Flags().StringVar(&CreateArtifactFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var CreateArtifactCmd = &cobra.Command{
	Use:   "create-artifact",
	Short: "CreateArtifact creates a specified artifact.",
	Long:  "CreateArtifact creates a specified artifact.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if CreateArtifactFromFile == "" {

			cmd.MarkFlagRequired("parent")

		}

	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {

		in := os.Stdin
		if CreateArtifactFromFile != "" {
			in, err = os.Open(CreateArtifactFromFile)
			if err != nil {
				return err
			}
			defer in.Close()

			err = jsonpb.Unmarshal(in, &CreateArtifactInput)
			if err != nil {
				return err
			}

		}

		if Verbose {
			printVerboseInput("Registry", "CreateArtifact", &CreateArtifactInput)
		}
		resp, err := RegistryClient.CreateArtifact(ctx, &CreateArtifactInput)

		if Verbose {
			fmt.Print("Output: ")
		}
		printMessage(resp)

		return err
	},
}
