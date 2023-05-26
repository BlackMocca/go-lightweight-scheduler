package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/joncalhoun/qson"
	"github.com/labstack/echo/v4"
)

type RestAPIMiddleware interface {
	InitContext(next echo.HandlerFunc) echo.HandlerFunc
	InputForm(next echo.HandlerFunc) echo.HandlerFunc
}

type restAPIMiddlewareServer struct {
}

func NewRestAPIMiddleware() RestAPIMiddleware {
	return &restAPIMiddlewareServer{}
}

func (m *restAPIMiddlewareServer) InputForm(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := form(c); err != nil {
			var code int
			var message interface{}
			if he, ok := err.(*echo.HTTPError); ok {
				code = he.Code
				message = he.Message
			}
			return echo.NewHTTPError(code, message)
		}
		return next(c)
	}
}

func (m *restAPIMiddlewareServer) InitContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		if ctx == nil {
			bgCtx := context.Background()
			newReq := c.Request().WithContext(bgCtx)

			c.SetRequest(newReq)
		}
		return next(c)
	}
}

func form(c echo.Context) error {
	var data = map[string]interface{}{}
	reqMethod := c.Request().Method
	Header := c.Request().Header

	if reqMethod == http.MethodPost || reqMethod == http.MethodPut || reqMethod == http.MethodDelete {
		contentType := Header.Get("Content-Type")
		if strings.Contains(contentType, echo.MIMEMultipartForm) {
			form, err := c.MultipartForm()
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{"message": http.ErrMissingBoundary.Error() + " or has not any parameter"})
			}
			bu, _ := qson.ToJSON(url.Values(form.Value).Encode())
			json.Unmarshal(bu, &data)

			data, err = parseOnKeyData(data)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{"message": err.Error()})
			}

			/* รูปสำหรับ ใช้งานทั่วไป */
			if val, ok := form.File["files"]; ok {
				c.Set("files", val)
			}
			/* รูปสำหรับ ใช้งานทั่วไป */
			if val, ok := form.File["file"]; ok {
				c.Set("file", val)
			}
		} else if strings.Contains(strings.ToLower(contentType), echo.MIMEApplicationJSON) {
			var err error
			if err := json.NewDecoder(c.Request().Body).Decode(&data); err != nil && err != io.EOF {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{"message": err.Error()})
			}
			data, err = parseOnKeyData(data)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{"message": err.Error()})
			}

		} else if strings.Contains(strings.ToLower(contentType), echo.MIMEApplicationForm) {
			postForm, err := c.FormParams()
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{"message": err.Error()})
			}
			if reqMethod == http.MethodDelete {
				buf := bytes.Buffer{}
				io.Copy(&buf, c.Request().Body)
				postForm, _ = url.ParseQuery(buf.String())
			}
			if len(postForm) > 0 {
				bu, _ := qson.ToJSON(postForm.Encode())
				json.Unmarshal(bu, &data)
			}
			data, err = parseOnKeyData(data)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{"message": err.Error()})
			}
		}
	}

	if len(data) > 0 {
		c.Set("params", data)
	}
	return nil
}

func parseOnKeyData(data map[string]interface{}) (map[string]interface{}, error) {
	if data != nil && len(data) == 1 {
		/*
			support on data from json format
		*/
		if v, ok := data["data"]; ok {
			valueType := reflect.ValueOf(v).Kind()
			if valueType == reflect.Map {
				data = v.(map[string]interface{})
			} else if valueType == reflect.String {
				data = map[string]interface{}{}
				if err := json.Unmarshal([]byte(v.(string)), &data); err != nil {
					return data, err
				}
			}
		}
	}

	return data, nil
}
