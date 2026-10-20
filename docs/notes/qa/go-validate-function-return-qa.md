# Q&A：確かめる関数は、何を返すとよいか（`validMethod` への代入、`error` を返す形、型で表す形）

udp-echo の Client（`playground/udp-echo/client/main.go`）で、サブコマンドを確かめる関数 `validateMethod` を書いたときに出た質問と、その回答をまとめたメモです。

> このメモには、説明のための**コードの例**を載せています。Client のコードは自分で書く方針なので（[playground/README.md](../../../playground/README.md)）、例はそのまま貼り付けるためではなく、**考え方を見るため**のものです。例は、Claude が Mac（Go 1.26.4）で `go vet` と実行をして確かめました。

> **このドキュメントで一番大事なこと**
>
> - `validMethod := method` は、**同じ値・同じ型のコピー**を作るだけ。名前に `valid` と付けても、コンパイラは何も確かめないので、**可読性はほとんど上がらない**
> - 「確かめ終わった値だ」と伝えたいなら、**関数の名前と戻り値**で伝える
> - Go でよくある形は、確かめる関数が **`error` を返し**、表示や終了は**呼び出した側（`main`）が決める**形
> - さらに一歩進むと、文字列を**意味のある型**（`Method` など）に**変換して返す**形がある。比較が1か所にまとまり、**打ち間違いをコンパイラが見つけてくれる**
> - 自分で型を作るときは、**ゼロ値（0）を「まだ決まっていない」**にしておくと、エラーのときに間違って使われにくい

関連ドキュメント：

- [udp-echo-client-review.md](../reviews/udp-echo-client-review.md)：Client のレビュー記録（#22：サブコマンドの名前、#26：サブコマンドのチェックの重複）
- [stdout-stderr-qa.md](stdout-stderr-qa.md)：標準エラー出力、終了コード、`os.Exit` すると `defer` が動かないこと（7-3）
- [go-ip-address-parse-qa.md](go-ip-address-parse-qa.md)：`net.ParseIP`・`netip.ParseAddr`（「確かめる」ではなく「変換する」関数の例）

## 目次

- [質問](#質問)
- [1. 結論](#1-結論)
- [2. 今の書き方](#2-今の書き方)
- [3. なぜ効果が薄いのか](#3-なぜ効果が薄いのか)
- [4. 書き方1：そのまま返す](#4-書き方1そのまま返す)
- [5. 書き方2：何も返さない](#5-書き方2何も返さない)
- [6. 書き方3：`error` を返す](#6-書き方3error-を返す)
- [7. 書き方4：確かめた結果を、別の型で返す](#7-書き方4確かめた結果を別の型で返す)
- [8. 4つを比べる](#8-4つを比べる)
- [9. Claude が Mac で確かめた結果](#9-claude-が-mac-で確かめた結果)
- [10. まとめ](#10-まとめ)

---

## 質問

> **Q1.** `validateMethod()` の中で、確かめた後の `method` を、わざわざ「確かめた後の変数」ということで `validMethod` に代入して、それを返している。これは正しいプログラムの書き方なのか？ 可読性のためにこのように書いたが、どうか？

---

## 1. 結論

| | 答え |
|---|---|
| 正しいか | **間違いではない**。ちゃんと動く |
| 可読性は上がるか | **ほとんど上がらない**。`validMethod` は `method` のただのコピーで、コンパイラは何も確かめない。読む人が「なぜ別の変数に？」と迷うこともある |
| どう書くとよいか | すぐ直すなら、**`return method` だけ**にする。Go らしく書くなら、**`error` を返す**形にして、さらに**意味のある型に変換して返す**形にする |

---

## 2. 今の書き方

```go
// validate method
func validateMethod(method string) string {
	if method != "0" && method != "1" {
		fmt.Fprintln(os.Stderr, "Invalid argument. The argument must be 0 (unicast) or 1 (broadcast).")
		os.Exit(1)
	}
	validMethod := method // ← ここ
	return validMethod
}
```

```go
func main() {
	// …
	validMethod := validateMethod(os.Args[1])
	subCmds := setSubCmds(validMethod, os.Args[2:])
	// …
}
```

この関数は、3つのことをしています。

| していること | 行 |
|---|---|
| ① 確かめる | `if method != "0" && method != "1"` |
| ② だめなら、表示して**終了する** | `fmt.Fprintln(…)`、`os.Exit(1)` |
| ③ 確かめた値を返す | `validMethod := method`、`return validMethod` |

質問は③についてですが、②にも Go らしい書き方があるので、一緒に見ていきます。

---

## 3. なぜ効果が薄いのか

### 3-1. 新しい変数は、ただのコピー

```go
validMethod := method
```

この1行で起きるのは、**同じ文字列を、別の名前の変数に入れる**ことだけです。

```text
 method      ──▶ "0"
 validMethod ──▶ "0"   ← 中身も型（string）も同じ
```

### 3-2. 名前に `valid` と付けても、何も守られない

変数の名前は、**人間のためのメモ**です。コンパイラは名前の意味を見ません。

```go
validMethod := "abc" // ← 確かめていない値でも、valid という名前の変数に入れられる
```

しかも、`validMethod` は関数の**中だけ**の変数なので、関数を出たら名前は消えます。呼び出す側から見ると、戻り値は**ただの `string`** です。

```text
 関数の中   ：validMethod という名前がある（でも中身はただの string）
 関数の外   ：string が1つ返ってくるだけ。「確かめ済み」という情報は残らない
```

### 3-3. 読む人が迷う

読む人は、別の変数に入れているのを見ると、「**何か変換しているのかな？**」「**元の `method` と何が違うんだろう？**」と考えます。何も変わっていないと分かるまで、少し時間を使わせてしまいます。

### 3-4. 「確かめた」ことは、どこで伝えるのか

| 伝える場所 | 例 |
|---|---|
| **関数の名前** | `validate〜`（確かめる）、`parse〜`（変換する） |
| **戻り値の型** | `error` を返す → 「失敗することがある関数だ」と分かる。自分で作った型を返す → 「もう確かめ済みの値だ」と分かる |

関数の中の変数名ではなく、**関数の外から見えるもの**で伝える、というのがポイントです。

---

## 4. 書き方1：そのまま返す

一番簡単な直し方です。

```go
func validateMethod(method string) string {
	if method != "0" && method != "1" {
		fmt.Fprintln(os.Stderr, "Invalid argument. The argument must be 0 (unicast) or 1 (broadcast).")
		os.Exit(1)
	}
	return method
}
```

- 動きは今とまったく同じです
- 1行減って、「何か変換している？」と迷わせることがなくなります

---

## 5. 書き方2：何も返さない

「確かめるだけ」の関数にする形です。

```go
func validateMethod(method string) {
	if method != "0" && method != "1" {
		fmt.Fprintln(os.Stderr, "Invalid argument. The argument must be 0 (unicast) or 1 (broadcast).")
		os.Exit(1)
	}
}
```

```go
func main() {
	// …
	validateMethod(os.Args[1])                // 確かめる（だめならここで終了）
	subCmds := setSubCmds(os.Args[1], os.Args[2:]) // 元の値をそのまま使う
	// …
}
```

- 戻り値が受け取った値と同じなら、**返す必要がない**、という考え方です
- 8回目のレビュー時点の `validateSubCmds` も、この形（何も返さない）になっています

---

## 6. 書き方3：`error` を返す

### 6-1. 今の形の弱点：関数の中で終了している

今の `validateMethod` は、関数の中で `os.Exit(1)` しています。小さなプログラムなら動きますが、弱点があります。

| 弱点 | 説明 |
|---|---|
| 使い回しにくい | 呼んだだけで**プログラムが終わる**。「だめなら別の方法を試す」「エラーをまとめて表示する」ができない |
| 試しにくい | テストで呼ぶと、テストのプログラムごと終わってしまう |
| `defer` が動かない | `os.Exit` は、予約しておいた `defer` を**実行せずに**終わる（[stdout-stderr-qa.md](stdout-stderr-qa.md) の7-3） |

### 6-2. Go でよくある形

Go では、「確かめる関数は **`error` を返すだけ**にして、表示や終了は**呼び出した側が決める**」形がよく使われます。

```go
// validateMethod は、method が正しいか確かめる。
// 正しくなければ error を返す（終了はしない）。
func validateMethod(method string) error {
	if method != "unicast" && method != "broadcast" {
		return fmt.Errorf("invalid method %q: must be unicast or broadcast", method)
	}
	return nil
}
```

```go
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: client <unicast|broadcast> [flags]")
		os.Exit(1)
	}

	if err := validateMethod(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err) // 表示するのは main
		os.Exit(1)                   // 終了するのも main
	}
	// …
}
```

```text
 validateMethod の仕事：確かめて、だめなら error を返す
 main の仕事         ：error を受け取って、表示して終了する
```

| 部品 | 意味 |
|---|---|
| `error` | Go の「失敗したこと」を表す型。`nil` なら失敗していない |
| `fmt.Errorf` | 書式付きで `error` を作る関数。`%q` は、値を `"…"` で囲んで表示する（空の文字列も `""` と見える） |
| `if err := …; err != nil { … }` | 呼んで、エラーがあれば処理する、という Go の定番の書き方 |

> `net.ResolveUDPAddr` や `conn.WriteToUDP` も、この形（値と `error` を返し、どうするかは呼び出した側が決める）になっています。自分で書く関数も、標準ライブラリと同じ形にそろえると、読む人が迷いません。

### 6-3. 例ではサブコマンドを `unicast`・`broadcast` にした

例では、`"0"`・`"1"` ではなく `"unicast"`・`"broadcast"` で比べています。レビュー記録の #22 のとおり、**使う人に意味が伝わる名前**にした方が分かりやすいからです。

---

## 7. 書き方4：確かめた結果を、別の型で返す

### 7-1. 考え方

「確かめ終わった」ことを、**型で表す**方法です。

```text
 文字列 "unicast" / "broadcast"
        │
        ▼
 parseMethod（変換する関数）
        │
        ├── 正しい   → Method 型の値（Unicast / Broadcast）
        └── 正しくない → error
```

関数の名前も、「確かめる（validate）」ではなく「**変換する（parse）**」にします。`net.ParseIP`（文字列 → `net.IP`）と同じ考え方です。

### 7-2. 型と定数を作る

```go
// Method は送り方を表す型。
type Method int

const (
	MethodUnknown Method = iota // 0：まだ決まっていない（ゼロ値）
	Unicast                     // 1
	Broadcast                   // 2
)
```

| 部品 | 意味 |
|---|---|
| `type Method int` | 中身は `int` だが、**別の型**として扱われる新しい型を作る |
| `const ( … )` | 定数をまとめて作る |
| `iota` | `const` の中で、0, 1, 2 … と**順番に数を振る**しくみ |

> **ゼロ値を「まだ決まっていない」にする理由**
>
> Go の変数は、何も入れないと**ゼロ値**（`int` なら 0）になります。もし `Unicast` を 0 にすると、**何も入れていない変数や、エラーのときに返した 0 が、ユニキャストとして扱われてしまいます**。0 を `MethodUnknown`（まだ決まっていない）にしておけば、うっかり使われても「決まっていない」と分かります。

### 7-3. 変換する関数

```go
// parseMethod は、文字列を Method に変換する。
// 知らない文字列なら error を返す。
func parseMethod(s string) (Method, error) {
	switch s {
	case "unicast":
		return Unicast, nil
	case "broadcast":
		return Broadcast, nil
	default:
		return MethodUnknown, fmt.Errorf("invalid method %q: must be unicast or broadcast", s)
	}
}
```

- **文字列と比べるのは、この関数の中だけ**です
- エラーのときは、`MethodUnknown`（ゼロ値）と `error` を返します

### 7-4. 使う側

```go
func main() {
	// …（長さのチェック）

	method, err := parseMethod(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// ここから先は、文字列ではなく Method 型で扱う
	if method == Broadcast {
		// …
	}
}
```

flag の組を作る関数なども、`string` ではなく **`Method` を受け取る**ようにします。

```go
func setSubCmds(method Method, args []string) networkArgs {
	// method == Unicast / method == Broadcast で分ける
	// …
}
```

### 7-5. よいところ：打ち間違いをコンパイラが見つける

文字列で比べていると、打ち間違いは**実行するまで分かりません**。

```go
if method == "unicst" { // ← 打ち間違い。でもコンパイルは通る。いつも false になるだけ
```

`Method` 型と定数で比べると、**コンパイルの時点で**エラーになります。Mac で試すと、次のようになりました。

```go
fmt.Println(m == Unicst)    // 定数の名前の打ち間違い
fmt.Println(m == "unicast") // Method と文字列を比べる
```

```text
./main.go:16:19: undefined: Unicst
./main.go:17:19: invalid operation: m == "unicast" (mismatched types Method and untyped string)
```

| | 文字列で比べる | `Method` 型で比べる |
|---|---|---|
| 打ち間違い | 実行してみないと分からない | **コンパイルエラー**で分かる |
| 比べる場所 | 使うところ全部（`"0"` が何か所にも出てくる） | **`parseMethod` の中の1か所だけ** |
| 呼び出す側から見て | ただの `string`。確かめ済みか分からない | `Method` なら、**`parseMethod` を通った値**だと分かる |
| 種類が増えたとき（例：マルチキャスト） | 比べている場所を全部探して直す | 定数を1つ足し、`parseMethod` に1行足す |

### 7-6. おまけ：表示用のメソッド

`Method` は中身が `int` なので、そのまま表示すると `1` や `2` になります。`String` メソッドを付けると、`fmt.Println` などで**名前で表示**されます。

```go
func (m Method) String() string {
	switch m {
	case Unicast:
		return "unicast"
	case Broadcast:
		return "broadcast"
	default:
		return "unknown"
	}
}
```

```text
 fmt.Println(Unicast)        → unicast
 fmt.Printf("%d", Unicast)   → 1
 var zero Method; fmt.Println(zero) → unknown
```

> `String() string` というメソッドを持つ型は、`fmt` が自動で使ってくれます（`fmt.Stringer` というインターフェース）。

---

## 8. 4つを比べる

| 書き方 | 戻り値 | 終了するのは | よいところ | 気になるところ |
|---|---|---|---|---|
| 今（`validMethod` に入れて返す） | `string` | 関数の中 | — | コピーするだけの1行がある |
| 1. そのまま返す | `string` | 関数の中 | すぐ直せる | 関数の中で終了する |
| 2. 何も返さない | なし | 関数の中 | 「確かめるだけ」と分かる | 関数の中で終了する |
| 3. `error` を返す | `error` | **`main`** | Go の定番の形。使い回せる、試せる | 呼び出す側に `if err != nil` が要る |
| 4. 型に変換して返す | `Method` と `error` | **`main`** | 比べる場所が1か所。打ち間違いをコンパイラが見つける | 型と定数を作る分、コードが少し増える |

**おすすめ**

- すぐ直すなら、**書き方1**で十分です
- Go らしい書き方を練習したいなら、**書き方3と4を組み合わせる**（`parseMethod` が `Method` と `error` を返す）のがおすすめです。レビュー記録の #22（サブコマンドの名前）・#26（チェックを1か所にまとめる）とも一緒に直せます
- `validateSubCmds`（IPアドレス・種類・ID のチェック）も、同じように**`error` を返す形**にできます

---

## 9. Claude が Mac で確かめた結果

7章の `Method` 型・`parseMethod`・`String` メソッドと、6章の `error` を返す `validateMethod` を、一時的な場所に小さなプログラムとして書いて、Mac（Go 1.26.4）で確かめました。`go vet` は問題なしでした。

| 渡した文字列 | `parseMethod` の結果 | `validateMethod` の結果 |
|---|---|---|
| `"unicast"` | `unicast`（中身は 1）、エラーなし | エラーなし |
| `"broadcast"` | `broadcast`（中身は 2）、エラーなし | エラーなし |
| `"0"` | `unknown`（中身は 0）、`invalid method "0": must be unicast or broadcast` | 同じエラー |
| `""`（空） | `unknown`（中身は 0）、`invalid method "": must be unicast or broadcast` | 同じエラー |

| そのほか | 結果 |
|---|---|
| 何も入れていない `Method` の変数を表示 | `unknown`（ゼロ値が `MethodUnknown`） |
| 引数なしで起動 | `usage: client <unicast\|broadcast> [flags]` を標準エラー出力に出して、終了コード 1 |
| 定数名の打ち間違い（`Unicst`） | コンパイルエラー：`undefined: Unicst` |
| `Method` と文字列を比べる | コンパイルエラー：`mismatched types Method and untyped string` |

---

## 10. まとめ

- `validMethod := method` は**ただのコピー**。名前に `valid` と付けても、コンパイラは何も確かめないし、関数の外には伝わらない。**間違いではないが、可読性はほとんど上がらない**
- 「確かめた」ことは、関数の中の変数名ではなく、**関数の名前と戻り値**で伝える
- すぐ直すなら **`return method`**
- Go の定番は、確かめる関数が **`error` を返し**、表示や終了は **`main` が決める**形。関数の中で `os.Exit` すると、使い回しにくく、`defer` も動かない
- さらに進むなら、**意味のある型（`Method`）に変換して返す**（`parseMethod`）。文字列と比べるのは1か所だけになり、**打ち間違いはコンパイラが見つける**
- 自分で型を作るときは、**ゼロ値を「まだ決まっていない」**にしておく
