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

	"github.com/thoas/go-funk"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/controller/openapi/builder"
	"sigs.k8s.io/yaml"
)

func Convert(version string, v2 bool) error {
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
		var versions []string
		for _, crdVersion := range crd.Spec.Versions {
			versions = append(versions, crdVersion.Name)
		}
		if !funk.ContainsString(versions, version) {
			return fmt.Errorf("CRD doesn't have version %s, but has %v", version, versions)
		}
	}
	var output interface{}

	if v2 {
		output, err = builder.BuildOpenAPIV2(crd, version, builder.Options{
			V2: true,
		})
	} else {
		output, err = builder.BuildOpenAPIV3(crd, version, builder.Options{})
	}

	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(output)
	if err != nil {
		return err
	}
	return nil
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
