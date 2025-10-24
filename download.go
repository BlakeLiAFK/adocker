package main

import (
	"fmt"
)

// downloadImage 下载镜像并保存为 tar 文件到当前目录
func downloadImage(image string) error {
	fmt.Printf("开始下载镜像: %s\n", image)

	// 1. 解析镜像名称
	ref := parseImageName(image)
	fmt.Printf("仓库: %s/%s, 标签: %s\n", ref.Registry, ref.Repository, ref.Tag)

	// 2. 获取认证 token
	fmt.Println("正在获取认证 token...")
	token, err := getAuthToken(ref.Registry, ref.Repository)
	if err != nil {
		return err
	}

	// 3. 获取 manifest
	fmt.Println("正在获取镜像 manifest...")
	manifest, err := getManifest(ref.Registry, ref.Repository, ref.Tag, token)
	if err != nil {
		return err
	}

	fmt.Printf("镜像包含 %d 个层\n", len(manifest.Layers))

	// 4. 下载 config
	fmt.Println("正在下载镜像配置...")
	configData, err := downloadBlob(ref.Registry, ref.Repository, manifest.Config.Digest, token)
	if err != nil {
		return fmt.Errorf("下载配置失败: %w", err)
	}
	fmt.Printf("配置下载完成 (大小: %d bytes)\n", len(configData))

	// 5. 下载所有 layers
	fmt.Println("正在下载镜像层...")
	layersData := make([][]byte, len(manifest.Layers))
	for i, layer := range manifest.Layers {
		fmt.Printf("  [%d/%d] 正在下载层 %s (大小: %.2f MB)...\n",
			i+1, len(manifest.Layers), layer.Digest[:19], float64(layer.Size)/(1024*1024))

		layerData, err := downloadBlob(ref.Registry, ref.Repository, layer.Digest, token)
		if err != nil {
			return fmt.Errorf("下载第 %d 层失败: %w", i+1, err)
		}

		layersData[i] = layerData
		fmt.Printf("  [%d/%d] 下载完成\n", i+1, len(manifest.Layers))
	}

	// 6. 生成 tar 文件到当前目录
	outputFile := getOutputFileName(ref.Repository, ref.Tag)
	fmt.Printf("正在生成 tar 文件: %s\n", outputFile)

	if err := createImageTar(ref.Repository, ref.Tag, configData, layersData, manifest, outputFile); err != nil {
		return fmt.Errorf("生成 tar 文件失败: %w", err)
	}

	fmt.Printf("✓ 镜像已成功保存到: %s\n", outputFile)
	return nil
}
