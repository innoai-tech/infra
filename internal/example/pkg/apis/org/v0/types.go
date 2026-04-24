package v0

import "time"

// Info 表示组织的基础信息。
type Info struct {
	// 组织名称
	Name string `json:"name" validate:"@string[0,5]"`
	// 组织类型
	Type Type `json:"type,omitempty"`
}

// Detail 表示组织详情。
type Detail struct {
	Info
	CreatedAt *time.Time `json:"createdAt,omitempty"`
}

// DataList 表示组织列表响应体。
type DataList struct {
	Data  []Info `json:"data"`
	Total int    `json:"total"`
}

// Type 表示组织类型。
type Type int

const (
	// TYPE_UNKNOWN 表示未知类型。
	TYPE_UNKNOWN Type = iota

	// TYPE__GOV 表示政府组织。
	TYPE__GOV
	// TYPE__COMPANY 表示企事业单位。
	TYPE__COMPANY
)
