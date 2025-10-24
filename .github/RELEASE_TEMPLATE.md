# 发布流程

## 自动发布

当推送符合 `v*.*.*` 格式的 tag 时,GitHub Actions 会自动:

1. 编译多平台二进制文件
2. 生成校验和文件
3. 创建 GitHub Release
4. 上传所有构建产物

## 手动发布步骤

```bash
# 1. 确保代码已提交
git add .
git commit -m "准备发布 v1.0.0"

# 2. 创建并推送 tag
git tag v1.0.0
git push origin v1.0.0

# 3. GitHub Actions 会自动构建并发布
```

## 支持的平台

- Windows (amd64, arm64)
- Linux (amd64, arm64)
- macOS (amd64, arm64)

## 构建产物

- `adocker-windows-amd64.exe` - Windows x64
- `adocker-windows-arm64.exe` - Windows ARM64
- `adocker-linux-amd64` - Linux x64
- `adocker-linux-arm64` - Linux ARM64
- `adocker-darwin-amd64` - macOS Intel
- `adocker-darwin-arm64` - macOS Apple Silicon
- `checksums.txt` - SHA256 校验和
