package swgen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/swaggest/swgen/internal/sample"
	"github.com/swaggest/swgen/internal/sample/experiment"
	"github.com/swaggest/swgen/internal/sample/experiment/variation"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

type TestSampleStruct struct {
	SimpleString string `json:"simple_string"`
	SimpleInt    int    `json:"simple_int"`

	Sub      TestSubStruct   `json:"sub"`
	SubSlice []TestSubStruct `json:"sub_slice"`
}

type TestSubStruct struct {
	SubInt int `json:"sample_int"`
}

type testEmptyStruct struct{}

type testSimpleStruct struct {
	SimpleString  string  `json:"simple_string"`
	SimpleInt     int     `json:"simple_int"`
	SimpleInt32   int32   `json:"simple_int32"`
	SimpleInt64   int64   `json:"simple_int64"`
	SimpleUInt32  uint32  `json:"simple_uint32"`
	SimpleUInt64  uint64  `json:"simple_uint64"`
	SimpleFloat32 float32 `json:"simple_float32"`
	SimpleFloat64 float64 `json:"simple_float64"`
	SimpleBool    bool    `json:"simple_bool"`
	IgnoreField   string  `json:"-"`
}

type testSimpleQueryStruct struct {
	SimpleString  string  `query:"simple_string"`
	SimpleInt     int     `query:"simple_int"`
	SimpleInt32   int32   `query:"simple_int32"`
	SimpleInt64   int64   `query:"simple_int64"`
	SimpleUInt32  uint32  `query:"simple_uint32"`
	SimpleUInt64  uint64  `query:"simple_uint64"`
	SimpleFloat32 float32 `query:"simple_float32"`
	SimpleFloat64 float64 `query:"simple_float64"`
	SimpleBool    bool    `query:"simple_bool"`
	IgnoreField   string  `query:"-"`
}

type testSimpleSlices struct {
	ListString  []string  `json:"list_string"`
	ListInt     []int     `json:"list_int"`
	ListInt32   []int32   `json:"list_int32"`
	ListInt64   []int64   `json:"list_int64"`
	ListUInt32  []uint32  `json:"list_uint32"`
	ListUInt64  []uint64  `json:"list_uint64"`
	ListFloat32 []float32 `json:"list_float32"`
	ListFloat64 []float64 `json:"list_float64"`
	ListBool    []bool    `json:"list_bool"`
}

type testSimpleMaps struct {
	MapString  map[string]string  `json:"map_string"`
	MapInt     map[string]int     `json:"map_int"`
	MapInt32   map[string]int32   `json:"map_int32"`
	MapInt64   map[string]int64   `json:"map_int64"`
	MapUInt32  map[string]uint32  `json:"map_uint32"`
	MapUInt64  map[string]uint64  `json:"map_uint64"`
	MapFloat32 map[string]float32 `json:"map_float32"`
	MapFloat64 map[string]float64 `json:"map_float64"`
	MapBool    map[string]bool    `json:"map_bool"`
}

type testSimpleMapList struct {
	MapListString  []map[string]string  `json:"map_list_string"`
	MapListInt     []map[string]int     `json:"map_list_int"`
	MapListInt32   []map[string]int32   `json:"map_list_int32"`
	MapListInt64   []map[string]int64   `json:"map_list_int64"`
	MapListUInt32  []map[string]uint32  `json:"map_list_uint32"`
	MapListUInt64  []map[string]uint64  `json:"map_list_uint64"`
	MapListFloat32 []map[string]float32 `json:"map_list_float32"`
	MapListFloat64 []map[string]float64 `json:"map_list_float64"`
	MapListBool    []map[string]bool    `json:"map_list_bool"`
}

type testSubTypes struct {
	TestSimpleStruct  testSimpleStruct  `json:"test_simple_struct"`
	TestSimpleSlices  testSimpleSlices  `json:"test_simple_slices"`
	TestSimpleMaps    testSimpleMaps    `json:"test_simple_maps"`
	TestSimpleMapList testSimpleMapList `json:"test_simple_map_list"`
}

type testPathParam struct {
	ID  uint64 `json:"id" path:"id" required:"false"`
	Cat string `json:"category" path:"category"`
}

type simpleTestReplacement struct {
	ID  uint64 `json:"id"`
	Cat string `json:"category"`
}

type deepReplacementTag struct {
	TestField1 string `json:"test_field_1" type:"number" format:"double"`
}

type testWrapParams struct {
	SimpleTestReplacement simpleTestReplacement `json:"simple_test_replacement"`
	ReplaceByTag          int                   `json:"should_be_sting" type:"string" format:"-"`
	DeepReplacementTag    deepReplacementTag    `json:"deep_replacement"`
}

type simpleDateTime struct {
	Time time.Time `json:"time"`
}

type sliceDateTime struct {
	Items []simpleDateTime `json:"items"`
}

type mapDateTime struct {
	Items map[string]simpleDateTime `json:"items"`
}

type paramStructMap struct {
	Field1 int                   `query:"field1"`
	Field2 string                `query:"field2"`
	Field3 simpleTestReplacement `query:"field3"`
	Field4 []int64               `query:"field4" collectionFormat:"csv"`
}

type AnonymousField struct {
	AnonProp int `json:"anonProp"`
}

type mixedStruct struct {
	AnonymousField
	FieldQuery int `query:"fieldQuery"`
	FieldBody  int `json:"fieldBody"`
}

type typeMapHolder struct {
	M typeMap `json:"m"`
}

type typeMap struct {
	R1 int `json:"1"`
	R2 int `json:"2"`
	R3 int `json:"3"`
	R4 int `json:"4"`
	R5 int `json:"5"`
}

type Gender int

func (Gender) NamedEnum() ([]interface{}, []string) {
	return []interface{}{
			PreferNotToDisclose,
			Male,
			Female,
			LGBT,
		}, []string{
			"PreferNotToDisclose",
			"Male",
			"Female",
			"LGBT",
		}
}

const (
	PreferNotToDisclose Gender = iota
	Male
	Female
	LGBT
)

type Flag string

func (Flag) NamedEnum() ([]interface{}, []string) {
	return []interface{}{Flag("Foo"), Flag("Bar")}, []string{"Foo", "Bar"}
}

type mixedStructWithEnumer struct {
	Gender Gender `query:"gender"`
	Flag   Flag   `query:"flag"`
}

type sliceType []testSimpleStruct

type NullFloat64 struct{}

func (NullFloat64) SwaggerDef() SwaggerData {
	typeDef := schemaFromCommonName(commonNameFloat)
	typeDef.TypeName = "NullFloat64"
	return SwaggerData{
		CommonFields: typeDef.CommonFields,
		SchemaObj:    typeDef,
	}
}

type NullBool struct{}

func (NullBool) SwaggerDef() SwaggerData {
	typeDef := schemaFromCommonName(commonNameBoolean)
	typeDef.TypeName = "NullBool"
	return SwaggerData{
		CommonFields: typeDef.CommonFields,
		SchemaObj:    typeDef,
	}
}

type NullString struct{}

func (NullString) SwaggerDef() SwaggerData {
	typeDef := schemaFromCommonName(commonNameString)
	typeDef.TypeName = "NullString"
	return SwaggerData{
		CommonFields: typeDef.CommonFields,
		SchemaObj:    typeDef,
	}
}

type NullInt64 struct{}

func (NullInt64) SwaggerDef() SwaggerData {
	typeDef := schemaFromCommonName(commonNameLong)
	typeDef.TypeName = "NullInt64"
	return SwaggerData{
		CommonFields: typeDef.CommonFields,
		SchemaObj:    typeDef,
	}
}

type NullDateTime struct{}

func (NullDateTime) SwaggerDef() SwaggerData {
	typeDef := schemaFromCommonName(commonNameDateTime)
	typeDef.TypeName = "NullDateTime"
	return SwaggerData{
		CommonFields: typeDef.CommonFields,
		SchemaObj:    typeDef,
	}
}

type NullDate struct{}

func (NullDate) SwaggerDef() SwaggerData {
	typeDef := schemaFromCommonName(commonNameDate)
	typeDef.TypeName = "NullDate"
	return SwaggerData{
		CommonFields: typeDef.CommonFields,
		SchemaObj:    typeDef,
	}
}

type NullTimestamp struct{}

func (NullTimestamp) SwaggerDef() SwaggerData {
	typeDef := schemaFromCommonName(commonNameLong)
	typeDef.TypeName = "NullTimestamp"
	return SwaggerData{
		CommonFields: typeDef.CommonFields,
		SchemaObj:    typeDef,
	}
}

type testDefaults struct {
	Field1 int            `json:"field1" default:"25"`
	Field2 float64        `json:"field2" default:"25.5"`
	Field3 string         `json:"field3" default:"test"`
	Field4 bool           `json:"field4" default:"true"`
	Field5 []int          `json:"field5" default:"[1, 2, 3]"`
	Field6 map[string]int `json:"field6" default:"{\"test\": 1}"`
	Field7 *uint          `json:"field7" default:"25"`
}

type NullTypes struct {
	Float     NullFloat64   `json:"null_float"`
	Bool      NullBool      `json:"null_bool"`
	String    NullString    `json:"null_string"`
	Int       NullInt64     `json:"null_int"`
	DateTime  NullDateTime  `json:"null_date_time"`
	Date      NullDate      `json:"null_date"`
	Timestamp NullTimestamp `json:"null_timestamp"`
}

type Unknown struct {
	Anything interface{}      `json:"anything"`
	Whatever *json.RawMessage `json:"whatever"`
}

var _ SchemaDefinition = definitionExample{}

type definitionExample struct{}

func (defEx definitionExample) SwaggerDef() SwaggerData {
	return SwaggerData{CommonFields: CommonFields{
		Type:   "string",
		Format: "byte",
	}}
}

func createPathItemInfo(path, method, title, description, tag string, deprecated bool, request, response interface{}) PathItemInfo {
	return PathItemInfo{
		Path:        path,
		Method:      method,
		Title:       title,
		Description: description,
		Tag:         tag,
		Deprecated:  deprecated,

		Request:  request,
		Response: response,
	}
}

func TestREST(t *testing.T) {
	gen := NewGenerator()
	gen.SetHost("localhost").
		SetBasePath("/").
		SetInfo("swgen title", "swgen description", "term", "2.0").
		SetLicense("BEER-WARE", "https://fedoraproject.org/wiki/Licensing/Beerware").
		SetContact("Dylan Noblitt", "http://example.com", "dylan.noblitt@example.com").
		ReflectGoTypes(true).
		IndentJSON(true)

	gen.AddTypeMap(simpleTestReplacement{}, "")
	gen.AddTypeMap(sliceType{}, float64(0))
	gen.AddTypeMap(typeMap{}, map[string]int{})

	var emptyInterface interface{}

	gen.SetPathItem(createPathItemInfo("/V1/test1", "GET", "test1 name",
		"test1 description", "v1", false, emptyInterface, testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/test2", "GET", "test2 name",
		"test2 description", "v1", false, testSimpleQueryStruct{}, testSimpleSlices{}))
	gen.SetPathItem(createPathItemInfo("/V1/test3", "PUT", "test3 name",
		"test3 description", "v1", false, testSimpleSlices{}, testSimpleMaps{}))
	gen.SetPathItem(createPathItemInfo("/V1/test4", "POST", "test4 name",
		"test4 description", "v1", false, testSimpleMaps{}, testSimpleMapList{}))
	gen.SetPathItem(createPathItemInfo("/V1/test5", "DELETE", "test5 name",
		"test5 description", "v1", false, testSimpleMapList{}, testSubTypes{}))
	gen.SetPathItem(createPathItemInfo("/V1/test6", "PATCH", "test6 name",
		"test6 description", "v1", false, testSubTypes{}, testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/test7", "OPTIONS", "test7 name",
		"test7 description", "v1", false, emptyInterface, testSimpleSlices{}))
	gen.SetPathItem(createPathItemInfo("/V1/test8", "GET", "test8v1 name",
		"test8v1 description", "v1", false, paramStructMap{}, map[string]testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/test9", "POST", "test9 name",
		"test9 description", "v1", false, mixedStruct{}, map[string]testSimpleStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/combine", "GET", "test1 name",
		"test1 description", "v1", true, emptyInterface, testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/combine", "PUT", "test3 name",
		"test3 description", "v1", true, testSimpleSlices{}, testSimpleMaps{}))
	gen.SetPathItem(createPathItemInfo("/V1/combine", "POST", "test4 name",
		"test4 description", "v1", true, testSimpleMaps{}, testSimpleMapList{}))
	gen.SetPathItem(createPathItemInfo("/V1/combine", "DELETE", "test5 name",
		"test5 description", "v1", true, testSimpleMapList{}, testSubTypes{}))
	gen.SetPathItem(createPathItemInfo("/V1/combine", "PATCH", "test6 name",
		"test6 description", "v1", true, testSubTypes{}, testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/combine", "OPTIONS", "test7 name",
		"test7 description", "v1", true, testSubTypes{}, testSimpleStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/pathParams/{category:[a-zA-Z]{32}}/{id:[0-9]+}", "GET", "test8 name",
		"test8 description", "V1", false, testPathParam{}, testSimpleStruct{}))

	//anonymous types:
	gen.SetPathItem(createPathItemInfo("/V1/anonymous1", "POST", "test10 name",
		"test10 description", "v1", false, testSimpleStruct{}, map[string]int64{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous2", "POST", "test11 name",
		"test11 description", "v1", false, testSimpleStruct{}, map[float64]string{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous3", "POST", "test12 name",
		"test12 description", "v1", false, testSimpleStruct{}, []string{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous4", "POST", "test13 name",
		"test13 description", "v1", false, testSimpleStruct{}, []int{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous5", "POST", "test14 name",
		"test14 description", "v1", false, testSimpleStruct{}, ""))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous6", "POST", "test15 name",
		"test15 description", "v1", false, testSimpleStruct{}, true))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous7", "POST", "test16 name",
		"test16 description", "v1", false, testSimpleStruct{}, map[string]testSimpleStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/typeReplacement1", "POST", "test9 name",
		"test9 description", "v1", false, testSubTypes{}, testWrapParams{}))

	gen.SetPathItem(createPathItemInfo("/V1/date1", "POST", "test date 1 name",
		"test date 1 description", "v1", false, testSimpleStruct{}, simpleDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/date2", "POST", "test date 2 name",
		"test date 2 description", "v1", false, testSimpleStruct{}, sliceDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/date3", "POST", "test date 3 name",
		"test date 3 description", "v1", false, testSimpleStruct{}, mapDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/date4", "POST", "test date 4 name",
		"test date 4 description", "v1", false, testSimpleStruct{}, []mapDateTime{}))

	gen.SetPathItem(createPathItemInfo("/V1/slice1", "POST", "test slice 1 name",
		"test slice 1 description", "v1", false, testSimpleStruct{}, []mapDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/slice2", "POST", "test slice 2 name",
		"test slice 2 description", "v1", false, testSimpleStruct{}, sliceType{}))

	gen.SetPathItem(createPathItemInfo("/V1/IDefinition1", "POST", "test IDefinition1 name",
		"test IDefinition1 description", "v1", false, definitionExample{}, definitionExample{}))
	gen.SetPathItem(createPathItemInfo("/V1/nullTypes", "POST", "test nulltypes",
		"test nulltypes", "v1", false, NullTypes{}, NullTypes{}))

	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes1", "POST", "testPrimitives",
		"test Primitives", "v1", false, "", 10))
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes2", "POST", "testPrimitives",
		"test Primitives", "v1", false, true, 1.1))
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes3", "POST", "testPrimitives",
		"test Primitives", "v1", false, int64(10), ""))
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes4", "POST", "testPrimitives",
		"test Primitives", "v1", false, int64(10), ""))

	gen.SetPathItem(createPathItemInfo("/V1/defaults1", "GET", "default",
		"test defaults", "v1", false, emptyInterface, testDefaults{}))
	gen.SetPathItem(createPathItemInfo("/V1/unknown", "POST", "test unknown types",
		"test unknown types", "v1", false, Unknown{}, Unknown{}))

	gen.SetPathItem(createPathItemInfo("/V1/empty", "POST", "test empty struct",
		"test empty struct", "v1", false, testEmptyStruct{}, testEmptyStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/struct-collision", "POST", "test struct name collision",
		"test struct name collision", "v1", false, TestSampleStruct{}, TestSampleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V2/struct-collision", "POST", "test struct name collision",
		"test struct name collision", "v2", false, sample.TestSampleStruct{}, sample.TestSampleStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/type-map", "POST", "test type mapping",
		"test type mapping", "v1", false, nil, typeMapHolder{}))

	bytes, err := gen.GenDocument()
	assert.NoError(t, err)

	assert.NoError(t, writeLastRun("test_REST_last_run.json", bytes))

	expected := readTestFile(t, "test_REST.json")
	assert.JSONEq(t, expected, string(bytes), coloredJSONDiff(expected, string(bytes)))
}

func getTestDataDir(filename string) string {
	pwd, err := os.Getwd()
	if err != nil {
		return filename
	}

	return path.Join(pwd, "testdata", filename)
}

func writeLastRun(filename string, data []byte) error {
	return ioutil.WriteFile(getTestDataDir(filename), data, os.ModePerm)
}

func readTestFile(t *testing.T, filename string) string {
	bytes, readError := ioutil.ReadFile(getTestDataDir(filename))
	assert.NoError(t, readError)

	return string(bytes)
}

func coloredJSONDiff(expected, generated string) string {
	expectedData := make(map[string]interface{})
	generatedData := make(map[string]interface{})
	expectedBytes := []byte(expected)
	generatedBytes := []byte(generated)

	if err := json.Unmarshal(expectedBytes, &expectedData); err != nil {
		return fmt.Sprintf("can not unmarshal expected data: %s", err.Error())
	}
	if err := json.Unmarshal(generatedBytes, &generatedData); err != nil {
		return fmt.Sprintf("can not unmarshal generated data: %s", err.Error())
	}

	diff, err := gojsondiff.New().Compare(expectedBytes, generatedBytes)
	if err != nil {
		return err.Error()
	}

	if diff.Modified() {
		config := formatter.AsciiFormatterConfig{
			Coloring:       true,
			ShowArrayIndex: true,
		}

		f := formatter.NewAsciiFormatter(expectedData, config)
		diffString, _ := f.Format(diff)
		return diffString
	}
	return ""
}

func TestCORSSupport(t *testing.T) {
	g := NewGenerator()
	g.EnableCORS(true, "X-ABC-Test").
		SetHost("localhost:1234")

	info := PathItemInfo{
		Path:        "/v1/test/handler",
		Title:       "TestHandler",
		Description: "This is just a test handler with GET request",
		Method:      "GET",
	}

	g.SetPathItem(info)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://localhost:1234/docs/swagger.json", nil)
	assert.NoError(t, err)

	g.ServeHTTP(w, r)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, DELETE, PUT, PATCH, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, api_key, Authorization, X-ABC-Test", w.Header().Get("Access-Control-Allow-Headers"))
}

func TestGenerator_JsonSchema(t *testing.T) {
	gen := NewGenerator()
	gen.SetHost("localhost")
	gen.SetInfo("swgen title", "swgen description", "term", "2.0")
	gen.IndentJSON(true)

	info := PathItemInfo{
		Method:   "GET",
		Path:     "/any",
		Request:  new(mixedStructWithEnumer),
		Response: new(map[string]testSimpleStruct),
	}
	obj := gen.SetPathItem(info)

	jsonSchema, err := gen.JSONSchema(*obj.Responses[http.StatusOK].Schema)
	assert.NoError(t, err)
	jsonSchemaJSON, err := json.MarshalIndent(jsonSchema, "", " ")
	assert.NoError(t, err)

	assert.NoError(t, writeLastRun("test_ResponseJsonSchema_last_run.json", jsonSchemaJSON))

	expected := readTestFile(t, "test_ResponseJsonSchema.json")
	assert.JSONEq(t, expected, string(jsonSchemaJSON), coloredJSONDiff(expected, string(jsonSchemaJSON)))

	jsonSchema, err = gen.ParamJSONSchema(obj.Parameters[0])
	assert.NoError(t, err)
	jsonSchemaJSON, err = json.MarshalIndent(jsonSchema, "", " ")
	assert.NoError(t, err)

	assert.NoError(t, writeLastRun("test_Param0JsonSchema_last_run.json", jsonSchemaJSON))
	expected = readTestFile(t, "test_Param0JsonSchema.json")
	assert.JSONEq(t, expected, string(jsonSchemaJSON), coloredJSONDiff(expected, string(jsonSchemaJSON)))
}

func TestGenerator_GenDocument_StructCollision(t *testing.T) {
	gen := NewGenerator()
	gen.SetHost("localhost")
	gen.SetInfo("swgen title", "swgen description", "term", "2.0")
	gen.IndentJSON(true)
	gen.ReflectGoTypes(true)

	info := PathItemInfo{
		Method:   http.MethodPost,
		Path:     "/any",
		Request:  new(experiment.PostRequest),
		Response: new(experiment.Entity),
	}
	gen.SetPathItem(info)

	generatedBytes, err := gen.GenDocument()
	assert.NoError(t, err)

	assert.NoError(t, writeLastRun("struct_collision_last_run.json", generatedBytes))

	expected := readTestFile(t, "struct_collision.json")
	assert.JSONEq(t, expected, string(generatedBytes), coloredJSONDiff(expected, string(generatedBytes)))
}

func TestGenerator_GenDocument_StructCollisionPackagePrefix(t *testing.T) {
	gen := NewGenerator()
	gen.SetHost("localhost")
	gen.SetInfo("swgen title", "swgen description", "term", "2.0")
	gen.IndentJSON(true)
	gen.ReflectGoTypes(true)
	gen.AddPackagePrefix(true)

	info := PathItemInfo{
		Method:   http.MethodPost,
		Path:     "/any",
		Request:  new(experiment.PostRequest),
		Response: new(experiment.Entity),
	}
	gen.SetPathItem(info)

	generatedBytes, err := gen.GenDocument()
	assert.NoError(t, err)

	assert.NoError(t, writeLastRun("struct_collision_package_prefix_last_run.json", generatedBytes))
	expected := readTestFile(t, "struct_collision_package_prefix.json")
	assert.JSONEq(t, expected, string(generatedBytes), coloredJSONDiff(expected, string(generatedBytes)))
}

type (
	experimentEntity            experiment.Entity
	experimentMetadata          experiment.Metadata
	experimentVariationEntity   variation.Entity
	experimentVariationMetadata variation.Metadata
)

func TestGenerator_GenDocument_StructCollisionWithExplicitRemapping(t *testing.T) {
	gen := NewGenerator()
	gen.SetHost("localhost")
	gen.SetInfo("swgen title", "swgen description", "term", "2.0")
	gen.IndentJSON(true)
	//gen.ReflectGoTypes(true)
	gen.AddTypeMap(new(experiment.Entity), new(experimentEntity))
	gen.AddTypeMap(new(variation.Entity), new(experimentVariationEntity))
	gen.AddTypeMap(new(experiment.Metadata), new(experimentMetadata))
	gen.AddTypeMap(new(variation.Metadata), new(experimentVariationMetadata))

	info := PathItemInfo{
		Method:   http.MethodPost,
		Path:     "/any",
		Request:  new(experiment.PostRequest),
		Response: new(experiment.Entity),
	}
	gen.SetPathItem(info)

	generatedBytes, err := gen.GenDocument()
	assert.NoError(t, err)

	assert.NoError(t, writeLastRun("struct_collision_mapped_last_run.json", generatedBytes))
	expected := readTestFile(t, "struct_collision_mapped.json")
	assert.JSONEq(t, expected, string(generatedBytes), coloredJSONDiff(expected, string(generatedBytes)))
}

func TestGenerator_GenDocument_CorrectRef(t *testing.T) {
	gen := NewGenerator()
	gen.IndentJSON(true)
	gen.ReflectGoTypes(true)

	gen.SetPathItem(createPathItemInfo("/V1/struct-collision", "POST", "test struct name collision", "test struct name collision", "v1", false, TestSampleStruct{}, TestSampleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V2/struct-collision", "POST", "test struct name collision", "test struct name collision", "v2", false, sample.TestSampleStruct{}, sample.TestSampleStruct{}))

	bytes, err := gen.GenDocument()
	assert.NoError(t, err)

	assert.NoError(t, writeLastRun("struct_collision_correct_ref_last_run.json", bytes))

	expected := readTestFile(t, "struct_collision_correct_ref.json")
	assert.JSONEq(t, expected, string(bytes), coloredJSONDiff(expected, string(bytes)))
}
