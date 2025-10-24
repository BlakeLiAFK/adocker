package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// pullAndLoad 拉取镜像并直接导入到 Docker (一气呵成)
func pullAndLoad(image string) error {
	fmt.Printf("开始拉取镜像: %s\n", image)

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

	// 6. 创建临时目录
	tempDir, err := os.MkdirTemp("", "adocker-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer func() {
		fmt.Printf("正在清理临时文件: %s\n", tempDir)
		os.RemoveAll(tempDir)
		fmt.Println("临时文件已清理")
	}()

	// 7. 生成临时 tar 文件
	tempTarFile := filepath.Join(tempDir, "image.tar")
	fmt.Printf("正在生成镜像包到临时目录: %s\n", tempTarFile)

	if err := createImageTar(ref.Repository, ref.Tag, configData, layersData, manifest, tempTarFile); err != nil {
		return fmt.Errorf("生成镜像包失败: %w", err)
	}

	// 8. 导入到 Docker
	fmt.Println("正在导入镜像到 Docker...")
	cmd := exec.Command("docker", "load", "-i", tempTarFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("导入镜像失败: %w", err)
	}

	fmt.Printf("✓ 镜像 %s/%s:%s 已成功导入到 Docker\n", ref.Registry, ref.Repository, ref.Tag)
	return nil
}
