# sing-box iOS（SFI）官方源码未签名构建

这个仓库**不放任何魔改代码**，只放一条复刻官方构建流程的流水线，源码在 CI 里直接从上游取：

- 核心：`SagerNet/sing-box`（默认 `v1.14.4`）
- 客户端：`sing-box` 仓库的 `clients/apple` 子模块 = `SagerNet/sing-box-for-apple`

## 流程（照抄官方 `.github/workflows/build.yml`）

官方 job `build_apple_library` + `build_apple`，去掉签名和 App Store 上传：

1. `xcode-27` runner + Go 1.26.10
2. `.github/patch_go_for_ios.sh`（官方脚本，修 Go 在 GOOS=ios 下的 ARM64 特性检测）
3. `make lib_install` → gomobile / gobind `v0.1.13`（`github.com/sagernet/gomobile`）
4. `go run ./cmd/internal/build_libbox -target apple -platform ios/arm64` → `Libbox.xcframework`
5. `xcodebuild archive -scheme SFI -configuration Release -destination 'generic/platform=iOS'`
6. `Payload/` 打 zip 成未签名 ipa

## 触发

- 手动：Actions → *Build SFI (official sing-box for iOS) - unsigned IPA* → Run workflow，可填 `core_ref`
- 自动：push 到 `main`

产物：artifact `SFI-unsigned` → `SFI-<version>-unsigned.ipa`

## 说明

- bundle id `io.nekohasekai.sfi`，App Group `group.io.nekohasekai.sfi` —— 与官方 iOS 客户端一致。
- 包是**未签名**的：自签时需要用你自己的证书 + 匹配上述 App Group 的 provisioning profile 重签，NetworkExtension（packet-tunnel-provider）权限也必须包含在 profile 里。
- 版本号 = `core_ref`（去掉 `v`），构建号 = workflow run number。
