# Q&A：IPアドレスの文字列を確かめるには（`net.ParseIP`・`netip.ParseAddr`、`ResolveUDPAddr` との違い）

udp-echo の Client で、宛先の IPアドレスを引数（`-addr`）で受け取るようにしたあとに出た質問と、その回答をまとめたメモです。

> **このドキュメントで一番大事なこと**
>
> - IPアドレスの文字列を解釈する関数は2系統ある：**`net.ParseIP`**（`net` パッケージ）と **`netip.ParseAddr`**（`net/netip` パッケージ）
> - `net.ParseIP` は、正しくない文字列なら **`nil`** を返す。IPv4 も **16バイトの形**で返すので、IPv4 か確かめるには **`To4()`** も使う
> - `netip.ParseAddr` は、正しくない文字列なら**エラー**を返す。エラーに「なぜダメか」が書いてある。IPv4 か確かめるには **`Is4`**
> - **`ResolveUDPAddr` はバリデーションの代わりにならない**。名前（`localhost` など）も DNS などで調べて変換するし、**空の文字列でもエラーにならない**
> - `netip.Addr` は **`flag.TextVar`** でそのまま受け取れる。正しくない文字列は `flag.Parse()` の時点でエラーになる
> - 「ネットワークアドレス」が `10.0.0.0/24` のような**範囲**のことなら、別の関数（`net.ParseCIDR`・`netip.ParsePrefix`）を使う

関連ドキュメント：

- [special-address-0000-qa.md](special-address-0000-qa.md)：`0.0.0.0` は宛先に使えないこと、`127.0.0.1`・`255.255.255.255` などの特別なアドレス
- [udp-tcp-listen-qa.md](udp-tcp-listen-qa.md)：`ResolveUDPAddr` は文字列を住所に変換するだけ
- [byte-sequence-and-interpretation-qa.md](byte-sequence-and-interpretation-qa.md)：`net.IP` の中身は `[]byte`（8-4）
- [udp-echo-client-review.md](../reviews/udp-echo-client-review.md)：Client のレビュー記録（4回目：宛先を引数で受け取る）

## 目次

- [質問](#質問)
- [1. 結論](#1-結論)
- [2. なぜ確かめる必要があるのか](#2-なぜ確かめる必要があるのか)
- [3. `net.ParseIP`](#3-netparseip)
- [4. `netip.ParseAddr`](#4-netipparseaddr)
- [5. 2つを比べる](#5-2つを比べる)
- [6. 注意：`ResolveUDPAddr` はバリデーションの代わりにならない](#6-注意resolveudpaddr-はバリデーションの代わりにならない)
- [7. もう一歩進んだ使い方](#7-もう一歩進んだ使い方)
- [8. Claude が Mac で確かめた結果](#8-claude-が-mac-で確かめた結果)
- [9. まとめ](#9-まとめ)

---

## 質問

> **Q1.** ネットワークアドレスを表した文字列をパースできるメソッドはないか？ 例えば、文字列として `127.0.0.1` を受け取ったときに、それが有効なアドレスかどうかをバリデーションしたい。

---

## 1. 結論

| やりたいこと | 使うもの |
|---|---|
| 文字列が正しい IPアドレスか確かめる（古くからある方法） | `net.ParseIP`（だめなら `nil`）。IPv4 に絞るなら `To4()` |
| 文字列が正しい IPアドレスか確かめる（新しい方法、エラー付き） | `netip.ParseAddr`（だめならエラー）。IPv4 に絞るなら `Is4` |
| 引数を受け取る時点で確かめる | `flag.TextVar` に `netip.Addr` を渡す |
| IPアドレスとポートをまとめて確かめる | `netip.ParseAddrPort` |
| `10.0.0.0/24` のような範囲を確かめる | `net.ParseCIDR`、`netip.ParsePrefix` |

---

## 2. なぜ確かめる必要があるのか

### 2-1. 引数は、ただの文字列

`flag.String` で受け取った `-addr` は、**ただの文字列**です。人が打ち間違えるかもしれません。

```text
 正しい：  -addr 10.0.0.1
 間違い：  -addr 10.0.0.256    ← 256 は 1バイト（0〜255）に入らない
           -addr 10.0.01       ← 数が3つしかない
           -addr abc           ← そもそも数ではない
           -addr ""            ← 空
```

### 2-2. 「パース」とは

**パース**（parse）とは、文字列を読んで、**決まった形として解釈する**ことです。IPアドレスなら、`"127.0.0.1"` という文字列を、**4バイトの値** `7f 00 00 01` として読み取ります。

```text
 文字列 "127.0.0.1"
   │
   │ パース：「数.数.数.数」の形か？ それぞれ 0〜255 か？
   ▼
 IPアドレス（4バイト）：7f 00 00 01
```

形が正しくなければ、パースに**失敗**します。つまり、**パースできたかどうかで、正しいアドレスかを確かめられます**。これがバリデーションに使える理由です。

> IPv4 のアドレスは4バイトなので、それぞれの数は **0〜255**（1バイトで表せる範囲）です（[integer-size-and-byte-format-qa.md](integer-size-and-byte-format-qa.md) の5章）。`256` が入っていたら、パースは失敗します。

---

## 3. `net.ParseIP`

### 3-1. 使い方の形

```text
net.ParseIP(文字列) → net.IP
```

| 結果 | 意味 |
|---|---|
| `net.IP` の値 | パースできた |
| **`nil`** | パースできなかった（正しくない文字列） |

**エラーは返しません**。`nil` かどうかで判断します。

### 3-2. IPv4 も 16バイトで返ってくる

`net.ParseIP` は、IPv4 のアドレスでも **16バイト**（IPv6 と同じ大きさ）の形で返します。Mac で `net.ParseIP("127.0.0.1")` の長さを確かめると、**16** でした。

```text
 net.ParseIP("127.0.0.1") の中身（16バイト）

 ┌──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┐
 │00│00│00│00│00│00│00│00│00│00│ff│ff│7f│00│00│01│
 └──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┘
  └──────────── IPv4 を IPv6 の形で表すための部分 ──────┘└─ 127.0.0.1 ─┘
```

そのため、**IPv4 かどうか**を確かめるには、**`To4()`** を使います。

| `To4()` の結果 | 意味 |
|---|---|
| 4バイトの `net.IP` | IPv4 のアドレス |
| `nil` | IPv4 ではない（IPv6 のアドレス、またはもともと `nil`） |

> DHCP のメッセージに IPアドレスを書くときも、`To4()` で4バイトにしてから `copy` します（[go-send-buffer-and-copy-qa.md](go-send-buffer-and-copy-qa.md) の4-2）。

### 3-3. 確かめる流れ

```text
 文字列を受け取る
     │
     ▼
 net.ParseIP で変換する
     │
     ├── nil ─────────────────▶ 正しくない IPアドレス → エラーにして終了
     │
     ▼
 To4() で4バイトにする
     │
     ├── nil ─────────────────▶ IPv4 ではない（IPv6 など） → エラーにして終了
     │
     ▼
 正しい IPv4 のアドレス
```

---

## 4. `netip.ParseAddr`

### 4-1. 使い方の形

`net/netip` は、Go 1.18 から入った**新しめのパッケージ**です。

```text
netip.ParseAddr(文字列) → (netip.Addr, error)
```

| 結果 | 意味 |
|---|---|
| エラーが `nil` | パースできた |
| エラーがある | パースできなかった。**エラーに理由が書いてある** |

### 4-2. エラーに理由が書いてある

Mac で試したときのエラーです。

| 文字列 | エラー |
|---|---|
| `256.1.1.1` | `ParseAddr("256.1.1.1"): IPv4 field has value >255` |
| `abc` | `ParseAddr("abc"): unable to parse IP` |
| `010.0.0.1` | `ParseAddr("010.0.0.1"): IPv4 field has octet with leading zero` |
| `127.0.0.1:9067` | `ParseAddr("127.0.0.1:9067"): unexpected character (at ":9067")` |

このエラーを**そのまま標準エラー出力に出す**だけで、使う人に「何が間違っているか」を伝えられます。

> `010.0.0.1` のように**先頭に0が付いた数**は、昔のソフトでは「8進数」として読まれることがありました（010 ＝ 8）。読み方がソフトによって変わると危ないので、今の Go は受け付けません。

### 4-3. 便利なメソッド

`netip.Addr` には、アドレスの種類を確かめるメソッドがあります（`go doc net/netip.Addr` で一覧を見られます）。

| メソッド | 確かめること | 例 |
|---|---|---|
| `Is4` | IPv4 か | `127.0.0.1` → true、`::1` → false |
| `IsLoopback` | ループバック（自分自身）か | `127.0.0.1` → true |
| `IsUnspecified` | 未指定（`0.0.0.0`）か | `0.0.0.0` → true |

宛先に `0.0.0.0` を指定されたら断る（[special-address-0000-qa.md](special-address-0000-qa.md)）、といったチェックにも使えます。

---

## 5. 2つを比べる

| | `net.ParseIP` | `netip.ParseAddr` |
|---|---|---|
| パッケージ | `net` | `net/netip`（Go 1.18〜） |
| 失敗したとき | **`nil`** を返す | **エラー**を返す（理由付き） |
| IPv4 の持ち方 | 16バイトの `[]byte`（IPv6 の形に入れる） | IPv4 と IPv6 を**区別して**持つ（スライスではない値） |
| IPv4 か確かめる | `To4()` が `nil` でないか | `Is4` |
| 比べる | `Equal` メソッドを使う | `==` で比べられる |
| `ResolveUDPAddr`・`WriteToUDP` などと組み合わせる | そのまま使える（`net.UDPAddr` の中身が `net.IP`） | `netip.AddrPort` を使うメソッド（`WriteToUDPAddrPort` など）がある |

**どちらを使ってもかまいません**。エラーの理由を表示したいなら `netip.ParseAddr`、今のコード（`ResolveUDPAddr` → `WriteToUDP`）に合わせたいなら `net.ParseIP` が書きやすいです。

> udp-echo の Server は `ReadFromUDPAddrPort`・`WriteToUDPAddrPort` を使っているので、`netip` の型（`netip.AddrPort`）とつながっています。

---

## 6. 注意：`ResolveUDPAddr` はバリデーションの代わりにならない

### 6-1. `ResolveUDPAddr` は「名前でも受け付けて、住所に変換する」関数

今のコードでは、`ResolveUDPAddr` に `-addr` と `-port` をつないだ文字列を渡しています。これがエラーを返すこともありますが、**正しい IPアドレスかを確かめる関数ではありません**。

| `-addr` の文字列 | `ResolveUDPAddr` の結果 | 何が起きているか |
|---|---|---|
| `127.0.0.1` | ✅ | 正しい |
| `""`（空） | ✅ **エラーにならない** | `:9067` になり、IPアドレスのない住所ができる。送ると `no route to host` |
| `localhost` | ✅ エラーにならない | **名前**として、`127.0.0.1` に変換された |
| `010.0.0.1` | ✅ エラーにならない | `netip.ParseAddr` では断られる文字列も、通ってしまう |
| `256.1.1.1` | ❌ `lookup 256.1.1.1: no such host` | IPアドレスとして間違っているのに、**名前として調べに行った**結果のエラー |
| `abc` | ❌ `lookup abc: no such host` | 同じく、名前として調べに行った |

### 6-2. 名前を調べに行く、とは

`ResolveUDPAddr` の「Resolve」は「**名前解決**」のことです。IPアドレスとして読めない文字列は、**ホスト名**（`localhost`、`example.com` など）だと考えて、`/etc/hosts` や DNS で調べに行きます。

```text
 "256.1.1.1" を渡すと……

   IPアドレスとして読める？ ── いいえ
       │
       ▼
   ホスト名として調べる（/etc/hosts、DNS）
       │
       ▼
   見つからない → "no such host"
```

エラーの理由が「アドレスの形が間違っている」ではなく「**そういう名前のマシンがない**」になるので、使う人には分かりにくくなります。また、DNS に問い合わせる分、**時間がかかる**こともあります。

### 6-3. 先に確かめる

IPアドレスだけを受け付けたいなら、**`ResolveUDPAddr` の前に**、`net.ParseIP` か `netip.ParseAddr` で確かめます。

```text
 -addr を受け取る
     │
     ▼
 ParseIP / ParseAddr で確かめる ── だめ ──▶ 理由を表示して終了
     │
     ▼
 ResolveUDPAddr（または、確かめた値から住所を作る）
     │
     ▼
 送る
```

> 逆に、「`localhost` のような**名前も受け付けたい**」なら、`ResolveUDPAddr` に任せるのが正しい使い方です。何を受け付けたいかで、選ぶ関数が変わります。

---

## 7. もう一歩進んだ使い方

### 7-1. `flag.TextVar` で、受け取るときに確かめる

`netip.Addr` は「**文字列から自分を作る**」メソッド（`UnmarshalText`）を持っています。`flag` パッケージの **`TextVar`** は、このメソッドを持つ型なら何でも受け取れます。

```text
 flag.TextVar に netip.Addr を渡しておくと……

   flag.Parse()
     │
     ├── -addr 10.0.0.1  → netip.Addr として受け取れる
     │
     └── -addr 256.1.1.1 → flag.Parse() の時点でエラー
           invalid value "256.1.1.1" for flag -addr:
           ParseAddr("256.1.1.1"): IPv4 field has value >255
```

文字列で受け取ってから自分でパースする手間が省けます。デフォルト値も `netip.Addr` で渡します。

### 7-2. `netip.ParseAddrPort`：IPアドレスとポートをまとめて

`"127.0.0.1:9067"` のように、**IPアドレスとポートが1つの文字列**になっているときは、`netip.ParseAddrPort` で両方まとめて確かめられます。

| 文字列 | 結果 |
|---|---|
| `127.0.0.1:9067` | ✅ アドレス `127.0.0.1`、ポート `9067` |
| `127.0.0.1:99999` | ❌ `invalid port "99999"`（ポートは 0〜65535。2バイトの範囲） |

今のコードは `-addr` と `-port` を分けているので、使うなら「つないでから確かめる」形になります。

### 7-3. 範囲（ネットワークアドレス）を扱うとき

「ネットワークアドレス」という言葉は、**1台のアドレス**ではなく、`10.0.0.0/24` のような**アドレスの範囲**を指すこともあります。

| 書き方 | 意味 | 使う関数 |
|---|---|---|
| `10.0.0.1` | 1台のアドレス | `net.ParseIP`、`netip.ParseAddr` |
| `10.0.0.0/24` | `10.0.0.0`〜`10.0.0.255` の範囲（CIDR 表記） | `net.ParseCIDR`、`netip.ParsePrefix` |

`netip.ParsePrefix` でパースした範囲には、「このアドレスが範囲に入っているか」を確かめるメソッド（`Contains`）があります。DHCP サーバーで「**貸し出す範囲**」を扱うとき（[learning-exercise.md](../../learning-exercise.md) の15章）に役立ちます。

udp-echo の宛先は1台のアドレスなので、`ParseIP` か `ParseAddr` の方です。

---

## 8. Claude が Mac で確かめた結果

Mac（Go 1.26.4）で確かめました。

**3つの関数に同じ文字列を渡す**（`ResolveUDPAddr` には `":9067"` をつないで渡した）

| 文字列 | `net.ParseIP` | `To4()` | `netip.ParseAddr` | `ResolveUDPAddr` |
|---|---|---|---|---|
| `127.0.0.1` | `127.0.0.1` | 4バイト | ✅ | ✅ |
| `10.0.0.1` | `10.0.0.1` | 4バイト | ✅ | ✅ |
| `256.1.1.1` | `nil` | `nil` | ❌ `IPv4 field has value >255` | ❌ `no such host` |
| `abc` | `nil` | `nil` | ❌ `unable to parse IP` | ❌ `no such host` |
| `localhost` | `nil` | `nil` | ❌ `unable to parse IP` | ✅ |
| `::1` | `::1` | `nil`（IPv4 ではない） | ✅（`Is4` は false） | ❌ `too many colons in address` |
| `127.0.0.1:9067` | `nil` | `nil` | ❌ `unexpected character` | ❌ `too many colons in address` |
| `010.0.0.1` | `nil` | `nil` | ❌ `octet with leading zero` | ✅ |
| `""`（空） | `nil` | `nil` | ❌ `unable to parse IP` | ✅ |

**そのほか**

| 確かめたこと | 結果 |
|---|---|
| `len(net.ParseIP("127.0.0.1"))` | **16** |
| `flag.TextVar` に `netip.Addr` を渡し、`-addr 10.0.0.1` | 受け取れた |
| 同じく `-addr 256.1.1.1` | `flag.Parse()` がエラー：`invalid value "256.1.1.1" for flag -addr: ParseAddr("256.1.1.1"): IPv4 field has value >255` |
| `netip.ParseAddrPort("127.0.0.1:9067")` | アドレス `127.0.0.1`、ポート `9067` |
| `netip.ParseAddrPort("127.0.0.1:99999")` | `invalid port "99999"` |
| `netip.ParsePrefix("10.0.0.0/24")` が `10.0.0.1` を含むか | true |
| `IsLoopback`（`127.0.0.1`）、`IsUnspecified`（`0.0.0.0`） | どちらも true |

---

## 9. まとめ

- IPアドレスの文字列は、**パースできたかどうか**で正しいか確かめられる
- **`net.ParseIP`**：だめなら `nil`。IPv4 も16バイトで返るので、IPv4 に絞るなら **`To4()`**
- **`netip.ParseAddr`**：だめなら**理由付きのエラー**。IPv4 に絞るなら **`Is4`**。`IsLoopback`・`IsUnspecified` なども使える
- **`ResolveUDPAddr` は名前解決もする**ので、バリデーションの代わりにならない。空の文字列も通ってしまう。IPアドレスだけを受け付けたいなら、**先に** `ParseIP` か `ParseAddr` で確かめる
- `netip.Addr` は **`flag.TextVar`** で受け取れて、`flag.Parse()` の時点で確かめられる
- 範囲（`10.0.0.0/24`）なら `net.ParseCIDR`・`netip.ParsePrefix`。DHCP サーバーの貸し出す範囲で使う
