package fastapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	openapi "github.com/go-openapi/spec"
	"log"
	"net/http"
	"reflect"
)

type Router struct {
	routesMap map[string]interface{}
}

func NewRouter() *Router {
	return &Router{
		routesMap: make(map[string]interface{}),
	}
}

func (r *Router) AddCall(path string, handler interface{}) {
	handlerType := reflect.TypeOf(handler)

	if handlerType.NumIn() != 2 {
		panic("Wrong number of arguments")
	}
	if handlerType.NumOut() != 2 {
		panic("Wrong number of return values")
	}

	ginCtxType := reflect.TypeOf(&gin.Context{})
	if !handlerType.In(0).ConvertibleTo(ginCtxType) {
		panic("First argument should be *gin.Context!")
	}
	// fmt.Println(handlerType.In(1).Kind() == reflect.Struct)
	if handlerType.In(1).Kind() != reflect.Struct {
		panic("Second argument must be a struct")
	}

	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if !handlerType.Out(1).Implements(errorInterface) {
		panic("Second return value should be an error")
	}
	if handlerType.Out(0).Kind() != reflect.Struct {
		panic("First return value be a struct")
	}

	r.routesMap[path] = handler
}

func (r *Router) GinHandler(c *gin.Context) {
	path := c.Param("path")
	log.Print(path)
	handlerFuncPtr, present := r.routesMap[path]
	if !present {
		c.JSON(http.StatusNotFound, gin.H{"error": "handler not found"})
		return
	}

	handlerType := reflect.TypeOf(handlerFuncPtr)
	inputType := handlerType.In(1)
	inputVal := reflect.New(inputType).Interface()
	err := c.BindJSON(inputVal)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	toCall := reflect.ValueOf(handlerFuncPtr)
	outputVal := toCall.Call(
		[]reflect.Value{
			reflect.ValueOf(c),
			reflect.ValueOf(inputVal).Elem(),
		},
	)

	returnedErr := outputVal[1].Interface()
	if returnedErr != nil || !outputVal[1].IsNil() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": returnedErr})
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": outputVal[0].Interface()})
}

func (r *Router) EmitOpenAPIDefinition() openapi.Swagger {
	sw := openapi.Swagger{}
	sw.Swagger = "2.0"
	sw.Info = &openapi.Info{}
	sw.Info.Title = "API generated with go-fastapi"
	sw.Info.Version = "1.0"
	sw.Paths = &openapi.Paths{
		Paths: make(map[string]openapi.PathItem),
	}
	sw.Definitions = make(map[string]openapi.Schema)

	definitionTypes := make(map[string]reflect.Type)
	for path, handlerFuncPtr := range r.routesMap {
		handlerType := reflect.TypeOf(handlerFuncPtr)
		inputType := handlerType.In(1)
		definitionTypes[inputType.Name()] = inputType
		for i := 0; i < inputType.NumField(); i++ {
			field := inputType.Field(i)
			if field.Type.Kind() == reflect.Struct {
				definitionTypes[field.Type.Name()] = field.Type
			}
		}

		outputType := handlerType.Out(0)
		definitionTypes[outputType.Name()] = outputType
		for i := 0; i < outputType.NumField(); i++ {
			field := outputType.Field(i)
			if field.Type.Kind() == reflect.Struct {
				definitionTypes[field.Type.Name()] = field.Type
			}
		}

		param := openapi.Parameter{}
		param.Name = "body"
		param.In = "body"
		param.Required = true
		param.Schema = openapi.RefSchema(
			fmt.Sprintf("#/definitions/%s", inputType.Name()),
		)

		op := &openapi.Operation{}
		op.Parameters = []openapi.Parameter{param}
		op.Responses = &openapi.Responses{}
		op.Responses.StatusCodeResponses = make(map[int]openapi.Response)
		ref := openapi.ResponseRef(
			fmt.Sprintf("#/definitions/%s", outputType.Name()),
		)
		op.Responses.StatusCodeResponses[http.StatusOK] = *ref

		pi := openapi.PathItem{}
		pi.Post = op
		sw.Paths.Paths[path] = pi
	}

	for definitionName, definitionType := range definitionTypes {
		props := make(map[string]openapi.Schema)
		for i := 0; i < definitionType.NumField(); i++ {
			field := definitionType.Field(i)
			fieldName := field.Tag.Get("json")
			if fieldName == "-" {
				continue
			}
			if fieldName == "" {
				fieldName = field.Name
			}

			schema := swaggerTypeFromGoType(field.Type)
			if schema == nil {
				continue
			}
			props[fieldName] = *schema
		}

		var definition openapi.Schema
		definition.Type = []string{"object"}
		definition.Properties = props
		sw.Definitions[definitionName] = definition
	}

	return sw
}

func swaggerTypeFromGoType(goType reflect.Type) *openapi.Schema {
	switch goType.Kind() {
	case reflect.Bool:
		return openapi.BoolProperty()
	case reflect.Int8:
		return openapi.Int8Property()
	case reflect.Int16:
		return openapi.Int16Property()
	case reflect.Int32:
		return openapi.Int32Property()
	case reflect.Int, reflect.Int64:
		return openapi.Int64Property()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return openapi.Int64Property()
	case reflect.Float32:
		return openapi.Float32Property()
	case reflect.Float64:
		return openapi.Float64Property()
	case reflect.String:
		return openapi.StringProperty()
	case reflect.Slice:
		return openapi.ArrayProperty(swaggerTypeFromGoType(goType.Elem()))
	case reflect.Array:
		return openapi.ArrayProperty(swaggerTypeFromGoType(goType.Elem()))
	case reflect.Map:
		return openapi.MapProperty(swaggerTypeFromGoType(goType.Elem()))
	case reflect.Struct:
		return openapi.RefProperty(
			fmt.Sprintf("#/definitions/%s", goType.Name()),
		)
	}
	return nil
}
