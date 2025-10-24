package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "pull":
		if len(os.Args) < 3 {
			fmt.Println("错误: 缺少镜像名称")
			fmt.Println("用法: dk pull <image>[:tag]")
			os.Exit(1)
		}
		image := os.Args[2]
		if err := pullAndLoad(image); err != nil {
			fmt.Printf("拉取镜像失败: %v\n", err)
			os.Exit(1)
		}

	case "dl":
		if len(os.Args) < 3 {
			fmt.Println("错误: 缺少镜像名称")
			fmt.Println("用法: dk dl <image>[:tag]")
			os.Exit(1)
		}
		image := os.Args[2]
		if err := downloadImage(image); err != nil {
			fmt.Printf("下载镜像失败: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("adocker - Docker 镜像拉取工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  adocker pull <image>[:tag]     拉取镜像并导入到 Docker")
	fmt.Println("  adocker dl <image>[:tag]       下载镜像并保存为 tar 文件")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  adocker pull nginx                         # 拉取并导入镜像")
	fmt.Println("  adocker pull nginx:1.20                    # 指定版本")
	fmt.Println("  adocker pull siglens/siglens:latest        # 第三方镜像")
	fmt.Println("  adocker pull registry.example.com/app:v1   # 自定义仓库")
	fmt.Println("  adocker dl nginx                           # 下载到当前目录")
}
