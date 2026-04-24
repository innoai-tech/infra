package v0

import (
	apiv0 "github.com/innoai-tech/infra/internal/example/pkg/apis/org/v0"
	"github.com/octohelm/courier/pkg/courierhttp"
)

// CreateOrg 定义创建组织接口。
type CreateOrg struct {
	courierhttp.MethodPost `path:"/orgs"`
	Body                   apiv0.Info `in:"body"`
}

// ListOrg 定义拉取组织列表接口。
type ListOrg struct {
	courierhttp.MethodGet `path:"/orgs"`
}

// GetOrg 定义查询组织详情接口。
type GetOrg struct {
	courierhttp.MethodGet `path:"/orgs/:orgName"`
	OrgName               string `name:"orgName" in:"path"`
}

// DeleteOrg 定义删除组织接口。
type DeleteOrg struct {
	courierhttp.MethodDelete `path:"/orgs/:orgName"`
	OrgName                  string `name:"orgName" in:"path"`
}
