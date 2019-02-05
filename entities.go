package swgen

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
)

// Document represent for a document object of swagger data
// see http://swagger.io/specification/
type Document struct {
	Version             string                 `json:"swagger"`                       // Specifies the Swagger Specification version being used
	Info                InfoObj                `json:"info"`                          // Provides metadata about the API
	Host                string                 `json:"host,omitempty"`                // The host (name or ip) serving the API
	BasePath            string                 `json:"basePath,omitempty"`            // The base path on which the API is served, which is relative to the host
	Schemes             []string               `json:"schemes,omitempty"`             // Values MUST be from the list: "http", "https", "ws", "wss"
	Paths               map[string]PathItem    `json:"paths"`                         // The available paths and operations for the API
	Definitions         map[string]SchemaObj   `json:"definitions,omitempty"`         // An object to hold data types produced and consumed by operations
	SecurityDefinitions map[string]SecurityDef `json:"securityDefinitions,omitempty"` // An object to hold available security mechanisms
	additionalData
}

// MarshalJSON marshal Document with additionalData inlined
func (s Document) MarshalJSON() ([]byte, error) {
	type i Document
	return s.marshalJSONWithStruct(i(s))
}

// InfoObj provides metadata about the API
type InfoObj struct {
	Title          string     `json:"title"` // The title of the application
	Description    string     `json:"description"`
	TermsOfService string     `json:"termsOfService"`
	Contact        ContactObj `json:"contact"`
	License        LicenseObj `json:"license"`
	Version        string     `json:"version"`
}

// ContactObj contains contact information for the exposed API
type ContactObj struct {
	Name  string `json:"name"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// LicenseObj license information for the exposed API
type LicenseObj struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// PathItem describes the operations available on a single path
// see http://swagger.io/specification/#pathItemObject
type PathItem struct {
	Ref     string        `json:"$ref,omitempty"`
	Get     *OperationObj `json:"get,omitempty"`
	Put     *OperationObj `json:"put,omitempty"`
	Post    *OperationObj `json:"post,omitempty"`
	Delete  *OperationObj `json:"delete,omitempty"`
	Options *OperationObj `json:"options,omitempty"`
	Head    *OperationObj `json:"head,omitempty"`
	Patch   *OperationObj `json:"patch,omitempty"`
	Params  *ParamObj     `json:"parameters,omitempty"`
}

// HasMethod returns true if in path item already have operation for given method
func (pi PathItem) HasMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet:
		return pi.Get != nil
	case http.MethodPost:
		return pi.Post != nil
	case http.MethodPut:
		return pi.Put != nil
	case http.MethodDelete:
		return pi.Delete != nil
	case http.MethodOptions:
		return pi.Options != nil
	case http.MethodHead:
		return pi.Head != nil
	case http.MethodPatch:
		return pi.Patch != nil
	}

	return false
}

// Map returns operations mapped by HTTP method
func (pi PathItem) Map() map[string]*OperationObj {
	result := make(map[string]*OperationObj, 7)
	if pi.Get != nil {
		result[http.MethodGet] = pi.Get
	}
	if pi.Put != nil {
		result[http.MethodPut] = pi.Put
	}
	if pi.Post != nil {
		result[http.MethodPost] = pi.Post
	}
	if pi.Delete != nil {
		result[http.MethodDelete] = pi.Delete
	}
	if pi.Options != nil {
		result[http.MethodOptions] = pi.Options
	}
	if pi.Head != nil {
		result[http.MethodHead] = pi.Head
	}
	if pi.Patch != nil {
		result[http.MethodPatch] = pi.Patch
	}
	return result
}

type securityType string

const (
	// SecurityBasicAuth is a HTTP Basic Authentication security type
	SecurityBasicAuth securityType = "basic"
	// SecurityAPIKey is an API key security type
	SecurityAPIKey securityType = "apiKey"
	// SecurityOAuth2 is an OAuth2 security type
	SecurityOAuth2 securityType = "oauth2"
)

type apiKeyIn string

const (
	// APIKeyInHeader defines API key in header
	APIKeyInHeader apiKeyIn = "header"
	// APIKeyInQuery defines API key in query parameter
	APIKeyInQuery apiKeyIn = "query"
)

type oauthFlow string

const (
	// Oauth2AccessCode is access code Oauth2 flow
	Oauth2AccessCode oauthFlow = "accessCode"
	// Oauth2Application is application Oauth2 flow
	Oauth2Application oauthFlow = "application"
	// Oauth2Implicit is implicit Oauth2 flow
	Oauth2Implicit oauthFlow = "implicit"
	// Oauth2Password is password Oauth2 flow
	Oauth2Password oauthFlow = "password"
)

// SecurityDef holds security definition
type SecurityDef struct {
	Type securityType `json:"type"`

	// apiKey properties
	In   apiKeyIn `json:"in,omitempty"`
	Name string   `json:"name,omitempty"` // Example: X-API-Key

	// oauth2 properties
	Flow             oauthFlow         `json:"flow,omitempty"`
	AuthorizationURL string            `json:"authorizationUrl,omitempty"` // Example: https://example.com/oauth/authorize
	TokenURL         string            `json:"tokenUrl,omitempty"`         // Example: https://example.com/oauth/token
	Scopes           map[string]string `json:"scopes,omitempty"`           // Example: {"read": "Grants read access", "write": "Grants write access"}

	Description string `json:"description,omitempty"`
}

// PathItemInfo some basic information of a path item and operation object
type PathItemInfo struct {
	Path        string
	Method      string
	Title       string
	Description string
	Tag         string
	Deprecated  bool

	// Request holds a sample of request structure, e.g. new(MyRequest)
	Request interface{}

	// Output holds a sample of successful response, e.g. new(MyResponse)
	Response interface{}

	// MIME types of input and output
	Produces []string
	Consumes []string

	Security       []string            // Names of security definitions
	SecurityOAuth2 map[string][]string // Map of names of security definitions to required scopes

	responses              map[int]interface{}
	SuccessfulResponseCode int

	additionalData
}

// RemoveResponse removes response with http status code and returns if it existed
func (p *PathItemInfo) RemoveResponse(statusCode int) bool {
	if nil == p.responses {
		return false
	}
	if _, ok := p.responses[statusCode]; ok {
		delete(p.responses, statusCode)
		return true
	}
	return false
}

// AddResponse adds response with http status code and output structure
func (p *PathItemInfo) AddResponse(statusCode int, output interface{}) *PathItemInfo {
	if nil == p.responses {
		p.responses = make(map[int]interface{}, 1)
	}
	p.responses[statusCode] = output
	return p
}

// AddResponses adds multiple responses with WithStatusCode
func (p *PathItemInfo) AddResponses(responses ...WithStatusCode) {
	if len(responses) == 0 {
		return
	}
	if nil == p.responses {
		p.responses = make(map[int]interface{}, len(responses))
	}
	for _, r := range responses {
		p.responses[r.StatusCode()] = r
	}
}

type description interface {
	Description() string
}

// WithStatusCode is an interface to expose http status code
type WithStatusCode interface {
	StatusCode() int
}

// Enum can be use for sending Enum data that need validate
type Enum struct {
	Enum      []interface{} `json:"enum,omitempty"`
	EnumNames []string      `json:"x-enum-names,omitempty"`
}

// LoadFromField loads enum from field tag: json array or comma-separated string
func (enum *Enum) LoadFromField(field reflect.StructField) {
	type namedEnum interface {
		// NamedEnum return the const-name pair slice
		NamedEnum() ([]interface{}, []string)
	}

	type plainEnum interface {
		Enum() []interface{}
	}

	if e, isEnumer := reflect.Zero(field.Type).Interface().(namedEnum); isEnumer {
		enum.Enum, enum.EnumNames = e.NamedEnum()
	}

	if e, isEnumer := reflect.Zero(field.Type).Interface().(plainEnum); isEnumer {
		enum.Enum = e.Enum()
	}

	if enumTag := field.Tag.Get("enum"); enumTag != "" {
		var e []interface{}
		err := json.Unmarshal([]byte(enumTag), &e)
		if err != nil {
			es := strings.Split(enumTag, ",")
			e = make([]interface{}, len(es))
			for i, s := range es {
				e[i] = s
			}
		}
		enum.Enum = e
	}
}

// SwaggerData holds parameter and schema information for swagger definition
type SwaggerData struct {
	CommonFields
	ParamObj
	SchemaObj
}

// SwaggerDef returns schema object
func (s SwaggerData) SwaggerDef() SwaggerData {
	return s
}

// Schema returns schema object
func (s SwaggerData) Schema() SchemaObj {
	s.SchemaObj.CommonFields = s.CommonFields
	return s.SchemaObj
}

// Param returns parameter object
func (s SwaggerData) Param() ParamObj {
	s.ParamObj.CommonFields = s.CommonFields
	return s.ParamObj
}

// CommonFields keeps fields shared between ParamObj and SchemaObj
type CommonFields struct {
	Title       string      `json:"title,omitempty"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Type        string      `json:"type,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	Format      string      `json:"format,omitempty"`

	MultipleOf float64  `json:"multipleOf,omitempty"`
	Maximum    *float64 `json:"maximum,omitempty"`
	Minimum    *float64 `json:"minimum,omitempty"`

	MaxLength     *int64 `json:"maxLength,omitempty"`
	MinLength     *int64 `json:"minLength,omitempty"`
	MaxItems      *int64 `json:"maxItems,omitempty"`
	MinItems      *int64 `json:"minItems,omitempty"`
	MaxProperties *int64 `json:"maxProperties,omitempty"`
	MinProperties *int64 `json:"minProperties,omitempty"`

	ExclusiveMaximum bool `json:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum bool `json:"exclusiveMinimum,omitempty"`
	UniqueItems      bool `json:"uniqueItems,omitempty"`

	Enum
}

// OperationObj describes a single API operation on a path
// see http://swagger.io/specification/#operationObject
type OperationObj struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary"`     // like a title, a short summary of what the operation does (120 chars)
	Description string                `json:"description"` // A verbose explanation of the operation behavior
	Parameters  []ParamObj            `json:"parameters,omitempty"`
	Produces    []string              `json:"produces,omitempty"`
	Consumes    []string              `json:"consumes,omitempty"`
	Responses   Responses             `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
	additionalData
}

// MarshalJSON marshal OperationObj with additionalData inlined
func (o OperationObj) MarshalJSON() ([]byte, error) {
	type i OperationObj
	return o.marshalJSONWithStruct(i(o))
}

// ParamObj describes a single operation parameter
// see http://swagger.io/specification/#parameterObject
type ParamObj struct {
	CommonFields
	Name   string        `json:"name,omitempty"`
	In     string        `json:"in,omitempty"`     // Possible values are "query", "header", "path", "formData" or "body"
	Items  *ParamItemObj `json:"items,omitempty"`  // Required if type is "array"
	Schema *SchemaObj    `json:"schema,omitempty"` // Required if type is "body"

	// CollectionFormat defines serialization:
	// "multi" is valid only for parameters in "query" or "formData": foo=value&foo=another_value
	// "csv" is comma-separated values: "foo,bar,baz"
	// "ssv" is space-separated values: "foo bar baz"
	// "tsv" is tab-separated values: "foo\tbar\tbaz"
	// "pipes" is pipe-separated values: "foo|bar|baz"
	CollectionFormat string `json:"collectionFormat,omitempty"`

	Required bool `json:"required,omitempty"`
	additionalData
}

func jsonRecode(v interface{}) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var decoded interface{}
	err = json.Unmarshal(jsonBytes, &decoded)
	if err != nil {
		return nil, err
	}

	if m, ok := decoded.(map[string]interface{}); ok {
		return m, nil
	}

	return nil, errors.New(`invalid json, map expected`)
}

// MarshalJSON marshal ParamObj with additionalData inlined
func (o ParamObj) MarshalJSON() ([]byte, error) {
	type i ParamObj
	return o.marshalJSONWithStruct(i(o))
}

// MarshalJSON marshal SchemaObj with additionalData inlined
func (o SchemaObj) MarshalJSON() ([]byte, error) {
	type i SchemaObj
	return o.marshalJSONWithStruct(i(o))
}

// ParamItemObj describes an property object, in param object or property of definition
// see http://swagger.io/specification/#itemsObject
type ParamItemObj struct {
	CommonFields
	Items            *ParamItemObj `json:"items,omitempty"`            // Required if type is "array"
	CollectionFormat string        `json:"collectionFormat,omitempty"` // "multi" - this is valid only for parameters in "query" or "formData"
}

// Responses list of response object
type Responses map[int]ResponseObj

// ResponseObj describes a single response from an API Operation
type ResponseObj struct {
	Ref         string      `json:"$ref,omitempty"`
	Description string      `json:"description,omitempty"`
	Schema      *SchemaObj  `json:"schema,omitempty"`
	Headers     interface{} `json:"headers,omitempty"`
	Examples    interface{} `json:"examples,omitempty"`
}

// SchemaObj describes a schema for json format
type SchemaObj struct {
	CommonFields
	Ref                  string               `json:"$ref,omitempty"`
	Items                *SchemaObj           `json:"items,omitempty"`                // if type is array
	AdditionalProperties *SchemaObj           `json:"additionalProperties,omitempty"` // if type is object (map[])
	Properties           map[string]SchemaObj `json:"properties,omitempty"`           // if type is object
	Example              interface{}          `json:"example,omitempty"`
	Nullable             bool                 `json:"x-nullable,omitempty"`
	TypeName             string               `json:"-"` // for internal using, passing typeName
	GoType               string               `json:"x-go-type,omitempty"`
	GoPropertyNames      map[string]string    `json:"x-go-property-names,omitempty"`
	GoPropertyTypes      map[string]string    `json:"x-go-property-types,omitempty"`
	additionalData
}

// NewSchemaObj Constructor function for SchemaObj struct type
func NewSchemaObj(jsonType, typeName string) (so *SchemaObj) {
	so = &SchemaObj{}
	so.Type = jsonType
	so.TypeName = typeName

	if typeName != "" {
		so.Ref = refDefinitionPrefix + typeName
	}
	return
}

// Checks whether current SchemaObj is "empty". A schema object is considered "empty" if it is an object without visible
// (exported) properties, an array without elements, or in other cases when it has neither regular nor additional
// properties, and format is not specified. SwaggerData objects that describe common types ("string", "integer", "boolean" etc.)
// are always considered non-empty. Same is true for "schema reference objects" (objects that have a non-empty Ref field).
func (o *SchemaObj) isEmpty() bool {
	if isCommonName(o.TypeName) || o.Ref != "" {
		return false
	}

	switch o.Type {
	case "object":
		return len(o.Properties) == 0
	case "array":
		return o.Items == nil
	default:
		return len(o.Properties) == 0 && o.AdditionalProperties == nil && o.Format == ""
	}
}

// Export returns a "schema reference object" corresponding to this schema object. A "schema reference object" is an abridged
// version of the original SchemaObj, having only two non-empty fields: Ref and TypeName. "SwaggerData reference objects"
// are used to refer original schema objects from other schemas.
func (o SchemaObj) Export() SchemaObj {
	return SchemaObj{
		Ref:      o.Ref,
		TypeName: o.TypeName,
	}
}

type additionalData struct {
	data map[string]interface{}
}

// AddExtendedField add field to additional data map
func (ad *additionalData) AddExtendedField(name string, value interface{}) {
	if ad.data == nil {
		ad.data = make(map[string]interface{})
	}

	ad.data[name] = value
}

func (ad additionalData) marshalJSONWithStruct(i interface{}) ([]byte, error) {
	result, err := json.Marshal(i)
	if err != nil {
		return result, err
	}

	if len(ad.data) == 0 {
		return result, nil
	}

	dataJSON, err := json.Marshal(ad.data)
	if err != nil {
		return dataJSON, err
	}

	if string(result) == "{}" {
		return dataJSON, nil
	}

	result = append(result[:len(result)-1], ',')
	result = append(result, dataJSON[1:]...)

	return result, nil
}
