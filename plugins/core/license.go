package core

import (
	"net/http"

	"github.com/eolinker/go-common/pm3"
)

func (p *plugin) licenseApis() []pm3.Api {
	return []pm3.Api{
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/common/license", []string{"context"}, []string{"license"}, p.licenseController.Get),
		pm3.CreateApiWidthDoc(http.MethodPost, "/api/v1/license/activate", []string{"context", "body"}, []string{"license"}, p.licenseController.Activate),
		pm3.CreateApiWidthDoc(http.MethodGet, "/api/v1/license/modules", []string{"context"}, []string{"modules"}, p.licenseController.Modules),
	}
}
