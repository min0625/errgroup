# errgroup/x/errgroup
[![Go Reference](https://pkg.go.dev/badge/github.com/min0625/errgroup/x/errgroup.svg)](https://pkg.go.dev/github.com/min0625/errgroup/x/errgroup)
[![codecov](https://codecov.io/gh/min0625/errgroup/branch/main/graph/badge.svg)](https://codecov.io/gh/min0625/errgroup)

[English](README.md) | **繁體中文**

[`github.com/min0625/errgroup`](../../README.zh-TW.md) 的情境感知(context-aware)變體,會將衍生的 `context.Context` 直接傳入每個 goroutine 函式,免去透過閉包(closure)擷取 context 的需要。

## 與根套件的差異

| | `github.com/min0625/errgroup` | `github.com/min0625/errgroup/x/errgroup` |
|---|---|---|
| Goroutine 函式簽章 | `func() error` | `func(context.Context) error` |
| 取得 context | 透過閉包擷取 | 以參數傳入 |
| 建構子 | `WithContext(ctx)` 回傳 `(*Group, context.Context)` | `New(ctx)` 回傳 `*Group` |
| 零值 | 有效,無 context 取消機制 | 有效,使用 `context.Background()` |

## 安裝

```sh
go get github.com/min0625/errgroup/x/errgroup
```

## 範例

```go
package main

import (
	"context"
	"fmt"

	"github.com/min0625/errgroup/x/errgroup"
)

func main() {
	// 此範例使用 "github.com/min0625/errgroup/x/errgroup",它會捕捉 panic。
	// 若改為匯入 "golang.org/x/sync/errgroup",則不會捕捉 panic。
	var g errgroup.Group

	defer func() {
		// 會捕捉到 panic。
		if p := recover(); p != nil {
			switch t := p.(type) {
			case errgroup.PanicValue:
				fmt.Println(t.Recovered)
			case errgroup.PanicError:
				fmt.Println(t.Recovered)
			}
		}
	}()

	g.Go(func(_ context.Context) error {
		// 做些事情
		return nil
	})

	g.Go(func(_ context.Context) error {
		panic("oops")
	})

	if err := g.Wait(); err != nil {
		// 處理錯誤
		fmt.Println(err)
		return
	}

	// Output: oops
}
```

## Panic 行為

由 `Go` 或 `TryGo` 啟動的 goroutine 若發生 panic,會被捕捉並在 `Wait` 中重新拋出,並包裝為:

- `PanicError` — 當 panic 的值實作了 `error` 介面時
- `PanicValue` — 其他所有情況

兩種型別都會提供 `Stack` 欄位,內含 panic 發生當下所擷取的堆疊追蹤(stack trace)。

若有多個 goroutine 同時 panic,只有第一個 panic 會被傳播,其餘會被靜默丟棄。

## 零值注意事項

零值的 `Group`(未透過 `New` 建立)是有效的,並會在第一次呼叫 `Go`、`TryGo`、`SetLimit` 或 `Wait` 時自動初始化。但其基礎 context 會固定為 `context.Background()`。若要使用自訂的可取消 context,請務必以 `New(ctx)` 建立群組。
