import Foundation

public enum AppConfiguration {
    public static let packageName: String = {
        guard let value = Bundle.main.object(forInfoDictionaryKey: "BasePackageIdentifier") as? String else {
            fatalError("Missing BasePackageIdentifier in Info.plist")
        }
        return value
    }()

    public static let appGroupID: String = {
        // 自签/侧载时 App Group 的名字由 provisioning profile 决定，
        // 硬编码官方名字（group.io.nekohasekai.sfi）必然拿不到容器 -> 启动即崩。
        // 优先读 profile 里真正派发下来的 group，其次才用 Info.plist。
        if let resolved = appGroupFromProvisioningProfile() {
            return resolved
        }
        if let value = Bundle.main.object(forInfoDictionaryKey: "AppGroupIdentifier") as? String,
           !value.isEmpty, !value.contains("$(") {
            return value
        }
        return ""
    }()

    private static func appGroupFromProvisioningProfile() -> String? {
        #if os(iOS) || os(tvOS)
            var bundles: [Bundle] = [Bundle.main]
            // 扩展进程：再往上找宿主 .app 里那份 profile
            let hostURL = Bundle.main.bundleURL
                .deletingLastPathComponent()
                .deletingLastPathComponent()
                .deletingLastPathComponent()
            if hostURL.pathExtension == "app", let host = Bundle(url: hostURL) {
                bundles.append(host)
            }
            for bundle in bundles {
                guard let url = bundle.url(forResource: "embedded", withExtension: "mobileprovision"),
                      let data = try? Data(contentsOf: url) else { continue }
                guard let start = data.range(of: Data("<?xml".utf8))?.lowerBound,
                      let end = data.range(of: Data("</plist>".utf8), options: [], in: start..<data.endIndex) else { continue }
                let plistData = data.subdata(in: start..<end.upperBound)
                guard let plist = try? PropertyListSerialization.propertyList(from: plistData, options: [], format: nil) as? [String: Any],
                      let entitlements = plist["Entitlements"] as? [String: Any],
                      let groups = entitlements["com.apple.security.application-groups"] as? [String],
                      let first = groups.first, !first.isEmpty else { continue }
                return first
            }
        #endif
        return nil
    }

    public static var teamID: String {
        guard let dotIndex = appGroupID.firstIndex(of: ".") else {
            // 没有 App Group 时不要连带把 app 打死（这里只用于 macOS XPC 校验）
            return ""
        }
        return String(appGroupID[..<dotIndex])
    }

    public static var extensionBundleID: String {
        "\(packageName).extension"
    }

    public static var systemExtensionBundleID: String {
        "\(packageName).system"
    }

    #if os(macOS)
        /// After a system-extension replace, launchd keeps the previous version's per-version provider job
        /// registered and still owning the NEMachServiceName endpoint; it refuses to let the new version
        /// reclaim a fixed endpoint name, so the app reaches the dead old listener. The version suffix gives
        /// each build its own endpoint. Must equal NEMachServiceName in SystemExtension/Info.plist, which is
        /// baked from $(APP_GROUP_IDENTIFIER).system.$(MARKETING_VERSION); Bundle.version is that same value.
        public static var systemExtensionMachServiceName: String {
            "\(appGroupID).system.\(Bundle.main.version)"
        }
    #endif

    public static var packetTunnelBundleIDs: [String] {
        if extensionBundleID == systemExtensionBundleID {
            return [extensionBundleID]
        }
        return [extensionBundleID, systemExtensionBundleID]
    }

    public static var fileProviderDomainID: String {
        "\(packageName).workingdir"
    }

    public static var widgetControlKind: String {
        "\(packageName).widget.ServiceToggle"
    }

    public static var profileUTType: String {
        "\(packageName).profile"
    }

    public static var backgroundTaskID: String {
        "\(packageName).update_profiles"
    }

    public static var iCloudContainerID: String {
        "iCloud.\(packageName)"
    }

    #if os(macOS) || JAILBREAK
        public static var rootHelperBundleID: String {
            "\(packageName).helper"
        }

        public static var rootHelperMachService: String {
            "\(appGroupID).helper"
        }
    #endif
}
