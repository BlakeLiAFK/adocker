package main

// ManifestList Docker manifest list (多架构)
type ManifestList struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Manifests     []struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
		Platform  struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
			Variant      string `json:"variant,omitempty"`
		} `json:"platform"`
	} `json:"manifests"`
}

// ManifestV2 Docker Registry API V2 manifest 结构
type ManifestV2 struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers []Layer `json:"layers"`
}

// Layer 镜像层信息
type Layer struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

// AuthToken Docker Hub 认证 token
type AuthToken struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// DockerManifest Docker tar 包中的 manifest.json
type DockerManifest []struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// ImageConfig 镜像配置 (简化版)
type ImageConfig struct {
	Architecture string `json:"architecture"`
	Config       struct {
		Hostname     string            `json:"Hostname"`
		Domainname   string            `json:"Domainname"`
		User         string            `json:"User"`
		AttachStdin  bool              `json:"AttachStdin"`
		AttachStdout bool              `json:"AttachStdout"`
		AttachStderr bool              `json:"AttachStderr"`
		Tty          bool              `json:"Tty"`
		OpenStdin    bool              `json:"OpenStdin"`
		StdinOnce    bool              `json:"StdinOnce"`
		Env          []string          `json:"Env"`
		Cmd          []string          `json:"Cmd"`
		Image        string            `json:"Image"`
		Volumes      map[string]struct{} `json:"Volumes"`
		WorkingDir   string            `json:"WorkingDir"`
		Entrypoint   []string          `json:"Entrypoint"`
		OnBuild      []string          `json:"OnBuild"`
		Labels       map[string]string `json:"Labels"`
	} `json:"config"`
	Container       string `json:"container"`
	ContainerConfig struct {
		Hostname     string            `json:"Hostname"`
		Domainname   string            `json:"Domainname"`
		User         string            `json:"User"`
		AttachStdin  bool              `json:"AttachStdin"`
		AttachStdout bool              `json:"AttachStdout"`
		AttachStderr bool              `json:"AttachStderr"`
		Tty          bool              `json:"Tty"`
		OpenStdin    bool              `json:"OpenStdin"`
		StdinOnce    bool              `json:"StdinOnce"`
		Env          []string          `json:"Env"`
		Cmd          []string          `json:"Cmd"`
		Image        string            `json:"Image"`
		Volumes      map[string]struct{} `json:"Volumes"`
		WorkingDir   string            `json:"WorkingDir"`
		Entrypoint   []string          `json:"Entrypoint"`
		OnBuild      []string          `json:"OnBuild"`
		Labels       map[string]string `json:"Labels"`
	} `json:"container_config"`
	Created       string `json:"created"`
	DockerVersion string `json:"docker_version"`
	History       []struct {
		Created    string `json:"created"`
		CreatedBy  string `json:"created_by"`
		EmptyLayer bool   `json:"empty_layer,omitempty"`
	} `json:"history"`
	OS     string `json:"os"`
	Rootfs struct {
		Type    string   `json:"type"`
		DiffIDs []string `json:"diff_ids"`
	} `json:"rootfs"`
}
