package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"apigov.dev/flame/cmd/flame/connection"
	"apigov.dev/flame/gapic"
	rpcpb "apigov.dev/flame/rpc"
	"github.com/spf13/cobra"
)

// specCmd represents the spec command
var specCmd = &cobra.Command{
	Use:   "spec",
	Short: "Upload files of an API spec.",
	Long:  "Upload files of an API spec.",
	Run: func(cmd *cobra.Command, args []string) {
		flagset := cmd.LocalFlags()
		spec, err := flagset.GetString("spec")
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		fmt.Printf("spec called with args %+v and spec %s\n", args, spec)

		client, err := connection.NewClient()
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		for _, arg := range args {
			matches, err := filepath.Glob(arg)
			if err != nil {
				log.Printf("%s\n", err.Error())
			}
			// for each match, upload the file
			for _, match := range matches {
				log.Printf("now upload %+v", match)
				fi, err := os.Stat(match)
				if err == nil {
					switch mode := fi.Mode(); {
					case mode.IsDir():
						fmt.Printf("upload directory %s\n", match)
						uploadDirectory(match, client, spec)
					case mode.IsRegular():
						fmt.Printf("upload file %s\n", match)
						uploadFile(match, client, spec)
					}
				} else {
					log.Printf("%+v", err)
				}
			}
		}
	},
}

func uploadDirectory(dirname string, client *gapic.FlameClient, spec string) error {
	return filepath.Walk(dirname,
		func(path string, info os.FileInfo, err error) error {
			log.Printf("%+s", path)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				uploadFile(path, client, spec)
			}
			return nil
		})
}

func uploadFile(filename string, client *gapic.FlameClient, spec string) {
	// does the file exist? if not, create it
	request := &rpcpb.GetFileRequest{}
	request.Name = spec + "/files/" + filename
	ctx := context.TODO()
	response, err := client.GetFile(ctx, request)
	log.Printf("response %+v\nerr %+v", response, err)
	if err != nil { // TODO only do this for NotFound errors
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("err %+v", err)
		} else {
			request := &rpcpb.CreateFileRequest{}
			request.Parent = spec
			request.FileId = filename
			request.File = &rpcpb.File{}
			request.File.Filename = filename
			request.File.Contents = bytes
			response, err := client.CreateFile(ctx, request)
			log.Printf("response %+v\nerr %+v", response, err)
		}
	}
}

func init() {
	uploadCmd.AddCommand(specCmd)
	specCmd.Flags().String("spec", "", "Resource name of spec to upload")
}
