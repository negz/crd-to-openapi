package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/controller/openapi/builder"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/yaml"
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
	var injectDefaultGVK bool

	flags.StringVar(&crdVersion, "crd-version", "", "set crd version")
	flags.BoolVar(&openapi2, "output-openapi-v2", false, "output in OpenAPI v2 format, the default format is OpenAPI v3")
	flags.BoolVar(&injectDefaultGVK, "inject-default-gvk", false, "inject default apiVersion and kind values")

	cmd := &cobra.Command{
		Short: "CRD To OpenAPI",
		Long:  "Read CRD from stdin and write OpenAPI spec to stdout.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Convert(crdVersion, openapi2, injectDefaultGVK)
		},
	}

	return cmd
}

func Convert(version string, v2, injectGVK bool) error {
	reader := bufio.NewReader(os.Stdin)
	src, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	crd, err := ReadCRD(src)
	if err != nil {
		return err
	}

	if len(crd.Spec.Versions) == 0 {
		return fmt.Errorf("CRD doesn't have any version")
	}

	if version == "" {
		var versions SortableVersions
		for _, crdVersion := range crd.Spec.Versions {
			versions = append(versions, crdVersion.Name)
		}
		sort.Sort(versions)
		version = versions[len(versions)-1]
	} else {
		exists := map[string]bool{}
		for _, v := range crd.Spec.Versions {
			exists[v.Name] = true
		}
		if !exists[version] {
			return fmt.Errorf("CRD doesn't have version %s, but has %v", version, exists)
		}
	}

	encoder := json.NewEncoder(os.Stdout) // TODO(negz): stdout
	encoder.SetIndent("", "  ")
	if v2 {
		output, err := builder.BuildOpenAPIV2(crd, version, builder.Options{V2: true})
		if err != nil {
			return err
		}
		return encoder.Encode(output)
	}

	output, err := builder.BuildOpenAPIV3(crd, version, builder.Options{})
	if err != nil {
		return err
	}

	if !injectGVK {
		return encoder.Encode(output)
	}

	// Schemas have names like io.upbound.aws.eks.v1beta2.Cluster. We want to
	// find the ones that represent our custom resource's GVKs, and add default
	// values for the top-level apiVersion and kind fields.

	// First reverse the group to get e.g. io.upbound.aws.eks.
	g := strings.Split(crd.Spec.Group, ".")
	for i, j := 0, len(g)-1; i < j; i, j = i+1, j-1 {
		g[i], g[j] = g[j], g[i]
	}
	reverseGroup := strings.Join(g, ".")

	for name, s := range output.Components.Schemas {
		if !strings.HasPrefix(name, reverseGroup+".") {
			// Not part of our API group.
			continue
		}

		for _, v := range crd.Spec.Versions {
			if strings.TrimPrefix(name, reverseGroup+".") != v.Name+"."+crd.Spec.Names.Kind {
				// Not our version or kind.
				continue
			}

			// This schema is for one of our CR versions.
			AddDefaultAPIVersionAndKind(s, schema.GroupVersionKind{
				Group:   crd.Spec.Group,
				Version: v.Name,
				Kind:    crd.Spec.Names.Kind,
			})
		}
	}

	return encoder.Encode(output)
}

func AddDefaultAPIVersionAndKind(s *spec.Schema, gvk schema.GroupVersionKind) {
	for name := range s.Properties {
		ss := s.Properties[name]
		switch name {
		case "apiVersion":
			ss.Default = gvk.GroupVersion().String()
		case "kind":
			ss.Default = gvk.Kind
		}
		s.Properties[name] = ss
	}
}

func ReadCRD(src []byte) (*apiextensionsv1.CustomResourceDefinition, error) {
	var crd apiextensionsv1.CustomResourceDefinition

	var err error
	if err = json.Unmarshal(src, &crd); err != nil {
		err = yaml.Unmarshal(src, &crd)
	}

	return &crd, err
}

type SortableVersions []string

func (a SortableVersions) Len() int      { return len(a) }
func (a SortableVersions) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortableVersions) Less(i, j int) bool {
	vi, vj := strings.TrimLeft(a[i], "v"), strings.TrimLeft(a[j], "v")
	major := regexp.MustCompile("^[0-9]+")
	viMajor, vjMajor := major.FindString(vi), major.FindString(vj)
	viRemaining, vjRemaining := strings.TrimLeft(vi, viMajor), strings.TrimLeft(vj, vjMajor)
	switch {
	case len(viRemaining) == 0 && len(vjRemaining) == 0:
		return viMajor < vjMajor
	case len(viRemaining) == 0 && len(vjRemaining) != 0:
		// stable version is greater than unstable version
		return false
	case len(viRemaining) != 0 && len(vjRemaining) == 0:
		// stable version is greater than unstable version
		return true
	}
	// neither are stable versions
	if viMajor != vjMajor {
		return viMajor < vjMajor
	}
	// assuming at most we have one alpha or one beta version, so if vi contains "alpha", it's the lesser one.
	return strings.Contains(viRemaining, "alpha")
}
