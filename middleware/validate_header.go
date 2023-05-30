package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/Blackmocca/go-lightweight-scheduler/internal/constants"
	"github.com/labstack/echo/v4"
)

func (m *restAPIMiddlewareServer) Authorization(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		switch m.authHeaderConfig.adapter {
		case constants.AUTH_ADAPTER_BASIC_AUTH:
			if err := m.authBasicAuth(c.Request().Header, m.authHeaderConfig.username, m.authHeaderConfig.password); err != nil {
				return err
			}
		case constants.AUTH_ADAPTER_APIKEY:
			if err := m.authWithAPIKey(c.Request().Header, m.authHeaderConfig.apikeyName, m.authHeaderConfig.apikeyValue); err != nil {
				return err
			}
		}
		return next(c)
	}
}

func (m *restAPIMiddlewareServer) existsHeaderKey(headers http.Header, key string) error {
	if headers.Get(key) == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("key %s must be required on header", key).Error())
	}
	return nil
}

func (m *restAPIMiddlewareServer) authBasicAuth(headers http.Header, username string, password string) error {
	key := "Authorization"
	if err := m.existsHeaderKey(headers, key); err != nil {
		return err
	}
	authorization := headers.Get(key)
	if strings.Index(authorization, "Basic ") == -1 {
		return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("This value is not Basic auth").Error())
	}
	encodedValue := strings.Split(authorization, " ")[1]
	expectValue := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))

	if encodedValue != expectValue {
		return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("Invalid Basic Auth Value").Error())
	}
	return nil
}

func (m *restAPIMiddlewareServer) authWithAPIKey(headers http.Header, apikeyname string, apikeyValue string) error {
	if err := m.existsHeaderKey(headers, apikeyname); err != nil {
		return err
	}
	if apikeyValue == "" || headers.Get(apikeyname) != apikeyValue {
		return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("Invalid Api Key Value").Error())
	}
	return nil
}
