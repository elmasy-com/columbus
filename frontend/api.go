package frontend

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type exampleObject struct {
	Summary     string `yaml:"summary"`
	Description string `yaml:"description"`
}

type mediaTypeObject struct {
	Schema  map[string]string `yaml:"schema"`
	Example exampleObject     `yaml:"example"`
}

type parameterObject struct {
	Name        string `yaml:"name"`
	In          string `yaml:"in"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
}

type responseObject struct {
	Description string                     `yaml:"description"`
	Content     map[string]mediaTypeObject `yaml:"content"`
}

type operationObject struct {
	OperationID string                    `yaml:"operationId"`
	Summary     string                    `yaml:"summary"`
	Description string                    `yaml:"description"`
	Parameters  []parameterObject         `yaml:"parameters"`
	Responses   map[string]responseObject `yaml:"responses"`
}

type pathItemObject struct {
	Get operationObject `yaml:"get"`
}

type paths struct {
	Paths map[string]pathItemObject `yaml:"paths"`
}

type apiData struct {
	Meta   metaData
	Method string
	Err    string
}

func fetchOpenAPIPaths(uri string) (paths, error) {

	resp, err := http.Get(uri)
	if err != nil {
		return paths{}, fmt.Errorf("failed to query %s: %w", uri, err)
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return paths{}, fmt.Errorf("failed to read body: %w", err)
	}

	ps := new(paths)

	err = yaml.Unmarshal(out, ps)

	return *ps, err
}

// func GetAPI(c *gin.Context) {

// 	dat := apiData{}
// 	dat.Meta = getMetaData(c.Request, "Columbus Project - API", DefaultDescription)

// 	ps, err := fetchOpenAPIPaths(dat.Meta.Proto + "://" + dat.Meta.Host + "/openapi.yaml")
// 	if err != nil {
// 		c.Error(fmt.Errorf("failed to fetch openapi.yaml: %w", err))
// 		Get500(c)
// 		return
// 	}

// 	c.JSON(http.StatusOK, ps)

// }

func GetAPI(c *gin.Context) {

	buf := new(bytes.Buffer)
	dat := apiData{Meta: getMetaData(c.Request, DefaultTitle, DefaultDescription)}

	err := templates.ExecuteTemplate(buf, "api", dat)
	if err != nil {
		c.Error(fmt.Errorf("failed to render api: %w", err))
		Get500(c)
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())

}
