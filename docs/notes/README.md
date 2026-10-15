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
└── troubleshooting/     ← トラブルの記録：実際に起きた問題と、原因・対処
```

| フォルダ | 置くもの | 迷ったときの目安 |
|---|---|---|
| `runbooks/` | 手順書 | 「これを見ながら実行する」ドキュメント |
| `qa/` | Q&A | 「○○とは何か？」「○○と××の違いは？」という質問への回答 |
| `results/` | 実施記録 | 「やってみたら、こうなった」という結果の記録 |
| `troubleshooting/` | トラブルの記録 | 「エラーが出た。原因はこれで、こう直した」という記録 |

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
| [socket-bind-qa.md](qa/socket-bind-qa.md) | **bindとは、ソケットに自分側の住所（IPアドレスとポート番号）を付けること**。bindしないとOSがポートを選ぶこと、socatの `sourceport` との違い、DHCPクライアントが68番にbindする理由 | bind、ソケット、送信元ポート、エフェメラルポート、`0.0.0.0` |
| [unicast-broadcast-multicast.md](qa/unicast-broadcast-multicast.md) | ユニキャスト・ブロードキャスト・マルチキャストの違いと、DHCPでの使われ方 | ブロードキャスト、マルチキャスト、`255.255.255.255`、DHCPv6 |
| [ssh-via-vmnet2-qa.md](qa/ssh-via-vmnet2-qa.md) | Router VMを踏み台にしたSSHがどこを通っているのか、`$SSH_CONNECTION` の読み方 | 踏み台、VMnet2、ルーティングテーブル、`$SSH_CONNECTION` |

**Linux・コマンド**

| ドキュメント | 内容 | 主なキーワード |
|---|---|---|
| [scp-qa.md](qa/scp-qa.md) | `scp` とは何か、`cp` との違い、コピーの向き3パターン | `scp`、SSH、`-3`、`-P` |
| [tilde-qa.md](qa/tilde-qa.md) | `~`（チルダ）とは何か、scpでどのマシンのホームを指すか、どのOSで使えるか | `~`、ホームディレクトリ、チルダ展開、シェル |
| [file-vs-folder-permission-qa.md](qa/file-vs-folder-permission-qa.md) | **ファイルとフォルダで `rwx` の意味が違う**こと。保存に実行権限は不要なこと、「読む」「実行する」「たどり着く」の違い | ファイルの `x`（実行）、フォルダの `x`（中に入る）、nano、`chmod +x` |

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

---

## ドキュメント同士の関係

STEP 1を進める中で、どのドキュメントからどのドキュメントが生まれたかを示します。

```text
learning-exercise.md（STEP 1：UDP通信を理解する）
│
├── qa/step1-udp-socket-qa.md ………………………… UDPソケットとポート番号
│     └── qa/unicast-broadcast-multicast.md ……… 送り先の種類
│
└── runbooks/step1-udp-tools-runbook.md ……………… STEP 1の手順書
      │
      ├── results/step1-results.md ………………………… 手順書を実行した結果の記録
      │     └── qa/socket-bind-qa.md ……………………………… bindとは何か（sourceport との違い）
      ├── runbooks/tcpdump-save-capture.md ………… キャプチャをファイルに保存する
      │
      └── troubleshooting/client-vm-offline-package-install.md … 0-1でaptが失敗
            ├── qa/ssh-via-vmnet2-qa.md ……………………… SSHの経路
            ├── qa/scp-qa.md ………………………………………… ファイルのコピー
            │     └── qa/tilde-qa.md ………………………………… パスの中の「~」
            └── troubleshooting/apt-permission-denied-qa.md … 5-3でaptが失敗、Linuxの権限
                  └── qa/file-vs-folder-permission-qa.md … ファイルとフォルダで rwx の意味が違う
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

### 3. リンクを書く

ほかのドキュメントへのリンクは、**自分のファイルがあるフォルダから見た相対パス**で書きます。

| リンクする場所 | 書き方の例（`qa/` の中のファイルから） |
|---|---|
| 同じフォルダのドキュメント | `[scp-qa.md](scp-qa.md)` |
| 別のフォルダのドキュメント | `[apt-permission-denied-qa.md](../troubleshooting/apt-permission-denied-qa.md)` |
| `docs/` 直下のドキュメント | `[vm-environment.md](../../vm-environment.md)` |

### 4. このREADMEを更新する

「ドキュメント一覧」と「ドキュメント同士の関係」に追記してください。
