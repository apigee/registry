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

var ReplaceArtifactInput rpcpb.ReplaceArtifactRequest

var ReplaceArtifactFromFile string

var ReplaceArtifactInputArtifactLabels []string

var ReplaceArtifactInputArtifactAnnotations []string

func init() {
	RegistryServiceCmd.AddCommand(ReplaceArtifactCmd)

	ReplaceArtifactInput.Artifact = new(rpcpb.Artifact)

	ReplaceArtifactCmd.Flags().StringVar(&ReplaceArtifactInput.Artifact.Name, "artifact.name", "", "Resource name.")

	ReplaceArtifactCmd.Flags().StringVar(&ReplaceArtifactInput.Artifact.MimeType, "artifact.mime_type", "", "A content type specifier for the artifact. ...")

	ReplaceArtifactCmd.Flags().BytesHexVar(&ReplaceArtifactInput.Artifact.Contents, "artifact.contents", []byte{}, "Input only. The contents of the artifact. ...")

	ReplaceArtifactCmd.Flags().StringArrayVar(&ReplaceArtifactInputArtifactLabels, "artifact.labels", []string{}, "key=value pairs. Labels attach identifying metadata to resources....")

	ReplaceArtifactCmd.Flags().StringArrayVar(&ReplaceArtifactInputArtifactAnnotations, "artifact.annotations", []string{}, "key=value pairs. Annotations attach non-identifying metadata to...")

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

		if len(ReplaceArtifactInputArtifactLabels) > 0 {
			ReplaceArtifactInput.Artifact.Labels = make(map[string]string)
		}
		for _, item := range ReplaceArtifactInputArtifactLabels {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			ReplaceArtifactInput.Artifact.Labels[split[0]] = split[1]
		}

		if len(ReplaceArtifactInputArtifactAnnotations) > 0 {
			ReplaceArtifactInput.Artifact.Annotations = make(map[string]string)
		}
		for _, item := range ReplaceArtifactInputArtifactAnnotations {
			split := strings.Split(item, "=")
			if len(split) < 2 {
				err = fmt.Errorf("Invalid map item: %q", item)
				return
			}

			ReplaceArtifactInput.Artifact.Annotations[split[0]] = split[1]
		}

		if Verbose {
			printVerboseInput("Registry", "ReplaceArtifact", &ReplaceArtifactInput)
		}
		resp, err := RegistryClient.ReplaceArtifact(ctx, &ReplaceArtifactInput)
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
