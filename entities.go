package swgen

import (
	"encoding/json"
	"errors"
	"strings"
)

// ServiceType data type for type of your service
type ServiceType string

const (
	// ServiceTypeRest define service type for RESTful service
	ServiceTypeRest ServiceType = "rest"
	// ServiceTypeJSONRPC define service type for JSON-RPC service
	ServiceTypeJSONRPC ServiceType = "json-rpc"
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
	Definitions         map[string]SchemaObj   `json:"definitions"`                   // An object to hold data types produced and consumed by operations
	SecurityDefinitions map[string]SecurityDef `json:"securityDefinitions,omitempty"` // An object to hold available security mechanisms
	additionalData
}

type _Document Document

// MarshalJSON marshal Document with additionalData inlined
func (s Document) MarshalJSON() ([]byte, error) {
	return s.marshalJSONWithStruct(_Document(s))
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
	case "GET":
		return pi.Get != nil
	case "POST":
		return pi.Post != nil
	case "PUT":
		return pi.Put != nil
	case "DELETE":
		return pi.Delete != nil
	case "OPTIONS":
		return pi.Options != nil
	case "HEAD":
		return pi.Head != nil
	case "PATCH":
		return pi.Patch != nil
	}

	return false
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

	Security       []string            // Names of security definitions
	SecurityOAuth2 map[string][]string // Map of names of security definitions to required scopes

	responses              map[int]interface{}
	SuccessfulResponseCode int

	additionalData

	// JSON SwaggerData reflection of Swagger entities
	requestParamsSchemaJson []byte
	requestBodySchemaJson   []byte
	responseSchemaJson      []byte
}

func (p *PathItemInfo) AddResponse(statusCode int, output interface{}) *PathItemInfo {
	if nil == p.responses {
		p.responses = make(map[int]interface{}, 1)
	}
	p.responses[statusCode] = output
	return p
}

func (p *PathItemInfo) AddResponses(responses ...PathItemResponse) {
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

type PathItemResponse interface {
	StatusCode() int
}

// Enum can be use for sending Enum data that need validate
type Enum struct {
	Enum      []interface{} `json:"enum,omitempty"`
	EnumNames []string      `json:"x-enum-names,omitempty"`
}

type namedEnum interface {
	// NamedEnum return the const-name pair slice
	NamedEnum() ([]interface{}, []string)
}

type enum interface {
	Enum() []interface{}
}

type SwaggerData struct {
	shared
	ParamObj
	SchemaObj
}

// Schema returns schema object
func (s SwaggerData) SwaggerDef() SwaggerData {
	return s
}

// Schema returns schema object
func (s SwaggerData) Schema() SchemaObj {
	s.SchemaObj.shared = s.shared
	return s.SchemaObj
}

// Param returns parameter object
func (s SwaggerData) Param() ParamObj {
	s.ParamObj.shared = s.shared
	return s.ParamObj
}

type shared struct {
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Type        string      `json:"type,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	Format      string      `json:"format,omitempty"`
	Enum
}

// OperationObj describes a single API operation on a path
// see http://swagger.io/specification/#operationObject
type OperationObj struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary"`     // like a title, a short summary of what the operation does (120 chars)
	Description string                `json:"description"` // A verbose explanation of the operation behavior
	Parameters  []ParamObj            `json:"parameters,omitempty"`
	Responses   Responses             `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
	additionalData
}

type _OperationObj OperationObj

// MarshalJSON marshal OperationObj with additionalData inlined
func (o OperationObj) MarshalJSON() ([]byte, error) {
	return o.marshalJSONWithStruct(_OperationObj(o))
}

// ParamObj describes a single operation parameter
// see http://swagger.io/specification/#parameterObject
type ParamObj struct {
	shared
	Name             string        `json:"name,omitempty"`
	In               string        `json:"in,omitempty"`               // Possible values are "query", "header", "path", "formData" or "body"
	Items            *ParamItemObj `json:"items,omitempty"`            // Required if type is "array"
	Schema           *SchemaObj    `json:"schema,omitempty"`           // Required if type is "body"
	CollectionFormat string        `json:"collectionFormat,omitempty"` // "multi" - this is valid only for parameters in "query" or "formData"
	Required         bool          `json:"required,omitempty"`
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

type _ParamObj ParamObj

// MarshalJSON marshal OperationObj with additionalData inlined
func (o ParamObj) MarshalJSON() ([]byte, error) {
	return o.marshalJSONWithStruct(_ParamObj(o))
}

// ParamItemObj describes an property object, in param object or property of definition
// see http://swagger.io/specification/#itemsObject
type ParamItemObj struct {
	Ref              string        `json:"$ref,omitempty"`
	Type             string        `json:"type"`
	Format           string        `json:"format,omitempty"`
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
	shared
	Ref                  string               `json:"$ref,omitempty"`
	Title                string               `json:"title,omitempty"`
	Items                *SchemaObj           `json:"items,omitempty"`                // if type is array
	AdditionalProperties *SchemaObj           `json:"additionalProperties,omitempty"` // if type is object (map[])
	Properties           map[string]SchemaObj `json:"properties,omitempty"`           // if type is object
	Example              interface{}          `json:"example,omitempty"`
	Nullable             bool                 `json:"nullable,omitempty"`
	TypeName             string               `json:"-"` // for internal using, passing typeName
	GoType               string               `json:"x-go-type,omitempty"`
	GoPropertyNames      map[string]string    `json:"x-go-property-names,omitempty"`
	GoPropertyTypes      map[string]string    `json:"x-go-property-types,omitempty"`
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
func (so *SchemaObj) isEmpty() bool {
	if isCommonName(so.TypeName) || so.Ref != "" {
		return false
	}

	switch so.Type {
	case "object":
		return len(so.Properties) == 0
	case "array":
		return so.Items == nil
	default:
		return len(so.Properties) == 0 && so.AdditionalProperties == nil && so.Format == ""
	}
}

// Export returns a "schema reference object" corresponding to this schema object. A "schema reference object" is an abridged
// version of the original SchemaObj, having only two non-empty fields: Ref and TypeName. "SwaggerData reference objects"
// are used to refer original schema objects from other schemas.
func (so SchemaObj) Export() SchemaObj {
	return SchemaObj{
		Ref:      so.Ref,
		TypeName: so.TypeName,
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
