package frontend

import (
	"net/http"
	"strings"

	"github.com/e1berd/drift/app/internal/i18n"
	"github.com/labstack/echo/v5"
)

const resourcePrefix = "frontends/"

// ResourceName is the AIP-122 name of an interface.
func ResourceName(id string) string {
	return resourcePrefix + id
}

// IDFromResourceName returns the id in "frontends/{id}", or "" for any other name.
func IDFromResourceName(name string) string {
	id, ok := strings.CutPrefix(name, resourcePrefix)
	if !ok || strings.Contains(id, "/") {
		return ""
	}
	return id
}

type resource struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type listResponse struct {
	Frontends []resource `json:"frontends"`
}

func localized(texts map[string]string, lang string) string {
	for _, candidate := range []string{lang, "en"} {
		if text, ok := texts[candidate]; ok {
			return text
		}
	}
	for _, text := range texts {
		return text
	}
	return ""
}

// RegisterAPI mounts the interface catalog (AIP-132) on a group that requires a signed-in identity.
func RegisterAPI(g *echo.Group, fe Frontend, translations *i18n.Bundle) {
	g.GET("/frontends", func(c *echo.Context) error {
		lang := translations.Detect(c.Request())
		catalog := fe.Catalog()
		response := listResponse{Frontends: make([]resource, 0, len(catalog))}
		for _, info := range catalog {
			response.Frontends = append(response.Frontends, resource{
				Name:        ResourceName(info.ID),
				DisplayName: localized(info.Name, lang),
				Description: localized(info.Description, lang),
			})
		}
		return c.JSON(http.StatusOK, response)
	})
}
