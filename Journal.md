# CLAUDE CODE 01
Tốt. Em đã đọc xong codebase. Đây là **Lightweight Mode** — bài toán đã rõ, chỉ cần xác định đúng điểm móc và cách triển khai đúng kiến trúc.

---

## Phân tích vấn đề

**Tại sao gõ ra chữ "a"?** JOKey intercept các keypress trước khi Neru nhận. Neru dùng `CGEventTap` (Objective-C) để bắt key, nhưng JOKey đã compose chúng thành ký tự tiếng Việt trước đó. Giải pháp: chuyển bộ gõ sang ABC khi Neru *kích hoạt*, và khôi phục khi *thoát*.

---

## Hai điểm móc quan trọng

```
mode_setup.go:77  → setModeLocked()   ← ĐÂY: mọi mode activation đều đi qua đây
mode_setup.go:62  → SetModeIdle()     ← ĐÂY: mọi lần thoát mode đều kết thúc ở đây
```

Hai hàm này là **single choke point** — không cần hook vào từng mode riêng lẻ.

---

## Kế hoạch (5–6 files)

Theo đúng kiến trúc Hexagonal của Neru, thứ tự sẽ là:

### 1. Port (interface) — `internal/core/ports/input_method.go` *(mới)*
```go
type InputMethodPort interface {
    GetCurrentSourceID() string
    SwitchToSourceID(id string) error
}
```
*Tại sao?* Hexagonal Architecture yêu cầu mọi dependency vào OS phải qua interface. Code Go thuần không được "biết" về Carbon API.

### 2. Adapter Objective-C — `internal/core/infra/platform/darwin/input_method_darwin.m` *(mới)*
Dùng Carbon framework: `TISCopyCurrentKeyboardInputSource()` + `TISSelectInputSource()`.

### 3. CGO bridge — `internal/core/infra/platform/darwin/input_method_darwin.go` *(mới, `//go:build darwin`)*
Wrap C functions sang Go, implement `InputMethodPort`.

### 4. Handler — `internal/app/modes/handler.go`
Thêm 2 field vào `Handler` struct:
```go
inputMethod          ports.InputMethodPort
previousInputSource  string
```

### 5. Mode setup — `internal/app/modes/mode_setup.go`
- `setModeLocked()` (line 77): lưu source hiện tại → `previousInputSource`, rồi switch sang ABC
- `SetModeIdle()` (line 62): restore `previousInputSource`

### 6. Wiring — `internal/app/app_initialization.go`
Inject `InputMethodAdapter` vào `Handler` khi khởi tạo.

---

## Dependency flow sau khi thêm

```
mode_setup.go
    ↓ gọi
Handler.inputMethod  (ports.InputMethodPort)
    ↓ implement bởi
darwin/input_method_darwin.go  (CGO bridge)
    ↓ gọi C
darwin/input_method_darwin.m  (Carbon TIS API)
```

---

## Risk duy nhất cần lưu ý

`TISSelectInputSource` là synchronous nhưng macOS propagate switch hơi delay (~1–2ms). Nếu gọi switch *sau* khi bật event tap, keypress đầu tiên vẫn có thể qua JOKey.

**Giải pháp**: Gọi switch input method *trước* `setAppModeLocked()` trong `setModeLocked()`, tức là trước khi event tap được enable.

---

**Câu hỏi để anh suy nghĩ trước khi approve:**

Anh thấy tại sao em chọn `setModeLocked()` thay vì hook vào từng `Activate()` của từng mode không? Và tại sao cần một Port interface thay vì gọi thẳng Carbon API từ `mode_setup.go`?

# GEMINI 01
🥷 Để làm được điều này chuẩn theo kiến trúc Hexagonal của dự án, chúng ta sẽ cần làm 3 bước chính. Anh xem qua luồng này xem có hợp lý không nhé:

1. **Định nghĩa Port (Domain/Ports):** Ta sẽ tạo một interface (ví dụ: `InputSourcePort`) trong thư mục `internal/core/ports/`. Interface này sẽ có các hàm như `GetInputSource() string` và `SetInputSource(id string) error`. Việc này giúp logic cốt lõi không bị trói buộc vào macOS.
2. **Viết Adapter (Infra/Platform):** Tại `internal/core/infra/platform/darwin/`, ta sẽ viết file Go dùng Cgo để gọi xuống API của macOS (như `TISCopyCurrentKeyboardInputSource` và `TISSelectInputSource` trong framework Carbon) để thực thi hai hàm trên. (Với Linux/Windows ta chỉ cần viết file stub rỗng để không lỗi build).
3. **Gắn vào App Lifecycle (App/Modes):** Khi một chế độ (như Grid, Hints) được gọi `Activate()` trong `internal/app/modes/`, ta sẽ lưu lại bộ gõ hiện tại và gọi hàm Set sang `com.apple.keylayout.ABC`. Khi gọi `Deactivate()`, ta sẽ khôi phục lại.

**Một rủi ro cần lưu ý:**
API đổi bộ gõ của macOS đôi khi phản hồi chậm (vài mili-giây). Nếu user gõ phím ngay lập tức ở khoảnh khắc kích hoạt Neru, ký tự đầu tiên có thể vẫn bị dính bộ gõ cũ. Để kiểm chứng, ta sẽ cần log thời gian chuyển đổi khi làm thực tế.

Anh thấy cấu trúc 3 bước này đã rõ ràng và đúng với triết lý của dự án chưa ạ? Nếu anh OK, em sẽ hướng dẫn anh viết từ bước 1 (định nghĩa Port) nhé!

# CLAUDE CODE 02
🥷 Đây là dạng **Quick Reference** — anh cần hiểu kế hoạch, không phải nghiên cứu domain mới. Em sẽ đi qua từng bước theo đúng thứ tự tư duy, không phải thứ tự file.

---

## Câu hỏi gốc: "Bộ gõ cần chuyển ở đâu trong code?"

Trước khi nói đến file nào, hãy hỏi: **ai biết khi nào Neru được kích hoạt?**

Câu trả lời là `setModeLocked()` trong `mode_setup.go:77`. Mọi mode (hints, grid, scroll, recursive-grid) đều đi qua đây khi bật lên. Tương tự, `SetModeIdle()` ở line 62 là nơi mọi mode kết thúc. Đây là hai điểm duy nhất cần động vào — không cần sửa từng mode riêng.

---

## Tại sao cần 6 files thay vì 1?

Câu hỏi tốt. Nếu viết tắt, anh có thể làm thế này trong `mode_setup.go`:

```go
// CÁCH SAI theo kiến trúc Neru
import "C"
// #include <Carbon/Carbon.h>
func setModeLocked(...) {
    C.TISSelectInputSource(...)  // gọi thẳng
}
```

Nhưng Neru cấm điều này — `depguard` sẽ báo lỗi. Lý do là: **code Go thuần không được phụ thuộc trực tiếp vào OS API**. Đây là nguyên tắc Hexagonal Architecture: core logic (modes) không được biết về Carbon, UIKit, hay bất kỳ framework OS nào.

Vậy cần xây một "đường ống" qua 3 lớp:

```
mode_setup.go  →  InputMethodPort  →  darwin adapter  →  Carbon API (Objective-C)
   (Go logic)      (interface)        (CGO bridge)         (.m file)
```

---

## Lớp 1: Port interface (file mới)

```go
// internal/core/ports/input_method.go
type InputMethodPort interface {
    GetCurrentSourceID() string
    SwitchToSourceID(id string) error
}
```

Đây chỉ là **hợp đồng** — "tôi cần ai đó làm được 2 việc này". `mode_setup.go` chỉ biết đến interface này, không biết bên dưới là Carbon hay cái gì khác. Khi test, anh có thể dùng mock thay vì gọi thật vào macOS.

---

## Lớp 2 & 3: Objective-C + CGO bridge (2 files mới)

File `.m` gọi Carbon framework — đây là macOS API cho phép đọc/đổi bộ gõ:

```
TISCopyCurrentKeyboardInputSource()  ← đọc bộ gõ hiện tại
TISSelectInputSource()               ← đổi sang bộ gõ khác
```

File `.go` (với `//go:build darwin`) wrap các C function đó và implement `InputMethodPort`. Build tag `darwin` đảm bảo file này chỉ compile trên macOS — Linux/Windows sẽ không bị ảnh hưởng.

---

## Lớp 4: Handler giữ state (file sửa)

Handler struct cần thêm:

```go
inputMethod         ports.InputMethodPort  // để gọi switch
previousInputSource string                 // để nhớ "trước khi bật là gì"
```

`previousInputSource` quan trọng vì: khi thoát Neru, anh muốn quay lại **đúng bộ gõ cũ**, không phải một bộ gõ hardcode nào đó.

---

## Lớp 5: Hook vào lifecycle (file sửa)

```go
// setModeLocked() — khi BẬT bất kỳ mode nào
h.previousInputSource = h.inputMethod.GetCurrentSourceID()  // lưu
h.inputMethod.SwitchToSourceID("com.apple.keylayout.ABC")   // chuyển

// SetModeIdle() — khi TẮT về idle
h.inputMethod.SwitchToSourceID(h.previousInputSource)       // khôi phục
```

Gọi switch *trước* khi event tap được enable — để tránh JOKey kịp intercept keystroke đầu tiên.

---

## Lớp 6: Wiring (file sửa)

`app_initialization.go` là nơi khởi tạo mọi thứ và "tiêm" (inject) các dependency vào nhau. Ở đây anh tạo `InputMethodAdapter` và truyền vào `Handler`.

---

## Tóm tắt mental model

```
Carbon API (OS)
    ↑ được gọi bởi
.m file (Objective-C)
    ↑ được wrap bởi
_darwin.go (CGO bridge, //go:build darwin)
    ↑ implement interface
InputMethodPort (contracts thuần Go)
    ↑ được dùng bởi
Handler.setModeLocked / SetModeIdle
    ↑ được inject bởi
app_initialization.go
```

Mỗi lớp chỉ biết đến lớp ngay trên nó — đây chính là Hexagonal Architecture.

---

Anh muốn bắt đầu implement từ lớp nào? Em gợi ý đi từ trong ra ngoài: Port interface → `.m` file → CGO bridge → Handler → Wiring.

# CLAUDE CODE 03

Tốt — em đã có đủ thứ để giải thích bằng code thật trong Neru, không phải ví dụ trừu tượng.

---

## Nhìn vào `mode_setup.go` thực tế

Dòng 77–85, hàm `setModeLocked()` trông như thế này:

```go
func (h *Handler) setModeLocked(appMode domain.Mode, overlayMode overlay.Mode) {
    h.setAppModeLocked(appMode)
    if h.enableEventTap != nil {
        h.enableEventTap()         // ← gọi qua function field, không phải gọi thẳng CGO
    }
    h.overlaySwitch(overlayMode)  // ← gọi qua overlayManager interface
}
```

Chú ý: không có `import "C"`, không có `#include` nào ở đây. `h.enableEventTap` chỉ là một `func()` được inject từ ngoài vào. Handler không biết bên dưới là CGO hay gì.

---

## So sánh với cách depguard sẽ chặn

Nếu anh viết thế này trong `mode_setup.go`:

```go
package modes

/*
#include <Carbon/Carbon.h>
*/
import "C"   // ← depguard sẽ báo lỗi ngay đây
```

**Tại sao depguard chặn?** Vì `mode_setup.go` thuộc package `modes` — đây là **application layer** (tầng logic nghiệp vụ). Depguard cấu hình trong Neru có quy tắc: *non-darwin-tagged code không được import platform-specific stuff*. `mode_setup.go` không có build tag `darwin`, nghĩa là nó phải compile được trên Linux và Windows. Import Carbon sẽ làm vỡ build trên Linux ngay lập tức.

---

## Nhìn vào một Port thực tế để thấy pattern

`system.go` trong `internal/core/ports/` có `SecureInputPort`:

```go
// SecureInputPort defines the interface for secure input detection
type SecureInputPort interface {
    IsSecureInputEnabled() bool
    ShowSecureInputNotification()
}
```

Interface này **không biết** về macOS API. Nhưng đâu đó trong `internal/core/infra/platform/darwin/` có một file `.go` với `//go:build darwin` implement nó bằng CGO + Objective-C.

Handler chỉ nhận `SecureInputPort` qua constructor — nó không quan tâm implementation là gì.

---

## Vậy "đường ống 3 lớp" hoạt động thế nào thực sự?

```
mode_setup.go           (package modes — application layer)
    │
    │  h.inputMethod.SwitchToSourceID("com.apple.keylayout.ABC")
    │  ↓  chỉ biết interface, không biết implementation
    │
ports/input_method.go   (contract thuần Go — không import gì cả)
    │  type InputMethodPort interface { ... }
    │
    │  ← implement bởi →
    │
infra/darwin/input_method_darwin.go   (//go:build darwin)
    │  import "C" ← OK ở đây vì file có build tag darwin
    │
    ↓
input_method_darwin.m   (Objective-C)
    TISCopyCurrentKeyboardInputSource()
    TISSelectInputSource()
```

**Điểm mấu chốt**: CGO chỉ xuất hiện ở tầng `infra/darwin/`, nơi duy nhất được phép biết về macOS. Mọi tầng trên đó chỉ nói chuyện qua interface.

---

## Câu hỏi để anh kiểm tra hiểu bài

Nếu Neru sau này muốn chạy trên Linux — bộ gõ không phải Carbon mà là iBus/fcitx — thì anh phải sửa file nào? Và file nào **không cần sửa**?

# GEMINI 03

🥷 Chào anh Vũ, đoạn từ dòng 109 đến 129 trong file `Journal.md` chính là câu trả lời xuất sắc cho câu hỏi thứ 2 mà em vừa đặt ra cho anh: **"Tại sao lại phải tạo Interface (Port) và viết nhiều file rườm rà như vậy thay vì gọi thẳng hàm hệ điều hành cho nhanh?"**

Đoạn này tác giả (Claude) đang giải thích lý do tại sao dự án Neru lại khắt khe trong việc chia tách các file (cụ thể là chia làm 6 bước như kế hoạch). Em sẽ phân tích sâu hơn đoạn này theo 3 ý chính nhé:

### 1. Sự cám dỗ của "Cách Sai" (Đường tắt)
```go
// CÁCH SAI theo kiến trúc Neru
import "C"
// #include <Carbon/Carbon.h>
func setModeLocked(...) {
    C.TISSelectInputSource(...)  // gọi thẳng
}
```
*   **Giải thích:** Nếu chỉ để giải quyết bài toán "đổi bộ gõ tiếng Việt", anh hoàn toàn có thể mở file `mode_setup.go` ra, chèn thẳng thư viện C (`import "C"`) của macOS vào, rồi gọi hàm `C.TISSelectInputSource(...)` ngay bên trong hàm `setModeLocked`. Xong! Code chỉ tốn đúng 1 file và vài dòng.
*   **Vấn đề:** Nếu làm vậy, file `mode_setup.go` (vốn là trái tim quản lý các chức năng của Neru) sẽ bị **"ô nhiễm"** bởi mã nguồn riêng của hệ điều hành Apple (Carbon API).

### 2. Sự ngăn cấm của Kiến trúc (Hexagonal Architecture & Depguard)
*   **Giải thích:** Tác giả nhấn mạnh: *"Neru cấm điều này — `depguard` sẽ báo lỗi."*
*   **Depguard là gì?** Nó là một công cụ linter (kiểm tra lỗi code tự động) được cấu hình sẵn trong dự án Neru. Công cụ này được set up một luật thép: **Bất kỳ file nào nằm trong thư mục `internal/app` hoặc `internal/core/domain` mà dám `import` các thư viện gắn mác OS (như Carbon, Cocoa, Cgo...) thì tự động báo lỗi không cho biên dịch.**
*   **Tại sao lại cấm?** Vì cốt lõi của Kiến trúc Lục giác (Hexagonal) là: *Core logic phải hoàn toàn "mù tịt" về thế giới bên ngoài*. Core không được biết nó đang chạy trên Mac, Windows hay Linux. Nếu anh nhét code macOS vào Core, mai mốt đem sang Linux biên dịch sẽ bị lỗi (panic) ngay lập tức vì Linux không có thư viện Carbon.

### 3. Đường ống 3 lớp (The Pipeline)
Để giải quyết sự ngăn cấm đó, tác giả vẽ ra một mô hình "đường ống":
```text
mode_setup.go  →  InputMethodPort  →  darwin adapter  →  Carbon API (Objective-C)
   (Go logic)      (interface)        (CGO bridge)         (.m file)
```
*   **Lớp 1 - Lệnh (mode_setup.go):** Chỉ nói: *"Ê, đổi bộ gõ cho tao!"*
*   **Lớp 2 - Hợp đồng (InputMethodPort):** Bản hợp đồng quy định: *"Ai ký vào đây thì phải biết cách đổi bộ gõ. Tôi không cần biết anh là ai."*
*   **Lớp 3 - Người thi hành (darwin adapter + .m file):** Anh thợ máy chuyên về macOS đọc hợp đồng, gật đầu nhận việc. Anh ta tự lôi đồ nghề của Apple (Carbon API, Objective-C) ra làm việc đó đằng sau hậu trường.

---

🤔 **Câu hỏi Socratic cho anh:**
Theo đường ống 3 lớp ở trên, giả sử bây giờ anh muốn làm thêm bản Neru cho Windows (chuyển bộ gõ bằng API của Windows), thì theo anh, chúng ta có cần phải sửa lại dù chỉ một dòng code nào trong file `mode_setup.go` không? Tại sao?

# GEMINI 04

🥷 Tuyệt vời anh Vũ! Claude đã tạo ra một bộ khung hoàn hảo cho phần Adapter trên macOS. Nhìn vào 4 file này, anh sẽ thấy rõ nhất cách mà Golang tương tác với các hệ điều hành thông qua Cgo (C-Go).

Để em giải thích từng file cho anh theo thứ tự của "đường ống" (từ Go đi xuống macOS) nhé:

### 1. `ports/input_method.go` (Bản Hợp Đồng)
```go
type InputMethodPort interface {
	GetCurrentSourceID() string
	SwitchToSourceID(id string) error
}
```
*   **Giải thích:** Đây là file thuần Go, nằm ở Core (trung tâm). Nó định nghĩa một "bản hợp đồng" yêu cầu phải có 2 kỹ năng: lấy ID bộ gõ và đổi bộ gõ.
*   **Vai trò:** Giúp Core không cần biết macOS là gì, cứ ai ký hợp đồng này (Adapter) thì Core sẽ xài.

---

### 2. `darwin/inputmethod.go` (Người Ký Hợp Đồng - CGO Bridge)
```go
//go:build darwin
...
/*
#include "inputmethod.h"
#include <stdlib.h>
*/
import "C"
...
var _ ports.InputMethodPort = (*InputMethodAdapter)(nil)
```
*   **Giải thích:**
    *   `//go:build darwin`: Lệnh này nói với compiler rằng file này CHỈ được biên dịch nếu đang chạy trên Mac (Darwin).
    *   `import "C"`: Đây là Cgo. Nó cho phép Go "nói chuyện" được với code C. Khối comment ngay phía trên nó (`#include...`) chính là code C thật được nhúng vào.
    *   `var _ ports.InputMethodPort = ...`: Đây là một mẹo nhỏ (compile-time check) rất hay dùng trong Go. Nó ép trình biên dịch kiểm tra xem cái `InputMethodAdapter` này đã code đủ 2 hàm trong hợp đồng (`GetCurrentSourceID` và `SwitchToSourceID`) chưa. Nếu thiếu, code sẽ báo lỗi ngay khi gõ chứ không đợi đến lúc chạy.
*   **Vai trò:** Nó là cầu nối. Khi Core gọi hàm của Go, file này sẽ dịch lệnh đó sang C và gọi xuống `C.getInputSourceID()`.

---

### 3. `darwin/inputmethod.h` (Biển báo C)
```c
const char* getInputSourceID(void);
int switchInputSourceByID(const char* sourceID);
```
*   **Giải thích:** File này gọi là file Header (`.h`) của ngôn ngữ C. Nó không chứa logic, nó chỉ "đăng bảng thông báo": *"Ê, tui có 2 hàm này nhé, hàm thứ nhất trả về chuỗi chữ (char*), hàm thứ hai nhận chuỗi chữ và trả về số nguyên (int)."*
*   **Vai trò:** Cgo trong file Go không thể "nhìn thẳng" vào file Objective-C (`.m`), nó chỉ hiểu code C. Nên file `.h` này làm nhiệm vụ mô tả các hàm C để file Go biết đường mà gọi.

---

### 4. `darwin/inputmethod_darwin.m` (Công nhân gánh tạ - Objective-C)
```objc
#import <Carbon/Carbon.h>
...
const char* getInputSourceID(void) { ... }
int switchInputSourceByID(const char* sourceID) { ... }
```
*   **Giải thích:** Đây là file Objective-C, nơi làm việc dơ bẩn nhất. Nó "import" framework `Carbon` của Apple vào.
    *   Hàm `getInputSourceID`: Nó gọi `TISCopyCurrentKeyboardInputSource` của Apple để lấy thông tin bộ gõ, lôi cái ID ra (ví dụ: `com.apple.keylayout.ABC`), biến đổi chuỗi của Apple (CFString) thành chuỗi cơ bản của C (char*) để gửi ngược lên cho Go. Cực kỳ cồng kềnh vì phải tự dọn dẹp bộ nhớ (malloc/free/CFRelease).
    *   Hàm `switchInputSourceByID`: Nó xin Apple danh sách tất cả bộ gõ đang cài (`TISCreateInputSourceList`), chạy vòng lặp tìm cái bộ gõ nào có ID trùng với ID mà Go truyền xuống. Nếu thấy, nó gọi lệnh thiêng liêng `TISSelectInputSource` để kích hoạt bộ gõ đó.
*   **Vai trò:** Thực thi mệnh lệnh trực tiếp trên hệ điều hành.

---

**Tóm lại đường ống này chạy như sau:**
**Go (Core)** $\rightarrow$ **Cgo (Go file)** $\rightarrow$ **C (Header .h)** $\rightarrow$ **Objective-C (.m file) (Carbon API)**.

Khi dữ liệu trả về, nó đi ngược lại từ Objective-C lên Cgo, rồi file Cgo sẽ giải phóng bộ nhớ C (bằng hàm `defer C.free...`) và biến chuỗi C thành chuỗi Go (`C.GoString`) để trả về cho phần Core gọn gàng sạch sẽ.

Anh Vũ đã nắm rõ bức tranh toàn cảnh của 4 file này chưa ạ? Bước tiếp theo (Bước 4 & 5), ta sẽ gắn cái "đường ống" này vào trái tim của Neru (`mode_setup.go`) nhé. Anh muốn tự sửa hay muốn em làm ạ?

# CLAUDE CODE 04

## Bug phát sinh sau khi thêm tính năng

Feature vừa merge xong, test cơ bản qua — activate Neru từ JOKey, bộ gõ chuyển sang ABC, deactivate, JOKey quay về. Nhưng anh Vũ phát hiện một case lạ:

> *"Khi đang ở bộ gõ JOKey, activate Neru vào recursive-grid, rồi **chưa deactivate** mà dùng phím tắt chuyển sang grid hoặc hints — thấy bộ gõ chuyển rất nhanh nhưng vẫn ở ABC. Rồi deactivate Neru, JOKey không quay về. Nó dính ở ABC luôn."*

Đây là một **mode-to-mode transition bug** — case mà integration test không cover, vì test thường chỉ kiểm tra activate → deactivate theo cặp thẳng.

---

## Phân tích ban đầu (/think)

Nhìn vào flow khi chuyển từ recursive-grid sang grid:

```
activateGridModeWithAction()
    └─ h.exitModeLocked()          ← dọn dẹp recursive-grid
        └─ performCommonCleanup()
            ├─ SwitchToSourceID(previousInputSource)  ← restore JOKey
            └─ h.previousInputSource = ""             ← xóa trắng
    └─ h.setModeLocked(ModeGrid, ...)
        ├─ h.previousInputSource = GetCurrentSourceID()  ← đọc lại OS ???
        └─ SwitchToSourceID("ABC")
```

Nhìn vào `setModeLocked` (file `mode_setup.go`):

```go
func (h *Handler) setModeLocked(appMode domain.Mode, overlayMode overlay.Mode) {
    h.setAppModeLocked(appMode)
    if h.inputMethod != nil {
        // BUG ở đây: không có guard, luôn ghi đè previousInputSource
        h.previousInputSource = h.inputMethod.GetCurrentSourceID()
        _ = h.inputMethod.SwitchToSourceID("com.apple.keylayout.ABC")
    }
    ...
}
```

Khi đang ở mode A, `previousInputSource` đang giữ "JOKey". Khi kích hoạt mode B, code gọi `exitModeLocked()` để dọn A, rồi gọi `setModeLocked()` cho B. Nhưng `setModeLocked` **luôn ghi đè** `previousInputSource` bằng giá trị đọc từ OS — lúc này OS đang ở ABC — nên "JOKey" bị mất.

**Hypothesis (lần 1):** Thêm guard `if h.previousInputSource == ""` vào `setModeLocked` để chỉ lưu lần đầu:

```go
if h.previousInputSource == "" {
    h.previousInputSource = h.inputMethod.GetCurrentSourceID()
}
```

Logic: nếu `previousInputSource` đã có giá trị (đang trong mode nào đó), không cần lưu lại. Chỉ lưu khi thực sự "từ idle bước vào mode đầu tiên".

---

## Fix lần 1 — và test thấy vẫn sai

Sau khi thêm guard, test lại: bug vẫn còn. Anh Vũ xác nhận: *"Vẫn không được em, anh dùng các phím tắt chuyển qua giữa các chế độ là JOKey sẽ bị mất."*

Rõ ràng hypothesis lần 1 chưa đủ. Cái guard `== ""` đúng về ý tưởng nhưng thiếu một mắt xích. Phải đào sâu hơn.

# CLAUDE CODE 05

## Hunt: đào sâu vào timing của dispatch_async

Để hiểu tại sao fix lần 1 không đủ, phải đọc kỹ lại từng bước trong `performCommonCleanup`:

```go
// cleanup.go
func (h *Handler) performCommonCleanup() {
    ...
    if h.inputMethod != nil && h.previousInputSource != "" {
        _ = h.inputMethod.SwitchToSourceID(h.previousInputSource)  // (1)
        h.previousInputSource = ""                                   // (2)
    }
    ...
}
```

Nhìn qua thì có vẻ ổn: bước (1) restore JOKey, bước (2) xóa `previousInputSource`. Sau đó `setModeLocked` chạy, thấy `previousInputSource == ""`, gọi `GetCurrentSourceID()` để lưu... đợi — **OS đang ở bộ gõ nào lúc đó?**

Kéo xuống xem implementation của `SwitchToSourceID` trong file Objective-C (`inputmethod_darwin.m`):

```objc
int switchInputSourceByID(const char* sourceID) {
    ...
    dispatch_async(dispatch_get_main_queue(), ^{
        TISSelectInputSource(src);   // ← chạy ASYNC trên main thread
        CFRelease(src);
    });
    return 0;
}
```

**Đây rồi.** `dispatch_async` là **bất đồng bộ**. Nó không chờ `TISSelectInputSource` chạy xong — nó chỉ *đẩy lệnh vào hàng đợi* của main thread rồi trả về ngay lập tức.

---

## Toàn bộ race condition, từng bước một

```
Thread Go (goroutine):                  Main thread macOS:
─────────────────────────────────────   ──────────────────────────────
performCommonCleanup()
  SwitchToSourceID("JOKey")
    dispatch_async(^{ TISSelect... })   ← đẩy vào queue, CHƯA chạy
    return 0                            ← Go thread tiếp tục ngay
  previousInputSource = ""
  (mode → idle)

setModeLocked(ModeGrid)
  previousInputSource == "" → true
  GetCurrentSourceID()
    → đọc OS: vẫn đang là "ABC"  ←     ← lệnh TISSelect vẫn nằm trong queue
  previousInputSource = "ABC"  ← ĐÂY LÀ LỖI
  SwitchToSourceID("ABC")             
                                        TISSelectInputSource("JOKey")  ← quá muộn
                                        (OS chuyển sang JOKey một chớp rồi ABC ghi đè)
```

Kết quả: khi deactivate Neru, `previousInputSource = "ABC"`, nên restore về "ABC" thay vì "JOKey".

**Root cause một câu:** `performCommonCleanup` xóa `previousInputSource` trước khi lệnh switch OS thực sự thực thi (do `dispatch_async`), khiến `setModeLocked` đọc giá trị bộ gõ sai từ OS và lưu "ABC" vào `previousInputSource`.

---

## Fix đúng: savedInputSource pattern

Guard `== ""` trong `setModeLocked` là cần thiết nhưng chưa đủ. Vấn đề là `previousInputSource` bị xóa về `""` *trước khi* `setModeLocked` được gọi, nên guard không có tác dụng — `previousInputSource` đã là `""` rồi, guard sẽ trigger đọc OS, và OS lúc đó vẫn là ABC.

Giải pháp: **đừng để `previousInputSource` mất đi trong quá trình transition**. Lưu nó lại trước khi gọi `exitModeLocked()`, rồi trả lại sau:

```go
// Trong activateGridModeWithAction, activateHintModeInternal,
// activateRecursiveGridModeWithAction, StartInteractiveScroll
// — ở nhánh !isRefresh (chuyển từ mode khác sang)

savedInputSource := h.previousInputSource   // lưu "JOKey" trước
h.exitModeLocked()                          // xóa previousInputSource, queue async restore
if savedInputSource != "" {
    h.previousInputSource = savedInputSource // trả lại "JOKey"
}
// → setModeLocked thấy previousInputSource != "" → guard bỏ qua GetCurrentSourceID()
// → "JOKey" được giữ nguyên đến lần deactivate thật sự
```

Với pattern này, `setModeLocked` chỉ gọi `GetCurrentSourceID()` khi thật sự cần — khi đi từ idle vào mode lần đầu. Còn khi chuyển mode-to-mode, `previousInputSource` đã có giá trị, guard bỏ qua hoàn toàn, không bao giờ đọc giá trị stale từ OS.

---

## Áp dụng vào tất cả 4 điểm activation

Pattern này phải áp dụng nhất quán ở mọi nơi gọi `exitModeLocked()` trong quá trình activation (không phải lúc exit thật sự):

| File | Hàm | Nhánh |
|------|-----|-------|
| `modes/hints.go` | `activateHintModeInternal` | `!isRefresh && !preserveActionMode` |
| `modes/grid.go` | `activateGridModeWithAction` | `!isRefresh` |
| `modes/recursive_grid.go` | `activateRecursiveGridModeWithAction` | `!isRefresh` |
| `modes/scroll.go` | `StartInteractiveScroll` | trước `exitModeLocked()` |

Ví dụ trong `hints.go` (có comment giải thích lý do):

```go
} else {
    // Preserve previousInputSource across mode transitions.
    // exitModeLocked → performCommonCleanup clears previousInputSource after
    // queuing an async input-source restore. If we let setModeLocked read
    // GetCurrentSourceID() immediately after, the async switch hasn't run yet
    // and it reads the wrong (current nav-mode) source. By saving and
    // re-applying the value, setModeLocked skips the stale read entirely.
    savedInputSource := h.previousInputSource
    h.exitModeLocked()
    if savedInputSource != "" {
        h.previousInputSource = savedInputSource
    }
}
```

---

## Bài học rút ra

**1. dispatch_async không phải là "làm xong rồi trả về"**

Khi một hàm Go gọi xuống Objective-C và implementation dùng `dispatch_async`, hàm Go nhận được `return 0` *trước khi* side effect thực sự xảy ra. Bất kỳ code Go nào đọc lại trạng thái OS ngay sau đó đều có thể nhận giá trị cũ. Đây là một **async timing hazard** cực kỳ phổ biến khi bridge giữa Go và Objective-C.

**2. Không bao giờ đọc lại state từ OS ngay sau khi ra lệnh thay đổi state đó**

Thay vì đọc lại (`GetCurrentSourceID()`), hãy **nhớ giá trị cần thiết trước khi thay đổi**. Pattern `savedInputSource` chính xác là nguyên tắc này: lưu cái gì cần giữ *trước* khi code cleanup chạy, đừng cố lấy lại từ OS *sau*.

**3. isRefresh vs mode-to-mode transition là hai case hoàn toàn khác nhau**

- `isRefresh = true`: cùng mode tái kích hoạt (ví dụ: `--repeat`). Không cần `exitModeLocked()` đầy đủ — mode, event tap đã đúng trạng thái.
- `isRefresh = false` + đang trong mode khác: đây mới là mode-to-mode transition, phải dùng savedInputSource pattern.
- `isRefresh = false` + đang idle: activation lần đầu từ idle, `previousInputSource == ""`, `setModeLocked` đọc OS bình thường.

Ba case này trông giống nhau nhưng cần xử lý khác nhau. Bug xuất hiện vì case giữa bị gộp nhầm vào case cuối.

**4. Test integration thường bỏ sót case chuyển mode không qua idle**

Test thông thường kiểm tra: idle → mode A → idle. Bug này chỉ xuất hiện ở: idle → mode A → mode B → idle. Đây là một lời nhắc nhở để viết test cho các *đường chuyển đổi* (transition paths), không chỉ các *trạng thái đầu-cuối* (states).

---

**Commit:** `20fd607` — `fix: preserve input source across mode-to-mode transitions`
