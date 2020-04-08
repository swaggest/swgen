package swgen

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/refl"
)

// Generator create swagger document
type Generator struct {
	oas3Proxy *openapi3.Reflector

	doc  Document
	host string // address of api in host:port format

	corsMu           sync.RWMutex // mutex for CORS public API
	corsEnabled      bool         // allow cross-origin HTTP request
	corsAllowHeaders []string

	definitionAlloc  map[string]refl.TypeString       // index of allocated TypeNames
	definitions      defMap                           // list of all definition objects
	defQueue         map[refl.TypeString]reflect.Type // queue of reflect.Type objects waiting for analysis
	paths            map[string]PathItem              // list all of paths object
	typesMap         map[refl.TypeString]interface{}
	defaultResponses map[int]interface{}

	indentJSON            bool
	reflectGoTypes        bool
	addPackagePrefix      bool
	capitalizeDefinitions bool

	mu sync.Mutex // mutex for Generator's public API
}

type defMap map[refl.TypeString]SchemaObj

func (m *defMap) GenDefinitions() (result map[string]SchemaObj) {
	if m == nil {
		return nil
	}

	result = make(map[string]SchemaObj)
	for t, typeDef := range *m {
		typeDef.Ref = "" // first (top) level Swagger definitions are never references
		if _, ok := result[typeDef.TypeName]; ok {
			typeName := t
			result[string(typeName)] = typeDef
		} else {
			result[typeDef.TypeName] = typeDef
		}
	}
	return
}

// NewGenerator create a new Generator
func NewGenerator() *Generator {
	g := &Generator{}

	g.definitions = make(map[refl.TypeString]SchemaObj)
	g.definitionAlloc = make(map[string]refl.TypeString)

	g.defQueue = make(map[refl.TypeString]reflect.Type)
	g.paths = make(map[string]PathItem) // list all of paths object
	g.typesMap = make(map[refl.TypeString]interface{})

	g.doc.Schemes = []string{"http", "https"}
	g.doc.Paths = make(map[string]PathItem)
	g.doc.Definitions = make(map[string]SchemaObj)
	g.doc.SecurityDefinitions = make(map[string]SecurityDef)
	g.doc.Version = "2.0"
	g.doc.BasePath = "/"

	// set default Access-Control-Allow-Headers of swagger.json
	g.corsAllowHeaders = []string{"Content-Type", "api_key", "Authorization"}

	return g
}

// IndentJSON controls JSON indentation
func (g *Generator) IndentJSON(enabled bool) *Generator {
	g.mu.Lock()
	g.indentJSON = enabled
	g.mu.Unlock()
	return g
}

// ReflectGoTypes controls JSON indentation
func (g *Generator) ReflectGoTypes(enabled bool) *Generator {
	g.mu.Lock()
	g.reflectGoTypes = enabled
	g.mu.Unlock()
	return g
}

// AddPackagePrefix controls prefixing definition name with package.
// With option enabled type `some/package.Entity` will have "PackageEntity" key in definitions, "Entity" otherwise.
func (g *Generator) AddPackagePrefix(enabled bool) *Generator {
	g.mu.Lock()
	g.addPackagePrefix = enabled
	g.mu.Unlock()
	return g
}

// CapitalizeDefinitions enables first char capitalization for definition names.
// With option enabled type `some/package.entity` will have "Entity" key in definitions, "entity" otherwise.
func (g *Generator) CapitalizeDefinitions(enabled bool) *Generator {
	g.mu.Lock()
	g.capitalizeDefinitions = enabled
	g.mu.Unlock()
	return g
}

// AddDefaultResponse adds http code and response structure that will be applied to all operations
func (g *Generator) AddDefaultResponse(httpCode int, response interface{}) {
	if g.defaultResponses == nil {
		g.defaultResponses = make(map[int]interface{})
	}
	g.defaultResponses[httpCode] = response
}

// EnableCORS enable HTTP handler support CORS
func (g *Generator) EnableCORS(b bool, allowHeaders ...string) *Generator {
	g.corsMu.Lock()
	g.corsEnabled = b
	if len(allowHeaders) != 0 {
		g.corsAllowHeaders = append(g.corsAllowHeaders, allowHeaders...)
	}
	g.corsMu.Unlock()
	return g
}

func (g *Generator) writeCORSHeaders(w http.ResponseWriter) {
	g.corsMu.RLock()
	defer g.corsMu.RUnlock()

	if !g.corsEnabled {
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(g.corsAllowHeaders, ", "))
}

func (g *Generator) updateServer() {
	if len(g.oas3Proxy.SpecEns().Servers) == 0 {
		g.oas3Proxy.SpecEns().Servers = append(g.oas3Proxy.SpecEns().Servers, openapi3.Server{})
	}

	if g.host != "" {
		g.oas3Proxy.SpecEns().Servers[0].URL = "http://" + g.host + g.doc.BasePath
	} else {
		g.oas3Proxy.SpecEns().Servers[0].URL = g.doc.BasePath
	}
}

// SetHost set host info for swagger specification
func (g *Generator) SetHost(host string) *Generator {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.host = host

	if g.oas3Proxy != nil {
		g.updateServer()
	}

	return g
}

// SetBasePath set host info for swagger specification
func (g *Generator) SetBasePath(basePath string) *Generator {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.doc.BasePath = "/" + strings.Trim(basePath, "/")

	if g.oas3Proxy != nil {
		g.updateServer()
	}

	return g
}

// SetContact set contact information for API
func (g *Generator) SetContact(name, url, email string) *Generator {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.doc.Info.Contact = ContactObj{
		Name:  name,
		URL:   url,
		Email: email,
	}

	if g.oas3Proxy != nil {
		if name != "" {
			g.oas3Proxy.SpecEns().Info.ContactEns().Name = &name
		}
		if url != "" {
			g.oas3Proxy.SpecEns().Info.ContactEns().URL = &url
		}
		if email != "" {
			g.oas3Proxy.SpecEns().Info.ContactEns().Email = &email
		}
	}

	return g
}

// SetOAS3Proxy enables OpenAPI3 spec collection with provided reflector.
func (g *Generator) SetOAS3Proxy(oas3Proxy *openapi3.Reflector) {
	g.mu.Lock()
	defer g.mu.Unlock()

	oas3Proxy.DefaultOptions = append(oas3Proxy.DefaultOptions,
		jsonschema.InterceptType(JSONSchemaInterceptType),
	)

	g.oas3Proxy = oas3Proxy
}

// SetInfo set information about API
func (g *Generator) SetInfo(title, description, term, version string) *Generator {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.doc.Info = InfoObj{
		Title:          title,
		Description:    description,
		TermsOfService: term,
		Version:        version,
	}

	if g.oas3Proxy != nil {
		g.oas3Proxy.Spec.Info.Title = title
		g.oas3Proxy.Spec.Info.Description = &description
		g.oas3Proxy.Spec.Info.Version = version
		g.oas3Proxy.Spec.Info.TermsOfService = &term
	}

	return g
}

// SetLicense set license information for API
func (g *Generator) SetLicense(name, url string) *Generator {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.doc.Info.License = LicenseObj{
		Name: name,
		URL:  url,
	}

	if g.oas3Proxy != nil {
		g.oas3Proxy.SpecEns().Info.LicenseEns().Name = name

		if url != "" {
			g.oas3Proxy.SpecEns().Info.LicenseEns().URL = &url
		}
	}

	return g
}

// AddExtendedField add vendor extension field to document
func (g *Generator) AddExtendedField(name string, value interface{}) *Generator {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.doc.AddExtendedField(name, value)

	if g.oas3Proxy != nil {
		g.oas3Proxy.SpecEns().WithMapOfAnythingItem(name, value)
	}

	return g
}

// AddSecurityDefinition adds shared security definition to document
func (g *Generator) AddSecurityDefinition(name string, def SecurityDef) *Generator {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.doc.SecurityDefinitions[name] = def

	if g.oas3Proxy != nil {
		ss := openapi3.SecuritySchemeOrRef{}
		sss := ss.SecuritySchemeEns()

		switch def.Type {
		case SecurityBasicAuth:
			sss.HTTPSecuritySchemeEns().Scheme = "basic"
			if def.Description != "" {
				sss.HTTPSecuritySchemeEns().Description = &def.Description
			}
		case SecurityAPIKey:
			sss.APIKeySecuritySchemeEns().Name = def.Name
			sss.APIKeySecuritySchemeEns().In = openapi3.APIKeySecuritySchemeIn(def.In)
			if def.Description != "" {
				sss.APIKeySecuritySchemeEns().Description = &def.Description
			}
		case SecurityOAuth2:
			switch def.Flow {
			case Oauth2Implicit:
				sss.OAuth2SecuritySchemeEns().Flows.ImplicitEns().AuthorizationURL = def.AuthorizationURL
				sss.OAuth2SecuritySchemeEns().Flows.ImplicitEns().Scopes = def.Scopes
			case Oauth2Password:
				sss.OAuth2SecuritySchemeEns().Flows.PasswordEns().TokenURL = def.TokenURL
				sss.OAuth2SecuritySchemeEns().Flows.PasswordEns().Scopes = def.Scopes
			}
			if def.Description != "" {
				sss.APIKeySecuritySchemeEns().Description = &def.Description
			}
		}

		g.oas3Proxy.SpecEns().ComponentsEns().SecuritySchemesEns().
			WithMapOfSecuritySchemeOrRefValuesItem(name, ss)
	}

	return g
}

// AddTypeMap adds mapping relation to treat values of same type as source as they were of same type as destination
func (g *Generator) AddTypeMap(source interface{}, destination interface{}) *Generator {
	g.mu.Lock()
	defer g.mu.Unlock()

	goTypeName := refl.GoType(refl.DeepIndirect(reflect.TypeOf(source)))
	g.typesMap[goTypeName] = destination

	if g.oas3Proxy != nil {
		g.oas3Proxy.AddTypeMapping(source, destination)
	}

	return g
}

func (g *Generator) getMappedType(t reflect.Type) (dst interface{}, found bool) {
	goTypeName := refl.GoType(refl.DeepIndirect(t))
	dst, found = g.typesMap[goTypeName]
	return
}

// genDocument returns document specification in JSON string (in []byte)
func (g *Generator) genDocument(host *string) ([]byte, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	var (
		data []byte
		err  error
	)

	// ensure that all definition in queue is parsed before generating
	g.parseDefInQueue()
	g.doc.Definitions = g.definitions.GenDefinitions()
	if g.host != "" || host == nil {
		g.doc.Host = g.host
	} else {
		g.doc.Host = *host
	}
	g.doc.Paths = make(map[string]PathItem)

	for path, item := range g.paths {
		g.doc.Paths[path] = item
	}

	if g.indentJSON {
		data, err = json.MarshalIndent(g.doc, "", "  ")
	} else {
		data, err = json.Marshal(g.doc)
	}

	return data, err
}

// GenDocument returns document specification in JSON string (in []byte)
func (g *Generator) GenDocument() ([]byte, error) {
	// pass nil here to set host as g.host
	return g.genDocument(nil)
}

// ServeHTTP implements http.Handler to server swagger.json document
func (g *Generator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := g.genDocument(&r.URL.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")

	g.writeCORSHeaders(w)

	_, _ = w.Write(data)
}

// Document is an accessor to generated document
func (g *Generator) Document() Document {
	return g.doc
}
