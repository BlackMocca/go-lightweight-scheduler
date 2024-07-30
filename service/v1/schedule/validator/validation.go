package validator

import (
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/xeipuuv/gojsonschema"
)

type Validation struct {
	triggerSchema            []byte
	unActivatedTriggerSchema []byte
}

func NewValidation() Validation {
	bu, err := ioutil.ReadFile("./assets/jsonschema/v1/schedule/trigger_schema.json")
	if err != nil {
		panic(err)
	}
	bu2, err := ioutil.ReadFile("./assets/jsonschema/v1/schedule/unactivated_trigger_schema.json")
	if err != nil {
		panic(err)
	}
	return Validation{triggerSchema: bu, unActivatedTriggerSchema: bu2}
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
			field = cast.ToString(item.Details()["property"])
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

func (v Validation) ValidateUnActivatedTrigger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		schema, err := v.getLoader(v.unActivatedTriggerSchema)
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
