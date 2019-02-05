package swgen

import (
	"encoding"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/swaggest/swgen/refl"
)

const (
	refDefinitionPrefix = "#/definitions/"
)

var (
	typeOfJSONRawMsg      = reflect.TypeOf((*json.RawMessage)(nil)).Elem()
	typeOfTime            = reflect.TypeOf((*time.Time)(nil)).Elem()
	typeOfTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// SchemaDefinition allows to return custom definitions
type SchemaDefinition interface {
	SwaggerDef() SwaggerData
}

// SchemaDefinitionFunc is a helper to return custom definitions
type SchemaDefinitionFunc func() SwaggerData

// SwaggerDef returns custom definition
func (f SchemaDefinitionFunc) SwaggerDef() SwaggerData {
	return f()
}

func (g *Generator) makeNameForType(t reflect.Type, baseTypeName string) string {
	goTypeName := refl.GoType(t)
	if g.capitalizeDefinitions {
		baseTypeName = strings.Title(baseTypeName)
	}

	for typeName, allocatedGoTypeName := range g.definitionAlloc {
		if goTypeName == allocatedGoTypeName {
			return typeName
		}
	}

	pkgPath := t.PkgPath()

	if g.addPackagePrefix && pkgPath != "" {
		pref := strings.Title(path.Base(pkgPath))
		baseTypeName = pref + baseTypeName
		pkgPath = path.Dir(pkgPath)
	}

	allocatedType, isAllocated := g.definitionAlloc[baseTypeName]
	if isAllocated && allocatedType != goTypeName {
		typeIndex := 2
		pref := strings.Title(path.Base(pkgPath))
		for {
			typeName := ""
			if pkgPath != "" {
				typeName = pref + baseTypeName
			} else {
				typeName = fmt.Sprintf("%sType%d", baseTypeName, typeIndex)
				typeIndex++
			}
			allocatedType, isAllocated := g.definitionAlloc[typeName]

			if !isAllocated || allocatedType == goTypeName {
				baseTypeName = typeName
				break
			}
			typeIndex++
			pref = strings.Title(path.Base(pkgPath)) + pref
			pkgPath = path.Dir(pkgPath)
		}
	}
	g.definitionAlloc[baseTypeName] = goTypeName
	return baseTypeName
}

func (g *Generator) addDefinition(t reflect.Type, typeDef *SchemaObj) {
	if typeDef.TypeName == "" {
		return // there should be no anonymous definitions in Swagger JSON
	}
	goTypeName := refl.GoType(t)
	if _, ok := g.definitions[goTypeName]; ok { // skip existing definition
		return
	}
	typeDef.TypeName = g.makeNameForType(t, typeDef.TypeName)
	g.definitions[refl.GoType(t)] = *typeDef
}

func (g *Generator) defExists(t reflect.Type) bool {
	_, found := g.getDefinition(t)
	return found
}

func (g *Generator) addToDefQueue(t reflect.Type) {
	g.defQueue[refl.GoType(t)] = t
}

func (g *Generator) defInQueue(t reflect.Type) (found bool) {
	_, found = g.defQueue[refl.GoType(t)]
	return
}

func (g *Generator) getDefinition(t reflect.Type) (typeDef SchemaObj, found bool) {
	typeDef, found = g.definitions[refl.GoType(t)]
	if !found && t.Kind() == reflect.Ptr {
		typeDef, found = g.definitions[refl.GoType(t.Elem())]
	}
	return
}

func (g *Generator) deleteDefinition(t reflect.Type) {
	delete(g.definitions, refl.GoType(t))
}

//
// Parse swagger schema object
// see http://swagger.io/specification/#schemaObject
//

// ResetDefinitions will remove all exists definitions and init again
func (g *Generator) ResetDefinitions() {
	g.definitions = make(defMap)
	g.defQueue = make(map[string]reflect.Type)
}

// ParseDefinition create a DefObj from input object, it should be a non-nil pointer to anything
// it reuse schema/json tag for property name.
func (g *Generator) ParseDefinition(i interface{}) SchemaObj {
	var (
		typeName string
		typeDef  SchemaObj
		t        = reflect.TypeOf(i)
		v        = reflect.ValueOf(i)
		ot       = t // original type
	)

	goTypeName := refl.GoType(t)
	_ = goTypeName

	if mappedTo, ok := g.getMappedType(t); ok {
		typeName = t.Name()
		t = reflect.TypeOf(mappedTo)
		v = reflect.ValueOf(mappedTo)
	}

	if definition, ok := i.(SchemaDefinition); ok {
		typeDef = definition.SwaggerDef().Schema()
		if typeDef.TypeName == "" {
			typeName = t.Name()
		} else {
			typeName = typeDef.TypeName
		}
		typeDef.TypeName = typeName
		if def, ok := g.getDefinition(t); ok {
			return SchemaObj{Ref: refDefinitionPrefix + def.TypeName, TypeName: def.TypeName}
		}
		defer g.parseDefInQueue()
		if g.reflectGoTypes {
			typeDef.GoType = refl.GoType(t)
		}
		g.addDefinition(t, &typeDef)

		return SchemaObj{Ref: refDefinitionPrefix + typeDef.TypeName, TypeName: typeDef.TypeName}
	}

	t = refl.DeepIndirect(t)

	name := refl.GoType(t) // todo remove
	_ = name

	switch t.Kind() {
	case reflect.Struct:
		if typeDef, found := g.getDefinition(t); found {
			return typeDef.Export()
		}

		typeDef = *NewSchemaObj("object", g.reflectTypeReliableName(t))
		if typeDef.TypeName == "" {
			typeDef.TypeName = typeName
		}
		typeDef.TypeName = g.makeNameForType(t, typeDef.TypeName)
		typeDef.Ref = refDefinitionPrefix + typeDef.TypeName
		typeDef.Properties = g.parseDefinitionProperties(v, &typeDef)

	case reflect.Slice, reflect.Array:
		elemType := refl.DeepIndirect(t.Elem())

		if typeDef, found := g.getDefinition(t); found {
			return typeDef.Export()
		}

		var itemSchema SchemaObj
		if elemType.Kind() != reflect.Struct || (elemType.Kind() == reflect.Struct && elemType.Name() != "") {
			itemSchema = g.genSchemaForType(elemType)
		} else {
			itemSchema = *NewSchemaObj("object", elemType.Name())
			itemSchema.Properties = g.parseDefinitionProperties(v.Elem(), &itemSchema)
		}

		typeDef = *NewSchemaObj("array", t.Name())
		typeDef.Items = &itemSchema
		if typeDef.TypeName == "" {
			typeDef.TypeName = typeName
		}
	case reflect.Map:
		elemType := refl.DeepIndirect(t.Elem())

		if typeDef, found := g.getDefinition(t); found {
			return typeDef.Export()
		}

		typeDef = *NewSchemaObj("object", t.Name())
		itemDef := g.genSchemaForType(elemType)
		typeDef.AdditionalProperties = &itemDef
		if typeDef.TypeName == "" {
			typeDef.TypeName = typeName
		}
	default:
		typeDef = g.genSchemaForType(t)
		typeDef.TypeName = typeDef.Type
		return typeDef
	}

	defer g.parseDefInQueue()

	if g.reflectGoTypes {
		typeDef.GoType = refl.GoType(ot)
	}

	if typeDef.TypeName != "" { // non-anonymous types should be added to definitions map and returned "in-place" as references
		typeDef.TypeName = g.makeNameForType(t, typeDef.TypeName)
		g.addDefinition(t, &typeDef)
		return typeDef.Export()
	}
	return typeDef // anonymous types are not added to definitions map; instead, they are returned "in-place" in full form
}

func (g *Generator) parseDefinitionProperties(v reflect.Value, parent *SchemaObj) map[string]SchemaObj {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	properties := make(map[string]SchemaObj, t.NumField())
	if g.reflectGoTypes && parent.GoPropertyNames == nil {
		parent.GoPropertyNames = make(map[string]string, t.NumField())
		parent.GoPropertyTypes = make(map[string]string, t.NumField())
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		oft := field.Type

		var tag = field.Tag.Get("json")

		if tag == "-" {
			continue
		}

		if tag == "" && field.Anonymous {
			fieldProperties := g.parseDefinitionProperties(v.Field(i), parent)
			for propertyName, property := range fieldProperties {
				properties[propertyName] = property
			}
			continue
		}

		// don't check if it's omitted
		if tag == "" {
			continue
		}

		propName := strings.Split(tag, ",")[0]
		var (
			obj      SchemaObj
			objReady bool
		)

		if mapped, found := g.getMappedType(field.Type); found {
			if def, ok := mapped.(SchemaDefinition); ok {
				obj = def.SwaggerDef().Schema()
				objReady = true
			} else {
				field.Type = reflect.TypeOf(mapped)
			}
		}

		if !objReady {
			if field.Type.Kind() == reflect.Interface && v.Field(i).Elem().IsValid() {
				obj = g.genSchemaForType(v.Field(i).Elem().Type())
			} else {
				typeName := refl.GoType(field.Type)
				_ = typeName
				obj = g.genSchemaForType(field.Type)
			}
		}

		obj.Enum.LoadFromField(field)

		if formatTag := field.Tag.Get("format"); formatTag != "" {
			obj.Format = formatTag
		}

		if description := field.Tag.Get("description"); description != "" {
			obj.Description = description
		}

		if defaultTag := field.Tag.Get("default"); defaultTag != "" {
			if defaultValue, err := g.caseDefaultValue(field.Type, defaultTag); err == nil {
				obj.Default = defaultValue
			}
		}
		if g.reflectGoTypes {
			if obj.Ref == "" {
				obj.GoType = refl.GoType(oft)
			}
			parent.GoPropertyNames[propName] = field.Name
			parent.GoPropertyTypes[propName] = refl.GoType(oft)
		}

		readSharedTags(field.Tag, &obj.CommonFields)

		properties[propName] = obj
	}

	return properties
}

func (g *Generator) caseDefaultValue(t reflect.Type, defaultValue string) (interface{}, error) {
	t = refl.DeepIndirect(t)
	kind := t.Kind()

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.ParseInt(defaultValue, 10, 64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.ParseUint(defaultValue, 10, 64)
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(defaultValue, 64)
	case reflect.String:
		return defaultValue, nil
	case reflect.Bool:
		return strconv.ParseBool(defaultValue)
	default:
		instance := reflect.New(t).Interface()
		if err := json.Unmarshal([]byte(defaultValue), instance); err != nil {
			return nil, err
		}
		return reflect.Indirect(reflect.ValueOf(instance)).Interface(), nil
	}
}

func (g *Generator) parseDefInQueue() {
	if len(g.defQueue) == 0 {
		return
	}

	for name, t := range g.defQueue {
		_ = name
		z := reflect.Zero(t).Interface()
		g.ParseDefinition(z)
	}
}

// reflectTypeReliableName returns real name of given reflect.Type
func (g *Generator) reflectTypeReliableName(t reflect.Type) string {
	if def, ok := reflect.Zero(t).Interface().(SchemaDefinition); ok {
		typeDef := def.SwaggerDef()
		if typeDef.TypeName != "" {
			return typeDef.TypeName
		}
	}
	if t.Name() != "" {
		// todo consider optionally processing package
		// return path.Base(t.PkgPath()) + t.Name()
		return t.Name()
	}
	return fmt.Sprintf("anon_%08x", reflect.Indirect(reflect.ValueOf(t)).FieldByName("hash").Uint())
}

func (g *Generator) genSchemaForType(t reflect.Type) SchemaObj {
	mapped, found := g.getMappedType(t)
	if found {
		t = reflect.TypeOf(mapped)
	}

	t = refl.DeepIndirect(t)
	smObj := SchemaObj{TypeName: t.Name()}
	typeName := refl.GoType(t)

	var floatZero float64

	switch t.Kind() {
	case reflect.Bool:
		smObj = schemaFromCommonName(commonNameBoolean)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		smObj = schemaFromCommonName(commonNameInteger)
	case reflect.Uint, reflect.Uint8, reflect.Uint16:
		smObj = schemaFromCommonName(commonNameInteger)
		smObj.Minimum = &floatZero
	case reflect.Int64:
		smObj = schemaFromCommonName(commonNameLong)
	case reflect.Uint32, reflect.Uint64:
		smObj = schemaFromCommonName(commonNameLong)
		smObj.Minimum = &floatZero
	case reflect.Float32:
		smObj = schemaFromCommonName(commonNameFloat)
	case reflect.Float64:
		smObj = schemaFromCommonName(commonNameDouble)
	case reflect.String:
		smObj = schemaFromCommonName(commonNameString)
	case reflect.Array, reflect.Slice:
		if t != typeOfJSONRawMsg {
			smObj.Type = "array"
			itemSchema := g.genSchemaForType(t.Elem())
			smObj.Items = &itemSchema
		}
	case reflect.Map:
		smObj.Type = "object"
		itemSchema := g.genSchemaForType(t.Elem())
		smObj.AdditionalProperties = &itemSchema
	case reflect.Struct:
		switch {
		case t == typeOfTime:
			smObj = schemaFromCommonName(commonNameDateTime)
		case reflect.PtrTo(t).Implements(typeOfTextUnmarshaler):
			smObj.Type = "string"
		default:
			name := g.reflectTypeReliableName(t)
			name = g.makeNameForType(t, name)
			smObj.Ref = refDefinitionPrefix + name
			if !g.defExists(t) || !g.defInQueue(t) {
				g.addToDefQueue(t)
			}
		}
	case reflect.Interface:
		if typeName == "mime/multipart.File" {
			smObj = SchemaObj{}
			smObj.Type = "file"
		} else if t.NumMethod() > 0 {
			panic("Non-empty interface is not supported: " + typeName)
		}
	default:
		panic(fmt.Sprintf("type %s is not supported: %s", t.Kind(), typeName))
	}

	if sd, ok := reflect.New(t).Interface().(SchemaDefinition); ok {
		smObj = sd.SwaggerDef().Schema()
		name := g.reflectTypeReliableName(t)
		name = g.makeNameForType(t, name)
		smObj = SchemaObj{Ref: refDefinitionPrefix + name}
		if !g.defExists(t) || !g.defInQueue(t) {
			g.addToDefQueue(t)
		}
	}

	if g.reflectGoTypes && smObj.Ref == "" {
		smObj.GoType = refl.GoType(t)
	}

	return smObj
}

//
// Parse struct to swagger parameter object of operation object
// see http://swagger.io/specification/#parameterObject
//

// ParseParameters parse input struct to swagger parameter object
func (g *Generator) ParseParameters(i interface{}) (string, []ParamObj) {
	v := reflect.ValueOf(i)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("struct expected for ParseParameters")
	}

	t := v.Type()

	if mappedTo, ok := g.getMappedType(t); ok {
		return g.ParseParameters(mappedTo)
	}

	requestTypeName := refl.GoType(v.Type())
	name := t.Name()
	numField := t.NumField()
	params := make([]ParamObj, 0, numField)

	filesFound := map[string]bool{}

	for i := 0; i < numField; i++ {
		field := t.Field(i)
		// we can't access the value of un-exportable or anonymous fields
		if field.PkgPath != "" || field.Anonymous {
			continue
		}

		var (
			nameTag string
			in      string
		)

		if tagIn := field.Tag.Get("in"); tagIn != "" {
			if tagName := field.Tag.Get("name"); tagName != "" {
				nameTag = tagName
				in = tagIn
			}
		}

		if in == "" {
			for _, tag := range []string{"path", "query", "header", "formData", "cookie", "file"} {
				tagValue := field.Tag.Get(tag)
				if tagValue != "" && tagValue != "-" {
					nameTag = tagValue
					in = tag
					break
				}
			}
		}

		if in == "" {
			continue
		}

		if in == "file" {
			in = "formData"
		}

		paramName := strings.Split(nameTag, ",")[0]
		param := ParamObj{}
		if def, ok := reflect.Zero(field.Type).Interface().(SchemaDefinition); ok {
			param = def.SwaggerDef().Param()
		} else {
			var schemaObj SchemaObj
			fieldTypeName := refl.GoType(field.Type)
			if fieldTypeName == "*mime/multipart.FileHeader" {
				schemaObj.Type = "file"
			} else {
				if mappedTo, ok := g.getMappedType(field.Type); ok {
					if def, ok := mappedTo.(SchemaDefinition); ok {
						schemaObj = def.SwaggerDef().Schema()
						if schemaObj.TypeName != "" {
							name = schemaObj.TypeName
						}
					} else {
						schemaObj = g.genSchemaForType(reflect.TypeOf(mappedTo))
					}
				} else {
					schemaObj = g.genSchemaForType(field.Type)
				}
			}

			if schemaObj.Type == "" {
				panic("unsupported field " + field.Name + " of type " + fieldTypeName + " in request of type " + requestTypeName)
			}

			param.CommonFields = schemaObj.CommonFields

			if schemaObj.Type == "array" && schemaObj.Items != nil {
				if schemaObj.Items.Ref != "" {
					fieldType := refl.DeepIndirect(field.Type)
					if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
						g.ParseDefinition(reflect.Zero(field.Type.Elem()).Interface())

						if def, ok := g.getDefinition(field.Type.Elem()); ok {
							schemaObj.Items = &def
						}
					}
				}

				if schemaObj.Items.Ref != "" || schemaObj.Items.Type == "array" || schemaObj.Items.Type == "object" {
					panic("unsupported array of struct or nested array in parameter: " + fieldTypeName)
				}

				param.Items = &ParamItemObj{}
				param.Items.CommonFields = schemaObj.Items.CommonFields
				param.Items.Title = ""
				param.Items.Description = ""
				param.CollectionFormat = "multi" // default for now
			}
		}

		if g.reflectGoTypes {
			param.AddExtendedField("x-go-name", field.Name)
			param.AddExtendedField("x-go-type", refl.GoType(field.Type))
		}

		param.Name = paramName

		param.Enum.LoadFromField(field)
		readSharedTags(field.Tag, &param.CommonFields)
		readStringTag(field.Tag, "collectionFormat", &param.CollectionFormat)

		if in == "path" { // always true for path
			param.Required = true
		} else {
			if in != "body" { // always unset for body
				// not required by default for others
				readBoolTag(field.Tag, "required", &param.Required)
			}
		}

		param.In = in
		if param.Type == "file" {
			if filesFound[param.Name] {
				continue
			} else {
				filesFound[param.Name] = true
			}
		}
		params = append(params, param)
	}

	return name, params
}

func readSharedTags(tag reflect.StructTag, param *CommonFields) {
	readStringTag(tag, "type", &param.Type)
	readStringTag(tag, "title", &param.Title)
	readStringTag(tag, "description", &param.Description)
	readStringTag(tag, "format", &param.Format)
	readStringTag(tag, "pattern", &param.Pattern)

	readIntPtrTag(tag, "maxLength", &param.MaxLength)
	readIntPtrTag(tag, "minLength", &param.MinLength)
	readIntPtrTag(tag, "maxItems", &param.MaxItems)
	readIntPtrTag(tag, "minItems", &param.MinItems)
	readIntPtrTag(tag, "maxProperties", &param.MaxProperties)
	readIntPtrTag(tag, "minProperties", &param.MaxProperties)

	readFloatTag(tag, "multipleOf", &param.MultipleOf)
	readFloatPtrTag(tag, "maximum", &param.Maximum)
	readFloatPtrTag(tag, "minimum", &param.Minimum)

	readBoolTag(tag, "exclusiveMaximum", &param.ExclusiveMaximum)
	readBoolTag(tag, "exclusiveMinimum", &param.ExclusiveMinimum)
	readBoolTag(tag, "uniqueItems", &param.UniqueItems)
}

func readStringTag(tag reflect.StructTag, name string, holder *string) {
	value, ok := tag.Lookup(name)
	if ok {
		if *holder != "" && value == "-" {
			*holder = ""
			return
		}
		*holder = value
	}
}

func readBoolTag(tag reflect.StructTag, name string, holder *bool) {
	value, ok := tag.Lookup(name)
	if ok {
		v, err := strconv.ParseBool(value)
		if err != nil {
			panic("failed to parse bool value " + value + " in tag " + name + ": " + err.Error())
		}
		*holder = v
	}
}

func readIntPtrTag(tag reflect.StructTag, name string, holder **int64) {
	value, ok := tag.Lookup(name)
	if ok {
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			panic("failed to parse int value " + value + " in tag " + name + ": " + err.Error())
		}
		*holder = new(int64)
		**holder = v
	}

}

func readFloatTag(tag reflect.StructTag, name string, holder *float64) {
	value, ok := tag.Lookup(name)
	if ok {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic("failed to parse float value " + value + " in tag " + name + ": " + err.Error())
		}
		*holder = v
	}
}

func readFloatPtrTag(tag reflect.StructTag, name string, holder **float64) {
	value, ok := tag.Lookup(name)
	if ok {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			panic("failed to parse float value " + value + " in tag " + name + ": " + err.Error())
		}
		*holder = new(float64)
		**holder = v
	}
}

// ResetPaths remove all current paths
func (g *Generator) ResetPaths() {
	g.paths = make(map[string]PathItem)
}

var regexFindPathParameter = regexp.MustCompile(`\{([^}:]+)(:[^\/]+)?(?:\})`)

// SetPathItem register path item with some information and input, output
func (g *Generator) SetPathItem(info PathItemInfo) *OperationObj {
	var (
		item  PathItem
		found bool
	)

	var params interface{}
	if _, ok := info.Request.(SchemaDefinition); ok {
		params = info.Request
	} else if refl.IsStruct(info.Request) {
		params = info.Request
	}

	var body interface{}
	if info.Method != http.MethodGet && info.Method != http.MethodHead {
		if _, ok := info.Request.(SchemaDefinition); ok {
			body = info.Request
		} else if refl.HasTaggedFields(info.Request, "json") ||
			refl.IsSliceOrMap(info.Request) ||
			!refl.IsStruct(info.Request) {
			body = info.Request
		}
	}
	var response = info.Response

	pathParametersSubmatches := regexFindPathParameter.FindAllStringSubmatch(info.Path, -1)
	if len(pathParametersSubmatches) > 0 {
		for _, submatch := range pathParametersSubmatches {
			if submatch[2] != "" { // Remove gorilla.Mux-style regexp in path
				info.Path = strings.Replace(info.Path, submatch[0], "{"+submatch[1]+"}", 1)
			}
		}
	}

	item, found = g.paths[info.Path]

	if found && item.HasMethod(info.Method) {
		return item.Map()[info.Method]
	}

	if !found {
		item = PathItem{}
	}

	operationObj := &OperationObj{}
	operationObj.Summary = info.Title
	operationObj.Description = info.Description
	operationObj.Deprecated = info.Deprecated
	operationObj.Produces = info.Produces
	operationObj.Consumes = info.Consumes
	operationObj.additionalData = info.additionalData
	if info.Tag != "" {
		operationObj.Tags = []string{info.Tag}
	}

	operationObj.Security = make([]map[string][]string, 0)
	if len(info.Security) > 0 {
		for _, sec := range info.Security {
			operationObj.Security = append(operationObj.Security, map[string][]string{sec: {}})
		}
	}

	if len(info.SecurityOAuth2) > 0 {
		for sec, scopes := range info.SecurityOAuth2 {
			operationObj.Security = append(operationObj.Security, map[string][]string{sec: scopes})
		}
	}

	if params != nil {
		if g.reflectGoTypes {
			operationObj.AddExtendedField("x-request-go-type", refl.GoType(reflect.TypeOf(params)))
		}

		_, params := g.ParseParameters(params)
		operationObj.Parameters = params
	}

	if g.defaultResponses != nil {
		for statusCode, r := range g.defaultResponses {
			g.parseResponseObject(operationObj, statusCode, r)
		}
	}
	if info.responses != nil {
		for statusCode, r := range info.responses {
			g.parseResponseObject(operationObj, statusCode, r)
		}
	}
	if response != nil || operationObj.Responses == nil {
		statusCode := http.StatusOK
		if info.SuccessfulResponseCode != 0 {
			statusCode = info.SuccessfulResponseCode
		}
		g.parseResponseObject(operationObj, statusCode, response)
	}

	if body != nil {
		if g.reflectGoTypes {
			operationObj.AddExtendedField("x-request-go-type", refl.GoType(reflect.TypeOf(body)))
		}

		typeDef := g.ParseDefinition(body)

		if !typeDef.isEmpty() {
			param := ParamObj{
				Name:     "body",
				In:       "body",
				Required: true,
				Schema:   &typeDef,
			}

			if operationObj.Parameters == nil {
				operationObj.Parameters = make([]ParamObj, 0, 1)
			}

			operationObj.Parameters = append(operationObj.Parameters, param)
		} else {
			g.deleteDefinition(reflect.TypeOf(body))
		}
	}

	switch strings.ToUpper(info.Method) {
	case http.MethodGet:
		item.Get = operationObj
	case http.MethodPost:
		item.Post = operationObj
	case http.MethodPut:
		item.Put = operationObj
	case http.MethodDelete:
		item.Delete = operationObj
	case http.MethodOptions:
		item.Options = operationObj
	case http.MethodHead:
		item.Head = operationObj
	case http.MethodPatch:
		item.Patch = operationObj
	}

	g.paths[info.Path] = item

	return operationObj
}

func (g *Generator) parseResponseObject(operationObj *OperationObj, statusCode int, responseObj interface{}) {
	if operationObj.Responses == nil {
		operationObj.Responses = make(Responses)
	}

	if responseObj != nil {
		schema := g.ParseDefinition(responseObj)
		var desc string
		if withDesc, ok := responseObj.(description); ok {
			desc = withDesc.Description()
		} else {
			desc = http.StatusText(statusCode)
		}
		// since we only response json object
		// so, type of response object is always object
		operationObj.Responses[statusCode] = ResponseObj{
			Description: desc,
			Schema:      &schema,
		}
	} else {
		operationObj.Responses[statusCode] = ResponseObj{
			Description: http.StatusText(statusCode),
		}
	}
}
