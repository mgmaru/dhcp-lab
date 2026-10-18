# docs/notes：学習メモ一覧

演習（[learning-exercise.md](../learning-exercise.md)）を進める中で書いた手順書・Q&A・実施記録・トラブルの記録をまとめたフォルダです。

## まず読むもの

| やりたいこと | 読むドキュメント |
|---|---|
| STEP 1を実際に手を動かして進めたい | [runbooks/step1-udp-tools-runbook.md](runbooks/step1-udp-tools-runbook.md) |
| Client VMにパッケージを入れたい | [troubleshooting/client-vm-offline-package-install.md](troubleshooting/client-vm-offline-package-install.md) |

---

## フォルダ構成

ドキュメントの**種類**ごとにフォルダを分けています。

```text
docs/notes/
├── README.md            ← このファイル（一覧）
├── runbooks/            ← 手順書：上から順に実行できる形で書いたもの
├── qa/                  ← Q&A：質問と、その回答
├── results/             ← 実施記録：手順書を実行した結果
├── troubleshooting/     ← トラブルの記録：実際に起きた問題と、原因・対処
├── cheatsheets/         ← 早見表：コマンドやオプションをすぐに引けるようにまとめたもの
└── reviews/             ← レビュー記録：自分で書いたコードのレビューと、修正の経過
```

| フォルダ | 置くもの | 迷ったときの目安 |
|---|---|---|
| `runbooks/` | 手順書 | 「これを見ながら実行する」ドキュメント |
| `qa/` | Q&A | 「○○とは何か？」「○○と××の違いは？」という質問への回答 |
| `results/` | 実施記録 | 「やってみたら、こうなった」という結果の記録 |
| `troubleshooting/` | トラブルの記録 | 「エラーが出た。原因はこれで、こう直した」という記録 |
| `cheatsheets/` | 早見表 | 「このオプション何だっけ？」とすぐに引くための一覧 |
| `reviews/` | レビュー記録 | 「何ができていなくて、どこまで直したか」というコードのレビューの記録 |

---

## ドキュメント一覧

### runbooks/：手順書

上から順に実行できる形で書いたドキュメントです。

| ドキュメント | 内容 | 関係するSTEP |
|---|---|---|
| [step1-udp-tools-runbook.md](runbooks/step1-udp-tools-runbook.md) | `nc`・`socat`・`tcpdump` で、Client VMとRouter VMの間でUDPを送り合い、パケットを観察する | STEP 1 |
| [tcpdump-save-capture.md](runbooks/tcpdump-save-capture.md) | tcpdumpのキャプチャ結果をpcapファイルやテキストファイルに保存し、読み直す・Wiresharkで開く | 全STEP |

### qa/：Q&A

質問と、その回答をまとめたドキュメントです。テーマごとに分けて載せています。

**ネットワーク**

| ドキュメント | 内容 | 主なキーワード |
|---|---|---|
| [step1-udp-socket-qa.md](qa/step1-udp-socket-qa.md) | UDPソケットとは何か、TCPとの違い、DHCPのクライアント側が68番に決まっている理由 | UDP、TCP、ソケット、ポート番号、エフェメラルポート |
| [socket-protocol-layers-qa.md](qa/socket-protocol-layers-qa.md) | **ネットワークの層**（アプリケーション・トランスポート・ネットワーク・リンク）と、**ソケットはその境目の窓口**であること。ソケットに指定するのはUDP/TCPまでで、DHCPやHTTPは指定しない理由 | 層、カプセル化、ソケット、`udp4`、HTTP、DHCP |
| [socket-bind-qa.md](qa/socket-bind-qa.md) | **bindとは、ソケットに自分側の住所（IPアドレスとポート番号）を付けること**。bindしないとOSがポートを選ぶこと、socatの `sourceport` との違い、DHCPクライアントが68番にbindする理由 | bind、ソケット、送信元ポート、エフェメラルポート、`0.0.0.0` |
| [udp-tcp-listen-qa.md](qa/udp-tcp-listen-qa.md) | Go の **`ResolveUDPAddr`（文字列を住所に変換するだけ）と `ListenUDP`（socket と bind をまとめてやる）**の違い。`ListenUDP` の「Listen」が TCP の listen と違うこと。**TCP の listen は接続の申し込みを受け付け始める段階**で、accept が相手ごとのソケットを返すこと。UDP と TCP の違いを図で比較 | `ResolveUDPAddr`、`ListenUDP`、`DialUDP`、listen、accept、backlog、3ウェイハンドシェイク、`UNCONN`、`LISTEN` |
| [udp-message-size-qa.md](qa/udp-message-size-qa.md) | **UDP のメッセージの大きさ**。今の Server は大きいメッセージを**黙って切って使う**こと、UDP の上限（**IPv4 で65507バイト**）以上の入れ物なら切れないこと、DHCP の576バイトは**クライアントが受け取る側**のルールであること。UDP には**シーケンス番号がなく1通で完結**すること、IP フラグメンテーション、大きいデータは **TCP を使う**か**区切る決まりを自分で作る**こと（TFTP・DNS・QUIC の例）。**形を確かめる**とは、OS が外した UDP ヘッダーではなく、**DHCP のメッセージの形**（240バイトの決まった形、magic cookie、オプションの3つ組）を確かめること。TCP の**並べ直しは OS がやる**こと | UDP、65507、576、Maximum DHCP Message Size、シーケンス番号、IP フラグメンテーション、1472、TCP、TFTP、DNS、QUIC、xid、DHCP のメッセージの形、magic cookie、オプション、op、受信バッファ、並べ直し、ACK |
| [tcp-stream-and-http-framing-qa.md](qa/tcp-stream-and-http-framing-qa.md) | **TCP には「データの区切り」がない**こと（送った回数と受け取った回数は一致しない。くっつく・分かれる）。区切りを作る方法（**長さを先に送る**、**区切りの文字**）。**TCP は「正しく全部届ける」、HTTP は「区切りと意味」**を担当する層の役割分担。HTTP/1.1 の Content-Length・チャンク形式、区切りの読み違い（リクエストスマグリング）、HTTP/2 のフレーム。UDP と TCP は守るものが反対 | TCP、バイトストリーム、区切り、フレーミング、Content-Length、chunked、HTTP/2、フレーム、ストリーム ID、`io.ReadFull`、`bufio`、層の役割分担 |
| [unicast-broadcast-multicast.md](qa/unicast-broadcast-multicast.md) | ユニキャスト・ブロードキャスト・マルチキャストの違いと、DHCPでの使われ方 | ブロードキャスト、マルチキャスト、`255.255.255.255`、DHCPv6 |
| [ssh-via-vmnet2-qa.md](qa/ssh-via-vmnet2-qa.md) | Router VMを踏み台にしたSSHがどこを通っているのか、`$SSH_CONNECTION` の読み方 | 踏み台、VMnet2、ルーティングテーブル、`$SSH_CONNECTION` |
| [tcpdump-n-option-qa.md](qa/tcpdump-n-option-qa.md) | tcpdumpの `-n` オプションの意味。付けないとIPアドレスやポート番号が名前（`68` → `bootpc` など）で表示されること、数字で見るのがおすすめな理由 | `-n`、`/etc/services`、`/etc/hosts`、`bootps`、`bootpc` |

**Linux・コマンド**

| ドキュメント | 内容 | 主なキーワード |
|---|---|---|
| [scp-qa.md](qa/scp-qa.md) | `scp` とは何か、`cp` との違い、コピーの向き3パターン | `scp`、SSH、`-3`、`-P` |
| [tilde-qa.md](qa/tilde-qa.md) | `~`（チルダ）とは何か、scpでどのマシンのホームを指すか、どのOSで使えるか | `~`、ホームディレクトリ、チルダ展開、シェル |
| [file-vs-folder-permission-qa.md](qa/file-vs-folder-permission-qa.md) | **ファイルとフォルダで `rwx` の意味が違う**こと。保存に実行権限は不要なこと、「読む」「実行する」「たどり着く」の違い | ファイルの `x`（実行）、フォルダの `x`（中に入る）、nano、`chmod +x` |
| [stdout-stderr-qa.md](qa/stdout-stderr-qa.md) | **標準出力（普通の結果）と標準エラー出力（エラーやお知らせ）**の違い、**どちらに出すかを決めるのはプログラム**であること、区別しないと何が起こるか、try-catch との関係、プログラムの中と外（外への出口の種類）、終了コードを返す理由と数字の分け方、`os.Stderr` と `fmt.Fprintln` の役割分担、`Fprintln` と `Println` の違い、リダイレクト（`>`・`2>`・`2>&1`・`\|`）、Goでの出力先、エラーの表示と終了コード | stdout、stderr、ファイルディスクリプタ、リダイレクト、パイプ、`2>&1`、`>&2`、終了コード、`echo $?`、`fmt`、`log`、try-catch、throw、ソケット、親プロセス、`&&`・`\|\|`、`grep` の終了コード、127・130、`os.Stderr`、`fmt.Fprintln`、`fmt.Println`、`Fprintf`、`io.Writer` |

**Go**

| ドキュメント | 内容 | 主なキーワード |
|---|---|---|
| [go-module-qa.md](qa/go-module-qa.md) | パッケージ・モジュール・go.mod の関係と、トップディレクトリに go.mod を1つ置く理由。この演習で何がパッケージになるか、`package main`（入口）とそれ以外（部品）の違い、**1つのフォルダ ＝ 1つのパッケージ**、別のパッケージ・モジュールのコードの使い方（インポートはパッケージ単位、大文字の名前だけ公開）、`main` 関数は1フォルダに1つ、別のモジュールをつなぐ方法（`require`・`replace`・go.work）、OSSを読むときのコツ | パッケージ、モジュール、go.mod、go.work、インポートパス、`package main`、`main` 関数、`init`、エクスポート、`require`、`replace`、`go get`、`cmd/`、`internal/` |
| [go-module-download-qa.md](qa/go-module-download-qa.md) | `require` したモジュールはどこから来て、どこに保存されるのか。npm（`node_modules`）との違い、`go get` の役割（go.mod への書き込みとダウンロード）、ビルドキャッシュの場所の決まり方 | モジュールプロキシ、モジュールキャッシュ、`~/go/pkg/mod`、go.sum、`go get`、`go mod tidy`、vendor、npm、ビルドキャッシュ、`GOCACHE` |
| [go-slice-buffer-reuse-qa.md](qa/go-slice-buffer-reuse-qa.md) | **受け取り用の入れ物を使い回しても、`buf[:n]` なら前のデータが混ざらない理由**。受け取りのメソッドは先頭から書き込み、n より後ろには何もしないこと。スライスの長さと容量、入れ物より長いメッセージは黙って切れること、**`buf[:n]` はコピーではなく指しているだけ**で、取っておくなら `string(…)` などでコピーすること | スライス、`[]byte`、`make`、len、cap、`buf[:n]`、バッファの使い回し、切り捨て、コピー、`string()` |
| [go-buffer-size-and-append-qa.md](qa/go-buffer-size-and-append-qa.md) | **受け取り用の入れ物の大きさ**の決め方（IPv4 の UDP の最大 65507バイト、DHCP は 576バイト以上）。入れ物より大きいメッセージは黙って切れること、**「最大＋1バイト」で切れたことに気づく**方法。**`make` は容量の分を確保する**のでメモリは減らないこと。**受け取りには長さ、`append` には容量**を指定する理由。**スライスは「箱」と「札」**でできていて、`append` は空きがなければ大きい箱に**引っ越す**こと。**容量は確保してある箱の数、長さは使ってよい箱の数**で、受け取りのメソッドが長さの範囲にしか書かない理由（`io.Reader` の約束、札のコピー） | スライス、len、cap、`make`、`append`、バッファの大きさ、切り捨て、`MSG_TRUNC`、65507、メモリの確保、箱と札、引っ越し、連結リスト |
| [go-slice-memory-basics-qa.md](qa/go-slice-memory-basics-qa.md) | **メモリとスライスの基礎**。メモリの最小の単位は**1バイト（1マス）**で、スライスの箱（要素1個）の大きさは型によって違うこと（`int` は64ビットのマシンで8バイト）。長さと容量は箱の数で数えること。最初の容量は32バイト分、`append` の容量は**直前の容量をもとに、小さいうちは2倍、大きくなると1.25倍に近づく**こと。余る空きの数は計算で決まること。溢れたら箱を足すのではなく**大きい箱の列に引っ越す**こと | メモリ、番地、バイト、`int`、`int32`、64ビット、`append`、容量の増え方、メモリの区切り、ゴミ集め |
| [go-send-buffer-and-copy-qa.md](qa/go-send-buffer-and-copy-qa.md) | **送る側にも入れ物は要るのか**。受け取りは OS が**書く**場所が要るが、送信は OS が**読む**データがあればよいこと、入れ物を全部送ると余計な0まで送られること。DHCP のように**決まった位置にフィールドを詰める**方法（直接代入・`binary.BigEndian.PutUint32`・`copy`）、構造体＋変換メソッドがおすすめな理由、練習用のミニフォーマット。**`copy(dst, src)` のしくみ**（先頭から上書き、短い方に合わせる、溢れた分は捨てる）、**書く位置は `dst` の切り出し方で決める**こと、切り出しても**新しい箱はできない**こと | `copy`、dst、src、`[]byte`、送信、`WriteToUDP`、オフセット、ビッグエンディアン、`binary.BigEndian`、`PutUint32`、DHCP のメッセージの形、構造体、`binary.Write`、`BinaryMarshaler`、切り出し、箱と札、GC、エスケープ解析 |

**データの表し方（情報工学の基礎）**

| ドキュメント | 内容 | 主なキーワード |
|---|---|---|
| [integer-size-and-byte-format-qa.md](qa/integer-size-and-byte-format-qa.md) | **ビットとバイト**（1バイト＝8ビット＝256通り、だから **0〜255**）、16進数でバイトを読む方法。Go の **`int` は8バイト**（64ビットのマシン）で、`uint32` は4バイト。固定長の欄は**値が小さくても決まったバイト数を全部使う**こと。**Go の型とフォーマットの大きさは別のもの**で、違うときは変換すること。大きい型から小さい型への変換は**圧縮ではなく切り捨て**で、範囲外の値は**エラーなしで壊れる**（300 → 44、−1 → 255）ので、変換の前に範囲を確かめること。ビッグエンディアン。ミニフォーマットの `buf` の長さと、コマンド引数（`flag`）での受け取り方 | ビット、バイト、2ⁿ、0〜255、16進数、`int`、`byte`、`uint8`、`uint32`、符号あり・なし、2の補数、型の変換、切り捨て、オーバーフロー、範囲チェック、ビッグエンディアン、`flag`、固定長、可変長 |

> Linuxの権限の基本（`rwx`、`755`、`644`、フォルダ構成）は、トラブルの記録 [apt-permission-denied-qa.md](troubleshooting/apt-permission-denied-qa.md) の6章で詳しく説明しています。

### results/：実施記録

手順書を実行した結果の記録です。

| ドキュメント | 内容 | 関係するSTEP |
|---|---|---|
| [step1-results.md](results/step1-results.md) | STEP 1の手順書を実行した結果（キャプチャ）と、その読み方 | STEP 1 |

### troubleshooting/：トラブルの記録

実際に起きた問題と、その原因・対処の記録です。

| ドキュメント | 起きたこと | 原因 |
|---|---|---|
| [client-vm-offline-package-install.md](troubleshooting/client-vm-offline-package-install.md) | Client VMで `apt install` が `Temporary failure resolving 'deb.debian.org'` で失敗した | Client VMは隔離LAN（VMnet2）にしかつながっておらず、インターネットに出られない |
| [apt-permission-denied-qa.md](troubleshooting/apt-permission-denied-qa.md) | `.deb` ファイルからの `apt install` が `Permission denied` で失敗した。**Linuxの権限のしくみの解説も含む** | aptがファイルを読むときに使う `_apt` ユーザーが、ホームディレクトリに入れない |

### cheatsheets/：早見表

コマンドやオプションを、すぐに引けるようにまとめたものです。

| ドキュメント | 内容 |
|---|---|
| [tcpdump-cheatsheet.md](cheatsheets/tcpdump-cheatsheet.md) | tcpdumpのよく使うオプション・フィルタ・目的別のコマンド例・出力の読み方・よくあるつまずき |

### reviews/：レビュー記録

自分で書いたコードのレビューと、修正の経過の記録です。

| ドキュメント | 対象のコード | 現在の状況 |
|---|---|---|
| [udp-echo-server-review.md](reviews/udp-echo-server-review.md) | `playground/udp-echo/server/` | M0〜M3 完了。Mac の上で S1〜S4 が揃った。次は VM で動かす（13回目） |
| [udp-echo-client-review.md](reviews/udp-echo-client-review.md) | `playground/udp-echo/client/` | M0（エラー処理）完了。M1（bind）は自分側のIPアドレスの指定を検討中。M2（送って返事を受け取る）はまだ |

---

## ドキュメント同士の関係

STEP 1を進める中で、どのドキュメントからどのドキュメントが生まれたかを示します。

```text
learning-exercise.md（STEP 1：UDP通信を理解する）
│
├── qa/step1-udp-socket-qa.md ………………………… UDPソケットとポート番号
│     ├── qa/socket-protocol-layers-qa.md ………… ソケットに指定するプロトコルと、ネットワークの層
│     ├── qa/unicast-broadcast-multicast.md ……… 送り先の種類
│     ├── qa/udp-tcp-listen-qa.md ………………………… ResolveUDPAddr・ListenUDP の中身、TCP の listen、UDP と TCP の違い
│     ├── qa/udp-message-size-qa.md ……………………… UDP の1通の大きさ、DHCP サーバーの入れ物、大きいデータの送り方
│     └── qa/tcp-stream-and-http-framing-qa.md …… TCP には区切りがない、HTTP が区切りを担当する
│
├── runbooks/step1-udp-tools-runbook.md ……………… STEP 1の手順書
│     │
│     ├── results/step1-results.md ………………………… 手順書を実行した結果の記録
│     │     └── qa/socket-bind-qa.md ……………………………… bindとは何か（sourceport との違い）
│     ├── cheatsheets/tcpdump-cheatsheet.md …………… tcpdumpの早見表
│     │     └── qa/tcpdump-n-option-qa.md ……………… -n オプションの意味
│     ├── runbooks/tcpdump-save-capture.md ………… キャプチャをファイルに保存する
│     │
│     └── troubleshooting/client-vm-offline-package-install.md … 0-1でaptが失敗
│           ├── qa/ssh-via-vmnet2-qa.md ……………………… SSHの経路
│           ├── qa/scp-qa.md ………………………………………… ファイルのコピー
│           │     └── qa/tilde-qa.md ………………………………… パスの中の「~」
│           └── troubleshooting/apt-permission-denied-qa.md … 5-3でaptが失敗、Linuxの権限
│                 └── qa/file-vs-folder-permission-qa.md … ファイルとフォルダで rwx の意味が違う
│
└── playground/README.md ……………………………………… udp-echo をGoで実装する方針
      ├── qa/go-module-qa.md ……………………………… パッケージ・モジュール・go.mod
      │     └── qa/go-module-download-qa.md ……… モジュールはどこから来て、どこに保存されるのか
      ├── qa/stdout-stderr-qa.md ……………………… 標準出力・標準エラー出力、終了コード（なぜ返すのか・数字の分け方）、try-catch、プログラムの中と外
      ├── qa/go-slice-buffer-reuse-qa.md ………… 受け取り用の入れ物の使い回しと buf[:n]、スライスはコピーではない
      │     └── qa/go-buffer-size-and-append-qa.md … 入れ物の大きさ、長さと容量の使い分け（append）、make と append がメモリでしていること（箱と札）、長さと容量とは、受け取りに長さが必要な理由
      │           └── qa/go-slice-memory-basics-qa.md … メモリの1マスとスライスの箱、型の大きさ、append で容量が増えるしくみ
      ├── reviews/udp-echo-server-review.md …… udp-echo の Server のレビュー記録
      └── reviews/udp-echo-client-review.md …… udp-echo の Client のレビュー記録
            └── qa/go-send-buffer-and-copy-qa.md … 送る側の入れ物、フィールドの詰め方、copy のしくみ（dst の切り出しと箱と札）
                  └── qa/integer-size-and-byte-format-qa.md … ビットとバイト、int の大きさ、型とフォーマットの大きさ、変換は切り捨て
```

---

## ドキュメントを追加するとき

### 1. 置くフォルダを決める

[フォルダ構成](#フォルダ構成)の「迷ったときの目安」を見て、種類に合うフォルダに置きます。

### 2. ファイル名を決める

| 種類 | フォルダ | ファイル名の付け方 | 例 |
|---|---|---|---|
| 特定のSTEPの手順書 | `runbooks/` | `stepN-<内容>-runbook.md` | `step1-udp-tools-runbook.md` |
| STEPに依存しない手順書 | `runbooks/` | `<ツール名>-<内容>.md` | `tcpdump-save-capture.md` |
| 特定のSTEPのQ&A | `qa/` | `stepN-<内容>-qa.md` | `step1-udp-socket-qa.md` |
| STEPに依存しないQ&A | `qa/` | `<内容>-qa.md` | `scp-qa.md`、`tilde-qa.md` |
| 実施記録 | `results/` | `stepN-results.md` | `step1-results.md` |
| トラブルの記録 | `troubleshooting/` | `<起きたこと>.md` | `client-vm-offline-package-install.md` |
| 早見表 | `cheatsheets/` | `<ツール名>-cheatsheet.md` | `tcpdump-cheatsheet.md` |
| レビュー記録 | `reviews/` | `<対象>-review.md` | `udp-echo-server-review.md` |

### 3. リンクを書く

ほかのドキュメントへのリンクは、**自分のファイルがあるフォルダから見た相対パス**で書きます。

| リンクする場所 | 書き方の例（`qa/` の中のファイルから） |
|---|---|
| 同じフォルダのドキュメント | `[scp-qa.md](scp-qa.md)` |
| 別のフォルダのドキュメント | `[apt-permission-denied-qa.md](../troubleshooting/apt-permission-denied-qa.md)` |
| `docs/` 直下のドキュメント | `[vm-environment.md](../../vm-environment.md)` |

### 4. このREADMEを更新する

「ドキュメント一覧」と「ドキュメント同士の関係」に追記してください。
