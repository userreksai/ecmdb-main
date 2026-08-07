package menu

import domainmenu "github.com/Duke1616/ecmdb/internal/menu"

func init() {
	for i := range DefaultMenus {
		switch DefaultMenus[i].Id {
		case 315:
			DefaultMenus[i].Sort = 1
		case 9:
			DefaultMenus[i].Sort = 3
		case 10:
			DefaultMenus[i].Sort = 4
		case 35:
			DefaultMenus[i].Sort = 5
		}
	}

	DefaultMenus = append(DefaultMenus,
		domainmenu.Menu{
			Id: 325, Pid: 8, Path: "/task/categories", Name: "task-category", Sort: 2,
			Component: "/views/task/category/index.vue", Status: domainmenu.Status(1), Type: domainmenu.Type(2),
			Meta: domainmenu.Meta{
				Title: "任务分类", Icon: "task-manager", Platforms: []string{"automation"},
			},
		},
		domainmenu.Menu{
			Id: 326, Pid: 325, Sort: 1000, Status: domainmenu.Status(1), Type: domainmenu.Type(3),
			Meta: domainmenu.Meta{Title: "任务分类管理", Platforms: []string{"automation"}},
			Endpoints: []domainmenu.Endpoint{
				{Path: "/api/manager/categories/create", Method: "POST", Resource: "TASK"},
				{Path: "/api/manager/categories/list", Method: "POST", Resource: "TASK"},
				{Path: "/api/manager/categories/update", Method: "POST", Resource: "TASK"},
				{Path: "/api/manager/categories/:id", Method: "DELETE", Resource: "TASK"},
				{Path: "/api/manager/categories/:id/tasks", Method: "POST", Resource: "TASK"},
				{Path: "/api/manager/list", Method: "POST", Resource: "TASK"},
			},
		},
	)
}
