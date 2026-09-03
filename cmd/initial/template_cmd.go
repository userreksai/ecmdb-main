package initial

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/userreksai/ecmdb-main/cmd/initial/ioc"
	"github.com/userreksai/ecmdb-main/cmd/initial/template"
)

var templateCmd = &cobra.Command{
	Use:   "ticket-notify-template",
	Short: "初始化工单模版",
	Long:  "单独初始化工单相关的模版数据，不影响系统版本",
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化 Ioc 注册，确保 TemplateSvc 可用
		app, err := ioc.InitApp()
		cobra.CheckErr(err)

		// 获取系统版本信息 (可选，仅仅为了确认连接正常)
		currentVersion, err := app.VerSvc.GetVersion(context.Background())
		if err != nil {
			fmt.Printf("⚠️  获取系统版本失败 (可能是首次运行): %v\n", err)
		} else {
			fmt.Printf("📊 当前系统版本: %s\n", currentVersion)
		}

		if dryRun {
			fmt.Printf("🔍 干运行模式 - 预览模版初始化操作\n")
			return
		}

		fmt.Printf("🚀 开始初始化工单模版...\n")
		fmt.Printf("==================================================\n")

		init := template.NewInitial(app)
		err = init.InitTemplate()
		cobra.CheckErr(err)

		fmt.Printf("==================================================\n")
		fmt.Printf("🎉 工单模版初始化完成!\n")
	},
}

func init() {
	// 将 template 子命令添加到 init 主命令中
	Cmd.AddCommand(templateCmd)
}
