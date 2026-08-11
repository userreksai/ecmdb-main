package menu

import "github.com/Duke1616/ecmdb/internal/menu"

func init() {
	for index := range DefaultMenus {
		if DefaultMenus[index].Id == 292 {
			DefaultMenus[index].Endpoints = append(DefaultMenus[index].Endpoints, menu.Endpoint{
				Path: "/api/dataio/import/preview", Method: "POST", Resource: "CMDB",
			})
			break
		}
	}

	DefaultMenus = append(DefaultMenus, menu.Menu{
		Id:        327,
		Pid:       15,
		Path:      "operation-log",
		Name:      "system-operation-log",
		Sort:      8,
		Component: "/views/system/operation-log/index.vue",
		Status:    menu.Status(1),
		Type:      menu.Type(2),
		Meta: menu.Meta{
			Title: "操作日志", Icon: "oneterm-operation_log", Platforms: []string{"system"},
		},
		Endpoints: []menu.Endpoint{{
			Path: "/api/operation-log/list", Method: "POST", Resource: "CMDB",
		}},
	})
}
