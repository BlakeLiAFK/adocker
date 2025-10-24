package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// createImageTar 创建镜像 tar 文件
func createImageTar(repository, tag string, configData []byte, layersData [][]byte, manifest *ManifestV2, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建 tar 文件失败: %w", err)
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	// 1. 写入 config 文件
	configFileName := strings.Replace(manifest.Config.Digest, "sha256:", "", 1) + ".json"
	if err := writeToTar(tw, configFileName, configData); err != nil {
		return fmt.Errorf("写入 config 文件失败: %w", err)
	}

	// 2. 写入每个 layer
	layerFiles := []string{}
	for i, layerData := range layersData {
		layerFileName := fmt.Sprintf("%s/layer.tar", strings.Replace(manifest.Layers[i].Digest, "sha256:", "", 1))
		layerFiles = append(layerFiles, layerFileName)

		// 解压 gzip 后再写入 (Docker Hub 的 layer 是 gzip 压缩的)
		if err := writeLayerToTar(tw, layerFileName, layerData); err != nil {
			return fmt.Errorf("写入 layer 文件失败: %w", err)
		}
	}

	// 3. 创建并写入 manifest.json
	dockerManifest := DockerManifest{
		{
			Config:   configFileName,
			RepoTags: []string{fmt.Sprintf("%s:%s", repository, tag)},
			Layers:   layerFiles,
		},
	}

	manifestData, err := json.Marshal(dockerManifest)
	if err != nil {
		return fmt.Errorf("序列化 manifest 失败: %w", err)
	}

	if err := writeToTar(tw, "manifest.json", manifestData); err != nil {
		return fmt.Errorf("写入 manifest.json 失败: %w", err)
	}

	// 4. 创建并写入 repositories 文件
	repositories := map[string]map[string]string{
		repository: {
			tag: strings.Replace(manifest.Layers[len(manifest.Layers)-1].Digest, "sha256:", "", 1),
		},
	}

	repoData, err := json.Marshal(repositories)
	if err != nil {
		return fmt.Errorf("序列化 repositories 失败: %w", err)
	}

	if err := writeToTar(tw, "repositories", repoData); err != nil {
		return fmt.Errorf("写入 repositories 文件失败: %w", err)
	}

	return nil
}

// writeToTar 写入数据到 tar 文件
func writeToTar(tw *tar.Writer, name string, data []byte) error {
	header := &tar.Header{
		Name:    name,
		Size:    int64(len(data)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := tw.Write(data); err != nil {
		return err
	}

	return nil
}

// writeLayerToTar 写入 layer 到 tar (需要解压 gzip)
func writeLayerToTar(tw *tar.Writer, name string, compressedData []byte) error {
	// 创建 layer 目录
	dir := filepath.Dir(name)
	dirHeader := &tar.Header{
		Name:     dir + "/",
		Mode:     0755,
		ModTime:  time.Now(),
		Typeflag: tar.TypeDir,
	}
	if err := tw.WriteHeader(dirHeader); err != nil {
		return err
	}

	// 解压 gzip
	gzReader, err := gzip.NewReader(strings.NewReader(string(compressedData)))
	if err != nil {
		return fmt.Errorf("解压 gzip 失败: %w", err)
	}
	defer gzReader.Close()

	// 读取解压后的数据
	decompressedData, err := io.ReadAll(gzReader)
	if err != nil {
		return fmt.Errorf("读取解压数据失败: %w", err)
	}

	// 写入解压后的 layer.tar
	header := &tar.Header{
		Name:    name,
		Size:    int64(len(decompressedData)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := tw.Write(decompressedData); err != nil {
		return err
	}

	return nil
}

// getOutputFileName 生成输出文件名
func getOutputFileName(repository, tag string) string {
	// 将 / 替换为 _
	safeName := strings.ReplaceAll(repository, "/", "_")
	// 移除 library/ 前缀
	safeName = strings.TrimPrefix(safeName, "library_")
	return fmt.Sprintf("%s_%s.tar", safeName, tag)
}
