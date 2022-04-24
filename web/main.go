package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"web/fastapi"
)

type EchoInput struct {
	Phrase string `json:"phrase"`
}

type EchoOutput struct {
	OriginalInput EchoInput `json:"original_input"`
}

func EchoHandler(ctx *gin.Context, in EchoInput) (out EchoOutput, err error) {
	out.OriginalInput = in
	return
}

func main() {
	handler := func(c *gin.Context) {
		name := c.Param("name")
		value := c.DefaultQuery("value", "VALUE")
		c.JSON(http.StatusOK, gin.H{
			"name":  name,
			"value": value,
		})
	}

	myRouter := fastapi.NewRouter()
	myRouter.AddCall("/echo", EchoHandler)

	swagger := myRouter.EmitOpenAPIDefinition()
	swagger.Info.Title = "My awesome API"
	prefix, indent := "", "    "
	jsonBytes, _ := json.MarshalIndent(swagger, prefix, indent)
	fmt.Println(string(jsonBytes))

	router := gin.Default()
	router.GET("/path/:name", handler)
	router.POST("/api/*path", myRouter.GinHandler)
	router.Run("0.0.0.0:8888")
}
