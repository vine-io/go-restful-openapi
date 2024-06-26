package restspec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/emicklei/go-restful/v3"
	spec "github.com/getkin/kin-openapi/openapi3"
)

func dummy(i *restful.Request, o *restful.Response) {}

type Sample struct {
	ID    string `swagger:"required"`
	Root  Item   `json:"root" description:"root desc"`
	Items []Item
}

type Item struct {
	ItemName string `json:"name"`
}

func asJSON(v interface{}) string {
	data, _ := json.MarshalIndent(v, " ", " ")
	return string(data)
}

func compareJSON(t *testing.T, actualJSONAsString string, expectedJSONAsString string) bool {
	success := false
	var actualMap map[string]interface{}
	json.Unmarshal([]byte(actualJSONAsString), &actualMap)
	var expectedMap map[string]interface{}
	err := json.Unmarshal([]byte(expectedJSONAsString), &expectedMap)
	if err != nil {
		var actualArray []interface{}
		json.Unmarshal([]byte(actualJSONAsString), &actualArray)
		var expectedArray []interface{}
		err := json.Unmarshal([]byte(expectedJSONAsString), &expectedArray)
		success = reflect.DeepEqual(actualArray, expectedArray)
		if err != nil {
			t.Fatalf("Unparsable expected JSON: %s, actual: %v, expected: %v", err, actualJSONAsString, expectedJSONAsString)
		}
	} else {
		success = reflect.DeepEqual(actualMap, expectedMap)
	}
	if !success {
		t.Log("---- expected -----")
		t.Log(withLineNumbers(expectedJSONAsString))
		t.Log("---- actual -----")
		t.Log(withLineNumbers(actualJSONAsString))
		t.Log("---- raw -----")
		t.Log(actualJSONAsString)
		t.Error("there are differences")
		return false
	}
	return true
}

// nolint:unused
func withLineNumbers(content string) string {
	var buffer bytes.Buffer
	lines := strings.Split(content, "\n")
	for i, each := range lines {
		buffer.WriteString(fmt.Sprintf("%d:%s\n", i, each))
	}
	return buffer.String()
}

// mergeStrings returns a new string slice without duplicates.
func mergeStrings(left, right []string) (merged []string) {
	include := func(next string) {
		for _, dup := range merged {
			if next == dup {
				return
			}
		}
		merged = append(merged, next)
	}
	for _, each := range left {
		include(each)
	}
	for _, each := range right {
		include(each)
	}
	return
}

func schemasFromStructWithConfig(sample interface{}, config Config) spec.Schemas {
	schemas := &spec.Schemas{}
	builder := schemaBuilder{Schemas: schemas, Config: config}
	builder.addModelFrom(sample)
	return *schemas
}

func schemasFromStruct(sample interface{}) spec.Schemas {
	return schemasFromStructWithConfig(sample, Config{})
}
