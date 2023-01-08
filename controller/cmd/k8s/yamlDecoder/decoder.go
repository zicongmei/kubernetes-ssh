package yamlDecoder

import (
	"fmt"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const decoderBufferSizeBytes = 10000

func Decode(input string) ([]*unstructured.Unstructured, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(input), decoderBufferSizeBytes)

	var resources []*unstructured.Unstructured
	for {
		u := &unstructured.Unstructured{}
		if err := decoder.Decode(u); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error in decode yaml: %v", err)
		}
		if len(u.Object) != 0 {
			resources = append(resources, u)
		}
	}
	return resources, nil
}
