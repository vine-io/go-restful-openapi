package restspec

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	spec "github.com/getkin/kin-openapi/openapi3"
)

type StringAlias string
type IntAlias int

type Apple struct {
	Species     string
	Volume      int `json:"vol"`
	Things      *[]string
	Weight      json.Number
	StringAlias StringAlias
	IntAlias    IntAlias
}

func TestAppleDef(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(Apple{})

	if got, want := len(*db.Schemas), 1; got != want {
		t.Errorf("got %v want %v", got, want)
	}

	schema := (*db.Schemas)["restspec.Apple"]
	if got, want := len(schema.Value.Required), 6; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Value.Required[0], "Species"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Value.Required[1], "vol"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Value.Properties["Things"].Value.Items.Value.Type.Includes("string"), true; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Value.Properties["Things"].Value.Items.Ref, ""; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Value.Properties["Weight"].Value.Type.Includes("number"), true; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Value.Properties["StringAlias"].Value.Type.Includes("string"), true; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Value.Properties["IntAlias"].Value.Type.Includes("integer"), true; got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

type MyDictionaryResponse struct {
	Dictionary1 map[string]DictionaryValue   `json:"dictionary1"`
	Dictionary2 map[string]interface{}       `json:"dictionary2"`
	Dictionary3 map[string][]byte            `json:"dictionary3"`
	Dictionary4 map[string]string            `json:"dictionary4"`
	Dictionary5 map[string][]DictionaryValue `json:"dictionary5"`
	Dictionary6 map[string][]string          `json:"dictionary6"`
}
type DictionaryValue struct {
	Key1 string `json:"key1"`
	Key2 string `json:"key2"`
}

func TestDictionarySupport(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(MyDictionaryResponse{})

	// Make sure that only the types that we want were created.
	if got, want := len(*db.Schemas), 2; got != want {
		t.Errorf("got %v want %v", got, want)
	}

	schema, schemaFound := (*db.Schemas)["restspec.MyDictionaryResponse"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if got, want := len(schema.Value.Required), 6; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if got, want := schema.Value.Required[0], "dictionary1"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := schema.Value.Required[1], "dictionary2"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := schema.Value.Required[2], "dictionary3"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := schema.Value.Required[3], "dictionary4"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := schema.Value.Required[4], "dictionary5"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := schema.Value.Required[5], "dictionary6"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		}
		if got, want := len(schema.Value.Properties), 6; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if property, found := schema.Value.Properties["dictionary1"]; !found {
				t.Errorf("could not find property")
			} else {
				if got, want := property.Value.AdditionalProperties.Schema.Ref, "#/components/schemas/restspec.DictionaryValue"; got != want {
					t.Errorf("got %v want %v", got, want)
				}
			}
			if property, found := schema.Value.Properties["dictionary2"]; !found {
				t.Errorf("could not find property")
			} else {
				if property.Value.AdditionalProperties.Schema != nil {
					t.Errorf("unexpected additional properties")
				}
			}
			if property, found := schema.Value.Properties["dictionary3"]; !found {
				t.Errorf("could not find property")
			} else {
				if got, want := property.Value.AdditionalProperties.Schema.Value.Type.Slice()[0], "string"; got != want {
					t.Errorf("got %v want %v", got, want)
				}
			}
			if property, found := schema.Value.Properties["dictionary4"]; !found {
				t.Errorf("could not find property")
			} else {
				if got, want := property.Value.AdditionalProperties.Schema.Value.Type.Slice()[0], "string"; got != want {
					t.Errorf("got %v want %v", got, want)
				}
			}
			if property, found := schema.Value.Properties["dictionary5"]; !found {
				t.Errorf("could not find property")
			} else {
				if got, want := len(property.Value.AdditionalProperties.Schema.Value.Type.Slice()), 1; got != want {
					t.Errorf("got %v want %v", got, want)
				} else {
					if got, want := property.Value.AdditionalProperties.Schema.Value.Type.Slice()[0], "array"; got != want {
						t.Errorf("got %v want %v", got, want)
					}
					if property.Value.AdditionalProperties.Schema.Value.Items == nil {
						t.Errorf("Items not set")
					} else {
						if got, want := property.Value.AdditionalProperties.Schema.Value.Items.Ref, "#/components/schemas/restspec.DictionaryValue"; got != want {
							t.Errorf("got %v want %v", got, want)
						}
					}
				}
			}
			if property, found := schema.Value.Properties["dictionary6"]; !found {
				t.Errorf("could not find property")
			} else {
				if got, want := len(property.Value.AdditionalProperties.Schema.Value.Type.Slice()), 1; got != want {
					t.Errorf("got %v want %v", got, want)
				} else {
					if got, want := property.Value.AdditionalProperties.Schema.Value.Type.Slice()[0], "array"; got != want {
						t.Errorf("got %v want %v", got, want)
					}
					if property.Value.AdditionalProperties.Schema.Value.Items == nil {
						t.Errorf("Items not set")
					} else {
						if got, want := property.Value.AdditionalProperties.Schema.Value.Items.Value.Type.Slice()[0], "string"; got != want {
							t.Errorf("got %v want %v", got, want)
						}
					}
				}
			}
		}
	}

	schema, schemaFound = (*db.Schemas)["restspec.DictionaryValue"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if got, want := len(schema.Value.Required), 2; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if got, want := schema.Value.Required[0], "key1"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := schema.Value.Required[1], "key2"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		}
	}
}

type MyRecursiveDictionaryResponse struct {
	Dictionary1 map[string]RecursiveDictionaryValue `json:"dictionary1"`
}
type RecursiveDictionaryValue struct {
	Key1 string                              `json:"key1"`
	Key2 map[string]RecursiveDictionaryValue `json:"key2"`
}

func TestRecursiveDictionarySupport(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(MyRecursiveDictionaryResponse{})

	// Make sure that only the types that we want were created.
	if got, want := len(*db.Schemas), 2; got != want {
		t.Errorf("got %v want %v", got, want)
	}

	schema, schemaFound := (*db.Schemas)["restspec.MyRecursiveDictionaryResponse"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if got, want := len(schema.Value.Required), 1; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if got, want := schema.Value.Required[0], "dictionary1"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		}
		if got, want := len(schema.Value.Properties), 1; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if property, found := schema.Value.Properties["dictionary1"]; !found {
				t.Errorf("could not find property")
			} else {
				if got, want := property.Value.AdditionalProperties.Schema.Ref, "#/components/schemas/restspec.RecursiveDictionaryValue"; got != want {
					t.Errorf("got %v want %v", got, want)
				}
			}
		}
	}

	schema, schemaFound = (*db.Schemas)["restspec.RecursiveDictionaryValue"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if got, want := len(schema.Value.Required), 2; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if got, want := schema.Value.Required[0], "key1"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := schema.Value.Required[1], "key2"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		}
		if got, want := len(schema.Value.Properties), 2; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if property, found := schema.Value.Properties["key1"]; !found {
				t.Errorf("could not find property")
			} else {
				if property.Value.AdditionalProperties.Schema != nil {
					t.Errorf("unexpected additional properties")
				}
			}
			if property, found := schema.Value.Properties["key2"]; !found {
				t.Errorf("could not find property")
			} else {
				if got, want := property.Value.AdditionalProperties.Schema.Ref, "#/components/schemas/restspec.RecursiveDictionaryValue"; got != want {
					t.Errorf("got %v want %v", got, want)
				}
			}
		}
	}
}

func TestReturningStringToStringDictionary(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(map[string]string{})

	if got, want := len(*db.Schemas), 1; got != want {
		t.Errorf("got %v want %v", got, want)
	}

	schema, schemaFound := (*db.Schemas)["map[string]string"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if schema.Value.AdditionalProperties.Schema == nil {
			t.Errorf("AdditionalProperties not set")
		} else {
			if got, want := schema.Value.AdditionalProperties.Schema.Value.Type.Slice()[0], "string"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		}
	}
}

func TestReturningStringToSliceObjectDictionary(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(map[string][]DictionaryValue{})

	if got, want := len(*db.Schemas), 2; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	schema, schemaFound := (*db.Schemas)["map[string]||restspec.DictionaryValue"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if schema.Value.AdditionalProperties.Schema == nil {
			t.Errorf("AdditionalProperties not set")
		} else {
			if got, want := len(schema.Value.AdditionalProperties.Schema.Value.Type.Slice()), 1; got != want {
				t.Errorf("got %v want %v", got, want)
			} else {
				if got, want := schema.Value.AdditionalProperties.Schema.Value.Type.Slice()[0], "array"; got != want {
					t.Errorf("got %v want %v", got, want)
				}
				if schema.Value.AdditionalProperties.Schema.Value.Items == nil {
					t.Errorf("Items not set")
				} else {
					if got, want := schema.Value.AdditionalProperties.Schema.Value.Items.Ref, "#/components/schemas/restspec.DictionaryValue"; got != want {
						t.Errorf("got %v want %v", got, want)
					}
				}
			}
		}
	}

	schema, schemaFound = (*db.Schemas)["restspec.DictionaryValue"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if got, want := len(schema.Value.Required), 2; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if got, want := schema.Value.Required[0], "key1"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := schema.Value.Required[1], "key2"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		}
	}
}

func TestAddSliceOfPrimitiveCreatesNoType(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom([]string{})

	if got, want := len(*db.Schemas), 0; got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

type StructForSlice struct {
	Value string
}

func TestAddSliceOfStructCreatesTypeForStruct(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom([]StructForSlice{})

	if got, want := len(*db.Schemas), 1; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	schema, schemaFound := (*db.Schemas)["restspec.StructForSlice"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if got, want := len(schema.Value.Required), 1; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if got, want := schema.Value.Required[0], "Value"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		}
		if got, want := len(schema.Value.Properties), 1; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if property, found := schema.Value.Properties["Value"]; !found {
				t.Errorf("could not find property")
			} else {
				if property.Value.AdditionalProperties.Schema != nil {
					t.Errorf("unexpected additional properties")
				}
			}
		}
	}
}

type (
	X struct {
		yy []Y
	}
	Y struct {
		X
	}
)

func TestPotentialStackOverflow(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(X{})

	if got, want := len(*db.Schemas), 2; got != want {
		t.Errorf("got %v want %v", got, want)
	}

	schema := (*db.Schemas)["restspec.X"]
	if got, want := len(schema.Value.Required), 1; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Value.Required[0], "yy"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

type Foo struct {
	b *Bar
	Embed
}

// nolint:unused
type Bar struct {
	Foo `json:"foo"`
	B   struct {
		Foo
		f Foo
		b []*Bar
	}
}

type Embed struct {
	C string `json:"c"`
}

func TestRecursiveFieldStructure(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	// should not panic
	db.addModelFrom(Foo{})
}

type email struct {
	Attachments [][]byte `json:"attachments,omitempty" optional:"true"`
}

// Definition Builder fails with [][]byte #77
func TestDoubleByteArray(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(email{})
	scParent, ok := (*db.Schemas)["restspec.email"]
	if !ok {
		t.Fail()
	}
	sc, ok := scParent.Value.Properties["attachments"]
	if !ok {
		t.Fail()
	}
	if got, want := sc.Value.Type.Slice()[0], "array"; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
}

type matrix struct {
	Cells [][]string
}

func TestDoubleStringArray(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(matrix{})
	scParent, ok := (*db.Schemas)["restspec.matrix"]
	if !ok {
		t.Log(db.Schemas)
		t.Fail()
	}
	sc, ok := scParent.Value.Properties["Cells"]
	if !ok {
		t.Log(db.Schemas)
		t.Fail()
	}
	if got, want := sc.Value.Type.Slice()[0], "array"; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
	tp := sc.Value.Items.Value.Type.Slice()
	if got, want := fmt.Sprint(tp), "[array]"; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
}

type Cell struct {
	Value string
}

type DoubleStructArray struct {
	Cells [][]Cell
}

func TestDoubleStructArray(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(DoubleStructArray{})

	schema, ok := (*db.Schemas)["restspec.DoubleStructArray"]
	if !ok {
		t.Logf("definitions: %v", db.Schemas)
		t.Fail()
	}
	//t.Logf("%+v", schema)
	if want, got := schema.Value.Properties["Cells"].Value.Type.Slice()[0], "array"; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
	if want, got := schema.Value.Properties["Cells"].Value.Items.Value.Type.Slice()[0], "array"; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
	if want, got := schema.Value.Properties["Cells"].Value.Items.Value.Items.Ref, "#/components/schemas/restspec.Cell"; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
}

type childPostBuildOpenAPISchema struct {
	Value string
}

func (m childPostBuildOpenAPISchema) PostBuildOpenAPISchemaHandler(sm *spec.Schema) {
	sm.Description = "child's description"
}

type parentPostBuildOpenAPISchema struct {
	Node childPostBuildOpenAPISchema
}

func (m parentPostBuildOpenAPISchema) PostBuildOpenAPISchemaHandler(sm *spec.Schema) {
	sm.Description = "parent's description"
}

func TestPostBuildSwaggerSchema(t *testing.T) {
	var obj interface{} = parentPostBuildOpenAPISchema{}
	if _, ok := obj.(PostBuildOpenAPISchema); !ok {
		t.Fatalf("object does not implement PostBuildSwaggerSchema interface")
	}
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(obj)
	sc, ok := (*db.Schemas)["restspec.parentPostBuildOpenAPISchema"]
	if !ok {
		t.Logf("definitions: %#v", db.Schemas)
		t.Fail()
	}
	// t.Logf("sc: %#v", sc)
	if got, want := sc.Value.Description, "parent's description"; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}

	sc, ok = (*db.Schemas)["restspec.childPostBuildOpenAPISchema"]
	if !ok {
		t.Logf("definitions: %#v", db.Schemas)
		t.Fail()
	}
	//t.Logf("sc: %#v", sc)
	if got, want := sc.Value.Description, "child's description"; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
}

func TestEmbedStruct(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(Foo{})
	if got, exists := (*db.Schemas)["restspec.Embed"]; exists {
		t.Errorf("got %v want nil", got)
	}
}

type timeHolder struct {
	when time.Time
}

func TestTimeField(t *testing.T) {
	db := schemaBuilder{Schemas: &spec.Schemas{}, Config: Config{}}
	db.addModelFrom(timeHolder{})
	sc := (*db.Schemas)["restspec.timeHolder"]
	pr := sc.Value.Properties["when"]
	if got, want := pr.Value.Format, "date-time"; got != want {
		t.Errorf("got [%v:%T] want [%v:%T]", got, got, want, want)
	}
}
