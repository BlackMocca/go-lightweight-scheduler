package validator

import (
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xeipuuv/gojsonschema"
)

type Validation struct {
	triggerSchema []byte
}

func NewValidation() Validation {
	bu, err := ioutil.ReadFile("./service/v1/schedule/validator/trigger_schema.json")
	if err != nil {
		panic(err)
	}
	return Validation{triggerSchema: bu}
}

func (v Validation) getLoader(bu []byte) (*gojsonschema.Schema, error) {
	loader := gojsonschema.NewSchemaLoader()
	loader.Draft = gojsonschema.Draft7
	loader.AutoDetect = false
	schema, err := loader.Compile(gojsonschema.NewBytesLoader(bu))
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func (v Validation) toMap(results []gojsonschema.ResultError) map[string]interface{} {
	m := map[string]interface{}{"errors": []interface{}{}}
	for _, item := range results {
		field := item.Field()
		if field == "(root)" {
			field = ""
		}

		m["errors"] = append(m["errors"].([]interface{}), []interface{}{
			map[string]interface{}{
				"field":       field,
				"description": item.Description(),
				"details":     item.Details(),
				"error_from":  item.Type(),
			},
		})
	}
	return m
}

func (v Validation) ValidateTrigger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		schema, err := v.getLoader(v.triggerSchema)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		var params = c.Get("params").(map[string]interface{})

		result, err := schema.Validate(gojsonschema.NewGoLoader(params))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if !result.Valid() {
			return echo.NewHTTPError(http.StatusBadRequest, v.toMap(result.Errors()))
		}

		return next(c)
	}
}
