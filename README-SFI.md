# SFI 源码快照（官方 sing-box iOS 客户端）

分支 `vendor/v1.14.4` = **完整源码**，不再是空的流水线仓库。

## 这份源码从哪来

| 部分 | 上游 | 版本 |
|---|---|---|
| 核心（Go） | `https://github.com/SagerNet/sing-box` | tag `v1.14.4` → commit `c005f35300ce76427dc285481ec5695f5b179cd2` |
| iOS 客户端（Swift） | `https://github.com/SagerNet/sing-box-for-apple` | commit `2dd154e7586188d5fff8de615e8c821ed25ec089` |

上游把客户端放在核心里当子模块：`sing-box/clients/apple`。
为了让源码在仓库里**能直接点开看**，我把这个子模块拍平成了普通目录（`clients/apple/`，463 个文件），并从 `.gitmodules` 里去掉了对应条目。

除这两处（拍平 + 加流水线文件）外，代码**一个字节没改**，和上游 tag 一致。

## 目录

- `cmd/`、`protocol/`、`option/`、`adapter/` —— sing-box 核心（Go）
- `clients/apple/` —— iOS 客户端（Swift + Xcode 工程 `sing-box.xcodeproj`，scheme `SFI`）
- `.github/workflows/sfi-unsigned-local.yml` —— 从**本仓库这份源码**构建未签名 ipa
- `.github/workflows/build.yml` / `lint.yml` / `test.yml` —— 上游自带，只在 `stable`/`testing`/`unstable`/`oldstable` 分支触发，在本仓库不会误跑

## 构建

Actions → *Build SFI from this checkout - unsigned IPA* → Run workflow

流程与上游 `build_apple_library` + `build_apple` 一致：

1. `make lib_install` → gomobile / gobind `v0.1.13`
2. `.github/patch_go_for_ios.sh`（上游脚本）
3. `go run ./cmd/internal/build_libbox -target apple -platform ios/arm64`
4. `xcodebuild archive -scheme SFI -configuration Release -destination 'generic/platform=iOS'`
   带 `BASE_PACKAGE_IDENTIFIER=io.nekohasekai.sfi`、`APP_GROUP_IDENTIFIER=group.io.nekohasekai.sfi`、`CODE_SIGNING_ALLOWED=NO`
5. `Payload/` → zip → `SFI-<version>-unsigned.ipa`

## 重签注意

自签用的 provisioning profile 必须包含：

- App Group `group.io.nekohasekai.sfi`
- `com.apple.developer.networking.networkextension` = `packet-tunnel-provider`

否则 NetworkExtension 会静默启动失败（能装、连不上）。

## 其他分支

`master` 分支只有流水线（源码在 CI 里从上游现拉），用来验证「上游最新 tag 能不能编过」。
