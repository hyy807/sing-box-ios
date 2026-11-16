#if os(iOS)
    import SwiftUI
    import UniformTypeIdentifiers

    /// 用 UIKit 的 UIDocumentPickerViewController 替代 SwiftUI 的 .fileImporter。
    ///
    /// 原因（已在真机复现并对照社区记录）：
    /// 1. SwiftUI .fileImporter 内部用 forOpeningContentTypes（asCopy: false），
    ///    对不支持“原地打开”的位置（iCloud Drive、第三方 App 目录、Telegram 落盘目录），
    ///    文件会显示但被禁用，表现为“能看见、点不动、回调不触发”。
    /// 2. 真机上 .fileImporter 还存在回调不触发、sheet 不关闭的已知缺陷；
    ///    UIKit 版本在同一条件下正常。
    ///
    /// asCopy: true 让系统先把文件复制进 app 的 Inbox，再交给 app，保证一定可选。
    public struct DocumentPicker: UIViewControllerRepresentable {
        public let contentTypes: [UTType]
        public let onComplete: (Result<[URL], Error>) -> Void

        public init(contentTypes: [UTType], onComplete: @escaping (Result<[URL], Error>) -> Void) {
            self.contentTypes = contentTypes
            self.onComplete = onComplete
        }

        public func makeCoordinator() -> Coordinator {
            Coordinator(self)
        }

        public func makeUIViewController(context: Context) -> UIDocumentPickerViewController {
            let picker = UIDocumentPickerViewController(forOpeningContentTypes: contentTypes, asCopy: true)
            picker.allowsMultipleSelection = false
            picker.shouldShowFileExtensions = true
            picker.delegate = context.coordinator
            return picker
        }

        public func updateUIViewController(_: UIDocumentPickerViewController, context _: Context) {}

        public final class Coordinator: NSObject, UIDocumentPickerDelegate {
            private let parent: DocumentPicker

            init(_ parent: DocumentPicker) {
                self.parent = parent
            }

            public func documentPicker(_: UIDocumentPickerViewController, didPickDocumentsAt urls: [URL]) {
                parent.onComplete(.success(urls))
            }

            public func documentPickerWasCancelled(_: UIDocumentPickerViewController) {
                parent.onComplete(.failure(CocoaError(.userCancelled)))
            }
        }
    }
#endif
