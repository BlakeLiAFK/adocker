# adocker - Docker镜像拉取工具 开发计划

## 项目目标
创建一个简单的 Go 命令行工具,用于从 Docker Hub 拉取镜像并保存为 tar 文件,以及从 tar 文件加载镜像到 Docker。

## 核心功能(已调整)
1. **adocker pull <image>** - 从 Docker Hub 拉取镜像并直接导入到 Docker(一气呵成,等同于 docker pull)
2. **adocker dl <image>** - 从 Docker Hub 下载镜像并保存为 tar 文件到当前目录

## 功能对比
- `adocker pull` = 下载 → 打包 tar(临时) → 导入 Docker → 删除临时文件
- `adocker dl` = 下载 → 打包 tar → 保存到当前目录

## 技术方案

### 1. Docker Registry API V2
使用 Docker Registry HTTP API V2 来拉取镜像:
- 获取认证 token (Docker Hub)
- 获取镜像 manifest
- 下载镜像 layers (blobs)
- 构建符合 OCI 标准的 tar 包

### 2. 镜像格式
生成的 tar 文件需要包含:
- manifest.json - 镜像清单
- config.json - 镜像配置
- layer.tar - 各层的 tar 文件
- repositories - 仓库信息

### 3. 加载功能
直接调用 Docker CLI 或使用 Docker Sadocker 加载 tar 文件

## 实现步骤

### 第一阶段:基础框架
- [x] 初始化项目结构
- [ ] 创建 main.go 入口文件
- [ ] 实现命令行参数解析 (使用标准库 flag)

### 第二阶段:Pull 功能实现
- [ ] 实现镜像名称解析 (支持 image:tag 格式)
- [ ] 实现 Docker Hub 认证
- [ ] 实现获取镜像 manifest
- [ ] 实现下载镜像 layers
- [ ] 实现下载镜像 config
- [ ] 实现生成 tar 文件

### 第三阶段:Load 功能实现
- [ ] 实现调用 docker load 命令加载 tar 文件

### 第四阶段:优化和测试
- [ ] 添加进度显示
- [ ] 添加错误处理
- [ ] 测试各种镜像场景

## 文件结构
```
adocker/
├── go.mod
├── go.sum
├── plan.md
├── main.go           # 入口文件,命令行参数解析
├── pull.go           # 拉取镜像逻辑
├── load.go           # 加载镜像逻辑
├── registry.go       # Docker Registry API 交互
├── tar.go            # tar 文件生成逻辑
└── types.go          # 数据结构定义
```

## 依赖包
- 标准库: net/http, encoding/json, archive/tar, compress/gzip
- 可能需要: github.com/docker/docker (Docker Sadocker)

## 实现细节

### Pull 命令流程
1. 解析镜像名称 (namespace/name:tag)
2. 默认 registry: registry-1.docker.io
3. 默认 tag: latest
4. 获取认证 token
5. 获取 manifest (v2 schema 2)
6. 下载 config blob
7. 下载所有 layer blobs
8. 生成 manifest.json 和 repositories 文件
9. 打包成 tar 文件,文件名: <image>_<tag>.tar

### Load 命令流程
1. 检查 tar 文件是否存在
2. 调用 docker load -i <tar-file>

## 注意事项
1. 使用最简单直接的方式,不过度封装
2. 支持 latest 和指定版本
3. tar 文件保存到当前目录
4. 需要处理多架构镜像的情况 (默认使用 linux/amd64)
5. 需要正确处理 Docker Hub 的认证机制

## 当前状态 (v2.0 - 重新设计)
- ✅ 项目已完成重构
- ✅ 核心功能已实现并测试通过
- ✅ 支持多架构镜像 (manifest list)
- ✅ 成功测试 pull 和 dl 命令

## 实现亮点
1. **简洁设计**: 只有两个命令,清晰明了
2. **多架构支持**: 自动处理 manifest list,默认选择 linux/amd64
3. **进度显示**: 实时显示下载进度和文件大小
4. **临时文件管理**: pull 命令自动清理临时文件
5. **错误处理**: 完善的错误提示和异常处理

## 测试结果
```bash
# 测试 1: 下载镜像到当前目录
$ adocker dl hello-world
✓ 镜像已成功保存到: hello-world_latest.tar (17KB)

# 测试 2: 拉取并导入镜像 (一气呵成)
$ adocker pull hello-world
开始拉取镜像: hello-world
...
正在导入镜像到 Docker...
✓ 镜像 library/hello-world:latest 已成功导入到 Docker
(临时文件已自动清理)
```

## 使用说明
```bash
# 编译
go build -o adocker.exe

# 拉取并导入镜像 (等同于 docker pull)
adocker pull nginx
adocker pull nginx:1.20
adocker pull siglens/siglens:latest

# 下载镜像到当前目录 (用于离线传输)
adocker dl nginx
adocker dl nginx:1.20

# 离线环境加载镜像
docker load -i nginx_latest.tar
```

## v2.0 改进内容
1. **命令重构**:
   - `adocker pull` → 拉取并导入(一气呵成)
   - `adocker dl` → 下载保存 tar
   - 移除 `adocker load` 命令

2. **临时文件管理**:
   - pull 使用临时目录,自动清理
   - dl 保存到当前目录

3. **用户体验优化**:
   - 更符合 docker 命令习惯
   - 减少手动操作步骤

4. **自定义仓库支持** (v2.1):
   - 支持 Docker Hub 和自定义镜像仓库
   - 智能识别仓库地址
   - 支持格式: `registry.example.com/app:v1`, `localhost:5000/image`

5. **日志增强**:
   - 显示临时文件路径
   - 显示清理过程

6. **CI/CD**:
   - GitHub Actions 自动发布
   - 支持 Windows/Linux/macOS 多平台
   - 自动生成 checksums
