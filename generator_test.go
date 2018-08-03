package swgen

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/kr/pretty"
	"github.com/swaggest/swgen/sample"
)

type TestSampleStruct struct {
	SimpleString string `json:"simple_string"`
	SimpleInt    int    `json:"simple_int"`
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
	ID  uint64 `json:"id" path:"id" required:"-"`
	Cat string `json:"category" path:"category"`
}

type simpleTestReplacement struct {
	ID  uint64 `json:"id"`
	Cat string `json:"category"`
}

type deepReplacementTag struct {
	TestField1 string `json:"test_field_1" swgen_type:"double"`
}

type testWrapParams struct {
	SimpleTestReplacement simpleTestReplacement `json:"simple_test_replacement"`
	ReplaceByTag          int                   `json:"should_be_sting" swgen_type:"string"`
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
}
type paramStructMapJSON struct {
	Field1 int                   `json:"field1"`
	Field2 string                `json:"field2"`
	Field3 simpleTestReplacement `json:"field3"`
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

type MixedStructs struct {
	mixedStruct
	mixedStructWithEnumer
}

type sliceType []testSimpleStruct

type NullFloat64 struct{}

func (NullFloat64) SwaggerDef() SwaggerData {
	typeDef := SchemaFromCommonName(CommonNameFloat)
	typeDef.TypeName = "NullFloat64"
	return SwaggerData{
		shared:    typeDef.shared,
		SchemaObj: typeDef,
	}
}

type NullBool struct{}

func (NullBool) SwaggerDef() SwaggerData {
	typeDef := SchemaFromCommonName(CommonNameBoolean)
	typeDef.TypeName = "NullBool"
	return SwaggerData{
		shared:    typeDef.shared,
		SchemaObj: typeDef,
	}
}

type NullString struct{}

func (NullString) SwaggerDef() SwaggerData {
	typeDef := SchemaFromCommonName(CommonNameString)
	typeDef.TypeName = "NullString"
	return SwaggerData{
		shared:    typeDef.shared,
		SchemaObj: typeDef,
	}
}

type NullInt64 struct{}

func (NullInt64) SwaggerDef() SwaggerData {
	typeDef := SchemaFromCommonName(CommonNameLong)
	typeDef.TypeName = "NullInt64"
	return SwaggerData{
		shared:    typeDef.shared,
		SchemaObj: typeDef,
	}
}

type NullDateTime struct{}

func (NullDateTime) SwaggerDef() SwaggerData {
	typeDef := SchemaFromCommonName(CommonNameDateTime)
	typeDef.TypeName = "NullDateTime"
	return SwaggerData{
		shared:    typeDef.shared,
		SchemaObj: typeDef,
	}
}

type NullDate struct{}

func (NullDate) SwaggerDef() SwaggerData {
	typeDef := SchemaFromCommonName(CommonNameDate)
	typeDef.TypeName = "NullDate"
	return SwaggerData{
		shared:    typeDef.shared,
		SchemaObj: typeDef,
	}
}

type NullTimestamp struct{}

func (NullTimestamp) SwaggerDef() SwaggerData {
	typeDef := SchemaFromCommonName(CommonNameLong)
	typeDef.TypeName = "NullTimestamp"
	return SwaggerData{
		shared:    typeDef.shared,
		SchemaObj: typeDef,
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
	return SwaggerData{shared: shared{
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
		AddExtendedField("x-service-type", ServiceTypeRest).
		ReflectGoTypes(true).
		IndentJSON(true)

	gen.AddTypeMap(simpleTestReplacement{}, "")
	gen.AddTypeMap(sliceType{}, float64(0))
	gen.AddTypeMap(typeMap{}, map[string]int{})

	var emptyInterface interface{}

	gen.SetPathItem(createPathItemInfo("/V1/test1", "GET", "test1 name", "test1 description", "v1", false, emptyInterface, testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/test2", "GET", "test2 name", "test2 description", "v1", false, testSimpleQueryStruct{}, testSimpleSlices{}))
	gen.SetPathItem(createPathItemInfo("/V1/test3", "PUT", "test3 name", "test3 description", "v1", false, testSimpleSlices{}, testSimpleMaps{}))
	gen.SetPathItem(createPathItemInfo("/V1/test4", "POST", "test4 name", "test4 description", "v1", false, testSimpleMaps{}, testSimpleMapList{}))
	gen.SetPathItem(createPathItemInfo("/V1/test5", "DELETE", "test5 name", "test5 description", "v1", false, testSimpleMapList{}, testSubTypes{}))
	gen.SetPathItem(createPathItemInfo("/V1/test6", "PATCH", "test6 name", "test6 description", "v1", false, testSubTypes{}, testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/test7", "OPTIONS", "test7 name", "test7 description", "v1", false, emptyInterface, testSimpleSlices{}))
	gen.SetPathItem(createPathItemInfo("/V1/test8", "GET", "test8v1 name", "test8v1 description", "v1", false, paramStructMap{}, map[string]testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/test9", "POST", "test9 name", "test9 description", "v1", false, mixedStruct{}, map[string]testSimpleStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/combine", "GET", "test1 name", "test1 description", "v1", true, emptyInterface, testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/combine", "PUT", "test3 name", "test3 description", "v1", true, testSimpleSlices{}, testSimpleMaps{}))
	gen.SetPathItem(createPathItemInfo("/V1/combine", "POST", "test4 name", "test4 description", "v1", true, testSimpleMaps{}, testSimpleMapList{}))
	gen.SetPathItem(createPathItemInfo("/V1/combine", "DELETE", "test5 name", "test5 description", "v1", true, testSimpleMapList{}, testSubTypes{}))
	gen.SetPathItem(createPathItemInfo("/V1/combine", "PATCH", "test6 name", "test6 description", "v1", true, testSubTypes{}, testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/combine", "OPTIONS", "test7 name", "test7 description", "v1", true, testSubTypes{}, testSimpleStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/pathParams/{category:[a-zA-Z]{32}}/{id:[0-9]+}", "GET", "test8 name", "test8 description", "V1", false, testPathParam{}, testSimpleStruct{}))

	//anonymous types:
	gen.SetPathItem(createPathItemInfo("/V1/anonymous1", "POST", "test10 name", "test10 description", "v1", false, testSimpleStruct{}, map[string]int64{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous2", "POST", "test11 name", "test11 description", "v1", false, testSimpleStruct{}, map[float64]string{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous3", "POST", "test12 name", "test12 description", "v1", false, testSimpleStruct{}, []string{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous4", "POST", "test13 name", "test13 description", "v1", false, testSimpleStruct{}, []int{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous5", "POST", "test14 name", "test14 description", "v1", false, testSimpleStruct{}, ""))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous6", "POST", "test15 name", "test15 description", "v1", false, testSimpleStruct{}, true))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous7", "POST", "test16 name", "test16 description", "v1", false, testSimpleStruct{}, map[string]testSimpleStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/typeReplacement1", "POST", "test9 name", "test9 description", "v1", false, testSubTypes{}, testWrapParams{}))

	gen.SetPathItem(createPathItemInfo("/V1/date1", "POST", "test date 1 name", "test date 1 description", "v1", false, testSimpleStruct{}, simpleDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/date2", "POST", "test date 2 name", "test date 2 description", "v1", false, testSimpleStruct{}, sliceDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/date3", "POST", "test date 3 name", "test date 3 description", "v1", false, testSimpleStruct{}, mapDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/date4", "POST", "test date 4 name", "test date 4 description", "v1", false, testSimpleStruct{}, []mapDateTime{}))

	gen.SetPathItem(createPathItemInfo("/V1/slice1", "POST", "test slice 1 name", "test slice 1 description", "v1", false, testSimpleStruct{}, []mapDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/slice2", "POST", "test slice 2 name", "test slice 2 description", "v1", false, testSimpleStruct{}, sliceType{}))

	gen.SetPathItem(createPathItemInfo("/V1/IDefinition1", "POST", "test IDefinition1 name", "test IDefinition1 description", "v1", false, definitionExample{}, definitionExample{}))
	gen.SetPathItem(createPathItemInfo("/V1/nullTypes", "POST", "test nulltypes", "test nulltypes", "v1", false, NullTypes{}, NullTypes{}))

	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes1", "POST", "testPrimitives", "test Primitives", "v1", false, "", 10))
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes2", "POST", "testPrimitives", "test Primitives", "v1", false, true, 1.1))
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes3", "POST", "testPrimitives", "test Primitives", "v1", false, int64(10), ""))
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes4", "POST", "testPrimitives", "test Primitives", "v1", false, int64(10), ""))

	gen.SetPathItem(createPathItemInfo("/V1/defaults1", "GET", "default", "test defaults", "v1", false, emptyInterface, testDefaults{}))
	gen.SetPathItem(createPathItemInfo("/V1/unknown", "POST", "test unknown types", "test unknown types", "v1", false, Unknown{}, Unknown{}))

	gen.SetPathItem(createPathItemInfo("/V1/empty", "POST", "test empty struct", "test empty struct", "v1", false, testEmptyStruct{}, testEmptyStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/struct-collision", "POST", "test struct name collision", "test struct name collision", "v1", false, TestSampleStruct{}, TestSampleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V2/struct-collision", "POST", "test struct name collision", "test struct name collision", "v2", false, sample.TestSampleStruct{}, sample.TestSampleStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/type-map", "POST", "test type mapping", "test type mapping", "v1", false, nil, typeMapHolder{}))

	bytes, err := gen.GenDocument()
	if err != nil {
		t.Fatalf("Failed to generate Swagger JSON document: %s", err.Error())
	}

	if err := writeLastRun("test_REST_last_run.json", bytes); err != nil {
		t.Fatalf("Failed write last run data to a file: %s", err.Error())
	}

	assertTrue(checkResult(bytes, "test_REST.json", t), t)
}

func TestJsonRpc(t *testing.T) {
	gen := NewGenerator()
	gen.SetHost("localhost")
	gen.SetInfo("swgen title", "swgen description", "term", "2.0")
	gen.SetLicense("BEER-WARE", "https://fedoraproject.org/wiki/Licensing/Beerware")
	gen.SetContact("Dylan Noblitt", "http://example.com", "dylan.noblitt@example.com")
	gen.AddExtendedField("x-service-type", ServiceTypeJSONRPC)
	gen.AddTypeMap(simpleTestReplacement{}, "")
	gen.AddTypeMap(sliceType{}, "")
	gen.IndentJSON(true)

	var emptyInterface interface{}

	gen.SetPathItem(createPathItemInfo("/V1/test1", "POST", "test1 name", "test1 description", "v1", true, emptyInterface, testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/test2", "POST", "test2 name", "test2 description", "v1", true, testSimpleQueryStruct{}, testSimpleSlices{}))
	gen.SetPathItem(createPathItemInfo("/V1/test3", "POST", "test3 name", "test3 description", "v1", true, testSimpleSlices{}, testSimpleMaps{}))
	gen.SetPathItem(createPathItemInfo("/V1/test4", "POST", "test4 name", "test4 description", "v1", true, testSimpleMaps{}, testSimpleMapList{}))
	gen.SetPathItem(createPathItemInfo("/V1/test5", "POST", "test5 name", "test5 description", "v1", true, testSimpleMapList{}, testSubTypes{}))
	gen.SetPathItem(createPathItemInfo("/V1/test6", "POST", "test6 name", "test6 description", "v1", true, testSubTypes{}, testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/test7", "POST", "test7 name", "test7 description", "v1", true, emptyInterface, testSimpleSlices{}))
	gen.SetPathItem(createPathItemInfo("/V1/test8", "POST", "test8v1 name", "test8v1 description", "v1", true, paramStructMapJSON{}, map[string]testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/test9", "POST", "test9 name", "test9 description", "v1", true, mixedStruct{}, map[string]testSimpleStruct{}))
	gen.SetPathItem(createPathItemInfo("/V1/test10", "POST", "test10 name", "test10 description", "v1", true, mixedStructWithEnumer{}, map[string]testSimpleStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/typeReplacement1", "POST", "test9 name", "test9 description", "v1", false, testSubTypes{}, testWrapParams{}))

	//anonymous types:
	gen.SetPathItem(createPathItemInfo("/V1/anonymous1", "POST", "test10 name", "test10 description", "v1", false, emptyInterface, map[string]int64{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous2", "POST", "test11 name", "test11 description", "v1", false, emptyInterface, map[float64]string{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous3", "POST", "test12 name", "test12 description", "v1", false, emptyInterface, []string{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous4", "POST", "test13 name", "test13 description", "v1", false, emptyInterface, []int{}))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous5", "POST", "test14 name", "test14 description", "v1", false, emptyInterface, ""))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous6", "POST", "test15 name", "test15 description", "v1", false, emptyInterface, true))
	gen.SetPathItem(createPathItemInfo("/V1/anonymous7", "POST", "test16 name", "test16 description", "v1", false, emptyInterface, map[string]testSimpleStruct{}))

	gen.SetPathItem(createPathItemInfo("/V1/date1", "POST", "test date 1 name", "test date 1 description", "v1", false, emptyInterface, simpleDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/date2", "POST", "test date 2 name", "test date 2 description", "v1", false, emptyInterface, sliceDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/date3", "POST", "test date 3 name", "test date 3 description", "v1", false, emptyInterface, mapDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/date4", "POST", "test date 4 name", "test date 4 description", "v1", false, emptyInterface, []mapDateTime{}))

	gen.SetPathItem(createPathItemInfo("/V1/slice1", "POST", "test slice 1 name", "test slice 1 description", "v1", false, emptyInterface, []mapDateTime{}))
	gen.SetPathItem(createPathItemInfo("/V1/slice2", "POST", "test slice 2 name", "test slice 2 description", "v1", false, emptyInterface, sliceType{}))

	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes1", "POST", "testPrimitives", "test Primitives", "v1", false, "", 10))
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes2", "POST", "testPrimitives", "test Primitives", "v1", false, true, 1.1))
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes3", "POST", "testPrimitives", "test Primitives", "v1", false, int64(10), ""))
	gen.SetPathItem(createPathItemInfo("/V1/primitiveTypes4", "POST", "testPrimitives", "test Primitives", "v1", false, int64(10), ""))

	gen.SetPathItem(createPathItemInfo("/V1/defaults1", "POST", "default", "test defaults", "v1", false, emptyInterface, testDefaults{}))

	bytes, err := gen.GenDocument()
	if err != nil {
		t.Fatalf("can not generate document: %s", err.Error())
	}

	if err := writeLastRun("test_JSON-RPC_last_run.json", bytes); err != nil {
		t.Fatalf("Failed write last run data to a file: %s", err.Error())
	}

	assertTrue(checkResult(bytes, "test_JSON-RPC.json", t), t)
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

func readTestFile(filename string) ([]byte, error) {
	bytes, readError := ioutil.ReadFile(getTestDataDir(filename))
	if readError != nil {
		return []byte{}, readError
	}

	return bytes, nil
}

func checkResult(generatedBytes []byte, expectedDataFileName string, t *testing.T) bool {
	expectedData := make(map[string]interface{})
	generatedData := make(map[string]interface{})

	expectedBytes, err := readTestFile(expectedDataFileName)
	if err != nil {
		t.Fatalf("can not read test file '%s': %s", expectedDataFileName, err.Error())
	}
	if err = json.Unmarshal(expectedBytes, &expectedData); err != nil {
		t.Fatalf("can not unmarshal '%s' data: %s", expectedDataFileName, err.Error())
	}
	if err = json.Unmarshal(generatedBytes, &generatedData); err != nil {
		t.Fatalf("can not unmarshal generated data: %s", err.Error())
	}

	for _, diff := range pretty.Diff(expectedData, generatedData) {
		pretty.Println(diff)
	}

	return reflect.DeepEqual(expectedData, generatedData)
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

	if _, err := g.SetPathItem(info); err != nil {
		t.Fatalf("error %v", err)
	}

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "http://localhost:1234/docs/swagger.json", nil)
	if err != nil {
		t.Fatalf("error when create request: %v", err)
	}

	g.ServeHTTP(w, r)

	assertTrue(w.Header().Get("Access-Control-Allow-Origin") == "*", t)
	assertTrue(w.Header().Get("Access-Control-Allow-Methods") == "GET, POST, DELETE, PUT, PATCH, OPTIONS", t)
	assertTrue(w.Header().Get("Access-Control-Allow-Headers") == "Content-Type, api_key, Authorization, X-ABC-Test", t)
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
	obj, err := gen.SetPathItem(info)
	if err != nil {
		t.Fatalf("error %v", err)
	}

	jsonSchema, err := gen.JSONSchema(*obj.Responses[http.StatusOK].Schema)
	if err != nil {
		t.Fatalf("error %v", err)
	}
	jsonSchemaJSON, err := json.MarshalIndent(jsonSchema, "", " ")
	if err != nil {
		t.Fatalf("error %v", err)
	}

	if err := writeLastRun("test_ResponseJsonSchema_last_run.json", jsonSchemaJSON); err != nil {
		t.Fatalf("Failed write last run data to a file: %s", err.Error())
	}
	checkResult(jsonSchemaJSON, "test_ResponseJsonSchema.json", t)

	jsonSchema, err = gen.ParamJSONSchema(obj.Parameters[0])
	if err != nil {
		t.Fatalf("error %v", err)
	}
	jsonSchemaJSON, err = json.MarshalIndent(jsonSchema, "", " ")
	if err != nil {
		t.Fatalf("error %v", err)
	}

	if err := writeLastRun("test_Param0JsonSchema_last_run.json", jsonSchemaJSON); err != nil {
		t.Fatalf("Failed write last run data to a file: %s", err.Error())
	}
	checkResult(jsonSchemaJSON, "test_Param0JsonSchema.json", t)

}
