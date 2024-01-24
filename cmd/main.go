package main

import (
	"crd-to-openapi/pkg"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	flags := pflag.NewFlagSet("crd-to-openapi", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := NewRootCommand(flags)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCommand(flags *pflag.FlagSet) *cobra.Command {
	var crdVersion string
	var openapi2 bool

	flags.StringVar(&crdVersion, "crd-version", "", "set crd version")
	flags.BoolVar(&openapi2, "output-openapi-v2", false, "output in OpenAPI v2 format, the default format is OpenAPI v3")

	cmd := &cobra.Command{
		Short: "CRD To OpenAPI",
		Long:  "Read CRD from stdin and write OpenAPI spec to stdout.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg.Convert(crdVersion, openapi2)
		},
	}

	return cmd
}
