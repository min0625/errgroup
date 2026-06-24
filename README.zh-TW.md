# errgroup
[![Go Reference](https://pkg.go.dev/badge/github.com/min0625/errgroup.svg)](https://pkg.go.dev/github.com/min0625/errgroup)
[![codecov](https://codecov.io/gh/min0625/errgroup/branch/main/graph/badge.svg)](https://codecov.io/gh/min0625/errgroup)

[English](README.md) | **繁體中文**

可直接替換 `golang.org/x/sync/errgroup` 的套件,會回收 goroutine 中的 panic,並在 `Wait` 呼叫中重新拋出。

參考:https://github.com/golang/go/issues/53757

## 套件

| 套件 | 說明 |
|---|---|
| `github.com/min0625/errgroup` | `golang.org/x/sync/errgroup` 的直接替換版,額外提供 panic 回收 |
| [`github.com/min0625/errgroup/x/errgroup`](x/errgroup/README.zh-TW.md) | 情境感知(context-aware)變體 — 直接將 `context.Context` 傳入每個 goroutine 函式 |

## 安裝
```sh
go get github.com/min0625/errgroup
```

## 範例
```go
package main

import (
	"fmt"

	"github.com/min0625/errgroup"
)

func main() {
	// 此範例使用 "github.com/min0625/errgroup",它會捕捉 panic。
	// 若改為匯入 "golang.org/x/sync/errgroup",則不會捕捉 panic。
	// 你可以在 Go Playground 試玩:https://go.dev/play/p/7pUX6uQ2mCH
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

	g.Go(func() error {
		// 做些事情
		return nil
	})

	g.Go(func() error {
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
