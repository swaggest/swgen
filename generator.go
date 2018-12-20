package swgen

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/swaggest/swgen/refl"
)

// Generator create swagger document
type Generator struct {
	doc  Document
	host string // address of api in host:port format

	corsMu           sync.RWMutex // mutex for CORS public API
	corsEnabled      bool         // allow cross-origin HTTP request
	corsAllowHeaders []string

	definitionAlloc  map[string]string       // index of allocated TypeNames
	definitions      defMap                  // list of all definition objects
	defQueue         map[string]reflect.Type // queue of reflect.Type objects waiting for analysis
	paths            map[string]PathItem     // list all of paths object
	typesMap         map[string]interface{}
	defaultResponses map[int]interface{}

	indentJSON       bool
	reflectGoTypes   bool
	addPackagePrefix bool

	mu sync.Mutex // mutex for Generator's public API
}

type defMap map[string]SchemaObj

func (m *defMap) GenDefinitions() (result map[string]SchemaObj) {
	if m == nil {
		return nil
	}

	result = make(map[string]SchemaObj)
	for t, typeDef := range *m {
		typeDef.Ref = "" // first (top) level Swagger definitions are never references
		if _, ok := result[typeDef.TypeName]; ok {
			typeName := t
			result[typeName] = typeDef
		} else {
			result[typeDef.TypeName] = typeDef
		}
	}
	return
}

// NewGenerator create a new Generator
func NewGenerator() *Generator {
	g := &Generator{}

	g.definitions = make(map[string]SchemaObj)
	g.definitionAlloc = make(map[string]string)

	g.defQueue = make(map[string]reflect.Type)
	g.paths = make(map[string]PathItem) // list all of paths object
	g.typesMap = make(map[string]interface{})

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

// SetHost set host info for swagger specification
func (g *Generator) SetHost(host string) *Generator {
	g.mu.Lock()
	g.host = host
	g.mu.Unlock()
	return g
}

// SetBasePath set host info for swagger specification
func (g *Generator) SetBasePath(basePath string) *Generator {
	basePath = "/" + strings.Trim(basePath, "/")
	g.mu.Lock()
	g.doc.BasePath = basePath
	g.mu.Unlock()
	return g
}

// SetContact set contact information for API
func (g *Generator) SetContact(name, url, email string) *Generator {
	ct := ContactObj{
		Name:  name,
		URL:   url,
		Email: email,
	}

	g.mu.Lock()
	g.doc.Info.Contact = ct
	g.mu.Unlock()
	return g
}

// SetInfo set information about API
func (g *Generator) SetInfo(title, description, term, version string) *Generator {
	info := InfoObj{
		Title:          title,
		Description:    description,
		TermsOfService: term,
		Version:        version,
	}

	g.mu.Lock()
	g.doc.Info = info
	g.mu.Unlock()
	return g
}

// SetLicense set license information for API
func (g *Generator) SetLicense(name, url string) *Generator {
	ls := LicenseObj{
		Name: name,
		URL:  url,
	}

	g.mu.Lock()
	g.doc.Info.License = ls
	g.mu.Unlock()
	return g
}

// AddExtendedField add vendor extension field to document
func (g *Generator) AddExtendedField(name string, value interface{}) *Generator {
	g.mu.Lock()
	g.doc.AddExtendedField(name, value)
	g.mu.Unlock()
	return g
}

// AddSecurityDefinition adds shared security definition to document
func (g *Generator) AddSecurityDefinition(name string, def SecurityDef) *Generator {
	g.mu.Lock()
	g.doc.SecurityDefinitions[name] = def
	g.mu.Unlock()
	return g
}

// AddTypeMap adds mapping relation to treat values of same type as source as they were of same type as destination
func (g *Generator) AddTypeMap(source interface{}, destination interface{}) *Generator {
	g.mu.Lock()

	goTypeName := refl.GoType(refl.DeepIndirect(reflect.TypeOf(source)))
	g.typesMap[goTypeName] = destination
	g.mu.Unlock()
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
