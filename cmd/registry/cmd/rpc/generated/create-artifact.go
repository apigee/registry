// Code generated. DO NOT EDIT.

package generated

import (
	"github.com/spf13/cobra"

	"fmt"

	"github.com/golang/protobuf/jsonpb"

	"os"

	rpcpb "github.com/apigee/registry/rpc"

	"strings"
)

var CreateArtifactInput rpcpb.CreateArtifactRequest

var CreateArtifactFromFile string

var CreateArtifactInputArtifactLabels []string

var CreateArtifactInputArtifactAnnotations []string

func init() {
	RegistryServiceCmd.AddCommand(CreateArtifactCmd)

	CreateArtifactInput.Artifact = new(rpcpb.Artifact)

	CreateArtifactCmd.Flags().StringVar(&CreateArtifactInput.Parent, "parent", "", "Required. The parent, which owns this collection...")

	CreateArtifactCmd.Flags().StringVar(&CreateArtifactInput.Artifact.Name, "artifact.name", "", "Resource name.")

	CreateArtifactCmd.Flags().StringVar(&CreateArtifactInput.Artifact.MimeType, "artifact.mime_type", "", "A content type specifier for the artifact. ...")

	CreateArtifactCmd.Flags().BytesHexVar(&CreateArtifactInput.Artifact.Contents, "artifact.contents", []byte{}, "Input only. The contents of the artifact. ...")

	CreateArtifactCmd.Flags().StringArrayVar(&CreateArtifactInputArtifactLabels, "artifact.labels", []string{}, "key=value pairs. Labels attach identifying metadata to resources....")

	CreateArtifactCmd.Flags().StringArrayVar(&CreateArtifactInputArtifactAnnotations, "artifact.annotations", []string{}, "key=value pairs. Annotations attach non-identifying metadata to...")

	CreateArtifactCmd.Flags().StringVar(&CreateArtifactInput.ArtifactId, "artifact_id", "", "Required. The ID to use for the artifact, which...")

	CreateArtifactCmd.Flags().StringVar(&CreateArtifactFromFile, "from_file", "", "Absolute path to JSON file containing request payload")

}

var CreateArtifactCmd = &cobra.Command{
	Use:   "create-artifact",
	Short: "CreateArtifact creates a specified artifact.",
	Long:  "CreateArtifact creates a specified artifact.",
	PreRun: func(cmd *cobra.Command, args []string) {

		if CreateArtifactFromFile == "" {

			cmd.MarkFlagRequired("parent")

			cmd.MarkFlagRequired("artifact_id")

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

		if len(CreateArtifactInputArtifactLabels) > 0 {
			CreateArtifactInput.Artifact.Labels = make(map[string]string)
		}
		for _, item := range CreateArtifactInputArtifactLabels {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			CreateArtifactInput.Artifact.Labels[split[0]] = split[1]
		}

		if len(CreateArtifactInputArtifactAnnotations) > 0 {
			CreateArtifactInput.Artifact.Annotations = make(map[string]string)
		}
		for _, item := range CreateArtifactInputArtifactAnnotations {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			CreateArtifactInput.Artifact.Annotations[split[0]] = split[1]
		}

		if Verbose {
			printVerboseInput("Registry", "CreateArtifact", &CreateArtifactInput)
		}
		resp, err := RegistryClient.CreateArtifact(ctx, &CreateArtifactInput)
		if err != nil {
			return err
		}

		if Verbose {
			fmt.Print("Output: ")
		}
		printMessage(resp)

		return err
	},
}
