package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	// DockerHubRegistry Docker Hub 镜像仓库地址
	DockerHubRegistry = "registry-1.docker.io"
	// DockerHubAuth Docker Hub 认证服务地址
	DockerHubAuth = "auth.docker.io"
)

// getAuthToken 获取认证 token
// 支持 Docker Hub 和自定义 registry
func getAuthToken(registry, repository string) (string, error) {
	// Docker Hub 使用专门的认证服务
	if registry == DockerHubRegistry {
		url := fmt.Sprintf("https://%s/token?service=registry.docker.io&scope=repository:%s:pull", DockerHubAuth, repository)

		resp, err := http.Get(url)
		if err != nil {
			return "", fmt.Errorf("获取认证 token 失败: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("获取认证 token 失败,状态码: %d", resp.StatusCode)
		}

		var token AuthToken
		if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
			return "", fmt.Errorf("解析认证 token 失败: %w", err)
		}

		if token.Token != "" {
			return token.Token, nil
		}
		return token.AccessToken, nil
	}

	// 自定义 registry 可能不需要认证,或使用基本认证
	// 这里返回空 token,如果需要认证会在后续请求中失败
	return "", nil
}

// getManifest 获取镜像 manifest (支持 manifest list)
func getManifest(registry, repository, tag, token string) (*ManifestV2, error) {
	// 首先尝试获取 manifest list
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, repository, tag)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 只有在有 token 时才设置认证头
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	// 同时接受 manifest list 和普通 manifest
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.docker.distribution.manifest.v2+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取 manifest 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取 manifest 失败,状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 首先尝试解析为 manifest list
	var manifestList ManifestList
	if err := json.Unmarshal(body, &manifestList); err == nil && len(manifestList.Manifests) > 0 {
		// 这是一个 manifest list,选择 linux/amd64 架构的 manifest
		var targetDigest string
		for _, m := range manifestList.Manifests {
			if m.Platform.OS == "linux" && m.Platform.Architecture == "amd64" {
				targetDigest = m.Digest
				break
			}
		}

		if targetDigest == "" {
			// 如果没有找到 amd64,使用第一个
			if len(manifestList.Manifests) > 0 {
				targetDigest = manifestList.Manifests[0].Digest
			} else {
				return nil, fmt.Errorf("manifest list 中没有可用的 manifest")
			}
		}

		// 获取具体架构的 manifest
		return getManifestByDigest(registry, repository, targetDigest, token)
	}

	// 直接解析为 ManifestV2
	var manifest ManifestV2
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("解析 manifest 失败: %w, 响应: %s", err, string(body))
	}

	return &manifest, nil
}

// getManifestByDigest 根据 digest 获取特定的 manifest
func getManifestByDigest(registry, repository, digest, token string) (*ManifestV2, error) {
	url := fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, repository, digest)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取 manifest 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取 manifest 失败,状态码: %d", resp.StatusCode)
	}

	var manifest ManifestV2
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("解析 manifest 失败: %w", err)
	}

	return &manifest, nil
}

// downloadBlob 下载 blob (layer 或 config)
func downloadBlob(registry, repository, digest, token string) ([]byte, error) {
	url := fmt.Sprintf("https://%s/v2/%s/blobs/%s", registry, repository, digest)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载 blob 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载 blob 失败,状态码: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 blob 数据失败: %w", err)
	}

	return data, nil
}

// ImageRef 镜像引用信息
type ImageRef struct {
	Registry   string // 仓库地址,如 registry-1.docker.io
	Repository string // 镜像仓库,如 library/nginx
	Tag        string // 标签,如 latest
}

// parseImageName 解析镜像名称
// 支持格式:
//   - nginx                              -> registry-1.docker.io/library/nginx:latest
//   - nginx:1.20                         -> registry-1.docker.io/library/nginx:1.20
//   - siglens/siglens:latest            -> registry-1.docker.io/siglens/siglens:latest
//   - registry.example.com/app:v1       -> registry.example.com/app:v1
//   - localhost:5000/myapp              -> localhost:5000/myapp:latest
func parseImageName(image string) ImageRef {
	ref := ImageRef{
		Registry: DockerHubRegistry,
		Tag:      "latest",
	}

	// 分离 tag
	parts := strings.SplitN(image, ":", 2)
	imagePath := parts[0]
	if len(parts) > 1 && !strings.Contains(parts[1], "/") {
		// 确保不是 localhost:5000/image 这种格式
		ref.Tag = parts[1]
	} else if len(parts) > 1 {
		// localhost:5000/image 格式,冒号后是端口号
		imagePath = image
	}

	// 分离 registry
	pathParts := strings.SplitN(imagePath, "/", 2)

	// 判断第一部分是否是 registry 地址
	// 包含 . 或 : 或是 localhost 则认为是自定义 registry
	if len(pathParts) > 1 && (strings.Contains(pathParts[0], ".") ||
		strings.Contains(pathParts[0], ":") ||
		pathParts[0] == "localhost") {
		ref.Registry = pathParts[0]
		ref.Repository = pathParts[1]

		// 处理 repository 中的 tag
		repoParts := strings.SplitN(ref.Repository, ":", 2)
		if len(repoParts) > 1 {
			ref.Repository = repoParts[0]
			ref.Tag = repoParts[1]
		}
	} else {
		// Docker Hub 格式
		ref.Repository = imagePath

		// 如果没有命名空间,添加 library (Docker Hub 官方镜像)
		if !strings.Contains(ref.Repository, "/") {
			ref.Repository = "library/" + ref.Repository
		}
	}

	return ref
}
