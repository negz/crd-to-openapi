package pkg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/controller/openapi/builder"
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
