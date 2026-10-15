# docs/notes：学習メモ一覧

演習（[learning-exercise.md](../learning-exercise.md)）を進める中で書いた手順書・Q&A・トラブルの記録をまとめたフォルダです。

## まず読むもの

| やりたいこと | 読むドキュメント |
|---|---|
| STEP 1を実際に手を動かして進めたい | [step1-udp-tools-runbook.md](step1-udp-tools-runbook.md) |
| Client VMにパッケージを入れたい | [client-vm-offline-package-install.md](client-vm-offline-package-install.md) |

---

## ドキュメント一覧

### 手順書

上から順に実行できる形で書いたドキュメントです。

| ドキュメント | 内容 | 関係するSTEP |
|---|---|---|
| [step1-udp-tools-runbook.md](step1-udp-tools-runbook.md) | `nc`・`socat`・`tcpdump` で、Client VMとRouter VMの間でUDPを送り合い、パケットを観察する | STEP 1 |
| [tcpdump-save-capture.md](tcpdump-save-capture.md) | tcpdumpのキャプチャ結果をpcapファイルやテキストファイルに保存し、読み直す・Wiresharkで開く | 全STEP |

### ネットワークの基礎

| ドキュメント | 内容 | 主なキーワード |
|---|---|---|
| [step1-udp-socket-qa.md](step1-udp-socket-qa.md) | UDPソケットとは何か、TCPとの違い、DHCPのクライアント側が68番に決まっている理由 | UDP、TCP、ソケット、ポート番号、エフェメラルポート |
| [unicast-broadcast-multicast.md](unicast-broadcast-multicast.md) | ユニキャスト・ブロードキャスト・マルチキャストの違いと、DHCPでの使われ方 | ブロードキャスト、マルチキャスト、`255.255.255.255`、DHCPv6 |
| [ssh-via-vmnet2-qa.md](ssh-via-vmnet2-qa.md) | Router VMを踏み台にしたSSHがどこを通っているのか、`$SSH_CONNECTION` の読み方 | 踏み台、VMnet2、ルーティングテーブル、`$SSH_CONNECTION` |

### トラブルの記録

実際に起きた問題と、その原因・対処の記録です。

| ドキュメント | 起きたこと | 原因 |
|---|---|---|
| [client-vm-offline-package-install.md](client-vm-offline-package-install.md) | Client VMで `apt install` が `Temporary failure resolving 'deb.debian.org'` で失敗した | Client VMは隔離LAN（VMnet2）にしかつながっておらず、インターネットに出られない |
| [apt-permission-denied-qa.md](apt-permission-denied-qa.md) | `.deb` ファイルからの `apt install` が `Permission denied` で失敗した | aptがファイルを読むときに使う `_apt` ユーザーが、ホームディレクトリに入れない |

### Linux・コマンド

| ドキュメント | 内容 | 主なキーワード |
|---|---|---|
| [scp-qa.md](scp-qa.md) | `scp` とは何か、`cp` との違い、コピーの向き3パターン | `scp`、SSH、`-3`、`-P` |
| [tilde-qa.md](tilde-qa.md) | `~`（チルダ）とは何か、scpでどのマシンのホームを指すか、どのOSで使えるか | `~`、ホームディレクトリ、チルダ展開、シェル |
| [apt-permission-denied-qa.md](apt-permission-denied-qa.md) | Linuxのフォルダ構成と、権限（`rwx`、`755`、`644`）のしくみ | 権限、`chmod`、`/tmp`、スティッキービット、`_apt` |
| [file-vs-folder-permission-qa.md](file-vs-folder-permission-qa.md) | **ファイルとフォルダで `rwx` の意味が違う**こと。保存に実行権限は不要なこと、「読む」「実行する」「たどり着く」の違い | ファイルの `x`（実行）、フォルダの `x`（中に入る）、nano、`chmod +x` |

> `apt-permission-denied-qa.md` は、トラブルの記録とLinuxの権限の解説を兼ねているので、両方の表に載せています。

---

## ドキュメント同士の関係

STEP 1を進める中で、どのドキュメントからどのドキュメントが生まれたかを示します。

```text
learning-exercise.md（STEP 1：UDP通信を理解する）
│
├── step1-udp-socket-qa.md ……………… UDPソケットとポート番号
│     └── unicast-broadcast-multicast.md … 送り先の種類
│
└── step1-udp-tools-runbook.md ………… STEP 1の手順書
      │
      ├── tcpdump-save-capture.md ……………… キャプチャをファイルに保存する
      │
      └── client-vm-offline-package-install.md … 0-1でaptが失敗
            ├── ssh-via-vmnet2-qa.md ……………… SSHの経路
            ├── scp-qa.md ……………………………… ファイルのコピー
            │     └── tilde-qa.md ……………………… パスの中の「~」
            └── apt-permission-denied-qa.md …… 5-3でaptが失敗、Linuxの権限
                  └── file-vs-folder-permission-qa.md … ファイルとフォルダで rwx の意味が違う
```

---

## ドキュメントの種類と名前の付け方

新しくドキュメントを追加するときの目安です。

| 種類 | ファイル名の付け方 | 例 |
|---|---|---|
| 手順書 | `stepN-<内容>-runbook.md` | `step1-udp-tools-runbook.md` |
| STEPに依存しない手順書 | `<ツール名>-<内容>.md` | `tcpdump-save-capture.md` |
| 特定のSTEPのQ&A | `stepN-<内容>-qa.md` | `step1-udp-socket-qa.md` |
| STEPに依存しないQ&A | `<内容>-qa.md` | `scp-qa.md`、`tilde-qa.md` |
| トラブルの記録 | `<起きたこと>.md` | `client-vm-offline-package-install.md` |

ドキュメントを追加したら、この README の一覧と「ドキュメント同士の関係」にも追記してください。
