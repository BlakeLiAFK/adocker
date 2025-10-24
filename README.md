# adocker - Docker 镜像拉取工具

一个简单的 Go 命令行工具,用于从 Docker Hub 拉取镜像。

## 功能特性

- ✅ **adocker pull** - 拉取镜像并直接导入到 Docker(一气呵成,等同于 `docker pull`)
- ✅ **adocker dl** - 下载镜像并保存为 tar 文件到当前目录
- ✅ 支持 Docker Hub 和自定义镜像仓库
- ✅ 支持 latest 和指定版本标签
- ✅ 自动处理多架构镜像 (默认 linux/amd64)
- ✅ 实时显示下载进度
- ✅ 生成与 `docker save` 兼容的 tar 文件

## 安装

### 方式 1: 下载预编译版本 (推荐)

从 [Releases](../../releases) 页面下载适合您系统的版本:

- **Windows**: `adocker-windows-amd64.exe` (x64) 或 `adocker-windows-arm64.exe` (ARM64)
- **Linux**: `adocker-linux-amd64` (x64) 或 `adocker-linux-arm64` (ARM64)
- **macOS**: `adocker-darwin-amd64` (Intel) 或 `adocker-darwin-arm64` (Apple Silicon)

Linux/macOS 需要添加执行权限:
```bash
chmod +x adocker-linux-amd64
```

### 方式 2: 从源码编译

```bash
go build -o adocker
```

## 使用方法

### 拉取并导入镜像 (adocker pull)

**一键完成下载、打包、导入,等同于 `docker pull`**

```bash
# 拉取官方镜像 (自动添加 library 命名空间)
adocker pull nginx
adocker pull hello-world

# 拉取指定版本
adocker pull nginx:1.20

# 拉取第三方镜像
adocker pull siglens/siglens:latest

# 拉取自定义仓库镜像
adocker pull registry.example.com/myapp:v1.0
adocker pull localhost:5000/test-image
```

**流程**: 下载镜像 → 打包 tar(临时目录) → 导入 Docker → 自动清理临时文件

### 下载镜像到本地 (adocker dl)

**下载镜像并保存为 tar 文件,用于离线传输**

```bash
# 下载官方镜像
adocker dl nginx

# 下载指定版本
adocker dl nginx:1.20

# 下载第三方镜像
adocker dl siglens/siglens:latest

# 下载自定义仓库镜像
adocker dl registry.example.com/myapp:v1.0
adocker dl localhost:5000/test-image
```

镜像会被保存为 tar 文件到当前目录,文件名格式: `<image>_<tag>.tar`
可以使用 `docker load -i <tar-file>` 加载到 Docker

## 支持的镜像格式

| 格式 | 示例 | 说明 |
|------|------|------|
| 官方镜像 | `nginx` | 自动添加 `library/` 前缀 |
| 官方镜像(带版本) | `nginx:1.20` | 指定版本号 |
| 第三方镜像 | `siglens/siglens:latest` | 包含命名空间 |
| 自定义仓库 | `registry.example.com/app:v1` | 完整仓库地址 |
| 本地仓库 | `localhost:5000/myapp` | 本地开发环境 |

## 工作原理

1. **adocker pull** (拉取并导入):
   - 使用 Docker Registry HTTP API V2 获取镜像信息
   - 下载镜像配置和所有层 (layers)
   - 在临时目录生成 tar 文件
   - 调用 `docker load` 导入镜像
   - 自动清理临时文件

2. **adocker dl** (下载保存):
   - 使用 Docker Registry HTTP API V2 获取镜像信息
   - 下载镜像配置和所有层 (layers)
   - 在当前目录生成 tar 文件

## 技术栈

- Go 标准库
- Docker Registry HTTP API V2
- 无外部依赖

## 示例

### 示例 1: 拉取并导入镜像

```bash
$ adocker pull nginx
开始拉取镜像: nginx
仓库: library/nginx, 标签: latest
正在获取认证 token...
正在获取镜像 manifest...
镜像包含 7 个层
正在下载镜像配置...
配置下载完成 (大小: 7234 bytes)
正在下载镜像层...
  [1/7] 正在下载层 sha256:... (大小: 27.15 MB)...
  [1/7] 下载完成
  ...
正在生成镜像包到临时目录: C:\Users\...\Temp\adocker-123456\image.tar
正在导入镜像到 Docker...
Loaded image: library/nginx:latest
✓ 镜像 library/nginx:latest 已成功导入到 Docker
正在清理临时文件: C:\Users\...\Temp\adocker-123456
临时文件已清理
```

### 示例 2: 下载镜像到本地

```bash
$ adocker dl nginx
开始下载镜像: nginx
仓库: library/nginx, 标签: latest
正在获取认证 token...
正在获取镜像 manifest...
镜像包含 7 个层
正在下载镜像配置...
正在下载镜像层...
  [1/7] 正在下载层 sha256:... (大小: 27.15 MB)...
  ...
正在生成 tar 文件: nginx_latest.tar
✓ 镜像已成功保存到: nginx_latest.tar

# 稍后可以加载
$ docker load -i nginx_latest.tar
```

## 文件结构

```
adocker/
├── main.go           # 入口文件,命令行参数解析
├── types.go          # 数据结构定义
├── registry.go       # Docker Registry API 交互
├── tar.go            # tar 文件生成逻辑
├── pull.go           # pull 命令实现 (拉取并导入)
├── download.go       # dl 命令实现 (下载保存)
├── plan.md           # 开发计划
└── README.md         # 本文件
```

## 离线使用场景

1. 在有网络的机器上下载镜像:
   ```bash
   adocker dl nginx
   adocker dl redis:7.0
   ```

2. 将 tar 文件复制到离线机器

3. 在离线机器上加载镜像:
   ```bash
   docker load -i nginx_latest.tar
   docker load -i redis_7.0.tar
   ```

## 命令对比

| 命令 | 功能 | 输出位置 | 清理临时文件 | 等同于 |
|------|------|----------|--------------|--------|
| `adocker pull` | 拉取并导入到 Docker | 直接导入 Docker | ✅ 是 | `docker pull` |
| `adocker dl` | 下载并保存为 tar | 当前目录 | ❌ 否 | `docker pull` + `docker save` |

## 发布说明

### 如何发布新版本

```bash
# 创建版本标签并推送
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions 会自动:
# 1. 编译 6 个平台的二进制文件
# 2. 生成 SHA256 校验和
# 3. 创建 GitHub Release
# 4. 上传所有构建产物
```

查看 [Releases](../../releases) 获取所有版本。

## 许可证

MIT
