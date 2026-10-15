# 手順書：tcpdumpのキャプチャ結果をファイルに保存する

tcpdumpで観察したパケットを、ファイルに保存して、あとから見直すための手順です。

[step1-udp-tools-runbook.md](step1-udp-tools-runbook.md) の「1. パケット観察を始める（tcpdump）」を読んで出た「キャプチャの結果をファイルに出力する方法はあるか？」という質問への回答をもとにしています。

関連ドキュメント：

- [step1-udp-tools-runbook.md](step1-udp-tools-runbook.md)：STEP 1の手順書（tcpdumpの基本的な使い方）
- [apt-permission-denied-qa.md](apt-permission-denied-qa.md)：権限分離と、ホームディレクトリの権限
- [scp-qa.md](scp-qa.md)：scpでファイルをコピーする方法

## 目次

- [1. 保存する方法は2通り](#1-保存する方法は2通り)
- [2. 方法1：pcapファイルに保存する（おすすめ）](#2-方法1pcapファイルに保存するおすすめ)
- [3. 方法2：画面の表示をテキストファイルに保存する](#3-方法2画面の表示をテキストファイルに保存する)
- [4. 手元のPCにコピーしてWiresharkで見る](#4-手元のpcにコピーしてwiresharkで見る)
- [5. トラブルシューティング](#5-トラブルシューティング)
- [6. まとめ](#6-まとめ)

---

## 1. 保存する方法は2通り

| 方法 | 保存される形式 | 向いている使い方 |
|---|---|---|
| ① `-w` オプションでpcapファイルに保存する（**おすすめ**） | **pcap形式**（パケットそのもの。バイナリ） | あとでtcpdumpやWiresharkで、何度でも詳しく見直す |
| ② 画面の表示をテキストファイルに保存する | テキスト（画面に出たとおりの文字） | メモやドキュメントにそのまま貼る |

```text
【方法1】pcapファイル
 ネットワーク ──▶ tcpdump -w ──▶ step1.pcap（パケットそのもの）
                                   │
                                   ├─▶ tcpdump -r で読む（見方を変えて何度でも）
                                   └─▶ Wireshark で開く

【方法2】テキストファイル
 ネットワーク ──▶ tcpdump ──▶ 画面に表示 ──▶ tee ──▶ step1.txt（画面に出た文字だけ）
```

> **pcap形式がおすすめな理由**
>
> pcapファイルにはパケットそのものが保存されています。そのため、あとから `-X` で中身を見たり、`-e` でMACアドレスを見たり、Wiresharkで開いたりと、**見方を変えて何度でも読み直せます**。テキストは、保存したときの見方でしか読めません。

---

## 2. 方法1：pcapファイルに保存する（おすすめ）

### 2-1. 保存先のフォルダを作る

キャプチャするVMで実行します。

```bash
mkdir -p /tmp/captures
```

> **なぜホームディレクトリではなく `/tmp` なのか**
>
> `~/step1.pcap` のようにホームディレクトリに直接保存しようとすると、`Permission denied` で失敗することがあります。理由は [2-5](#2-5-注意ホームディレクトリに直接保存すると失敗する) で説明します。

### 2-2. 保存する

```bash
sudo tcpdump -i <LAN_IF> -n -w /tmp/captures/step1.pcap 'udp port 9067 or udp port 9068'
```

| オプション | 意味 |
|---|---|
| `-i <LAN_IF>` | キャプチャするインターフェース（VMnet2側のNIC） |
| `-n` | IPやポートを名前に変換せず、数字のまま扱う |
| `-w ファイル名` | 画面に表示する代わりに、パケットをファイルに保存する |
| `'udp port ...'` | 9067番か9068番のUDPだけを保存する |

- 実行中は、画面に**何も表示されません**（ファイルに保存しているため）。
- 通信を送り終わったら **`Ctrl + C`** で止めます。保存したパケットの数が表示されます。

```text
^C4 packets captured
4 packets received by filter
0 packets dropped by kernel
```

> ファイルの拡張子は `.pcap` にするのが一般的です。Wiresharkなどのツールが、拡張子で形式を判断しやすくなります。

### 2-3. 保存しながら画面にも表示する

`--print` を付けると、ファイルに保存しながら、いつもどおり画面にも表示します。

```bash
sudo tcpdump -i <LAN_IF> -n --print -w /tmp/captures/step1.pcap 'udp port 9067 or udp port 9068'
```

| オプション | 意味 |
|---|---|
| `--print` | `-w` でファイルに保存するときも、画面に表示する（Debian 13 のtcpdump 4.99で使える） |

> 手順書の「1. パケット観察を始める」で、観察しながら保存もしたいときは、このコマンドを使ってください。

### 2-4. 保存したファイルを読む

```bash
tcpdump -n -r /tmp/captures/step1.pcap
```

| オプション | 意味 |
|---|---|
| `-r ファイル名` | ネットワークの代わりに、ファイルからパケットを読む |

- 画面に出る内容は、いつものtcpdumpと同じです。
- ファイルを読むだけなので、`sudo` は**不要**です。

**見方を変えて読み直す例**

| やりたいこと | コマンド |
|---|---|
| 中身を16進数と文字で見る | `tcpdump -n -X -r /tmp/captures/step1.pcap` |
| MACアドレスも見る | `tcpdump -n -e -r /tmp/captures/step1.pcap` |
| 9067番宛てだけに絞る | `tcpdump -n -r /tmp/captures/step1.pcap 'udp dst port 9067'` |

> 保存したあとでも、フィルタ（`'udp dst port 9067'` など）で絞り込めます。

### 2-5. 注意：ホームディレクトリに直接保存すると失敗する

次のように、ホームディレクトリに直接保存しようとすると失敗することがあります。

```bash
sudo tcpdump -i <LAN_IF> -w ~/step1.pcap
```

```text
tcpdump: /home/kajih/step1.pcap: Permission denied
```

#### 原因：tcpdumpが途中で専用ユーザーに切り替わる

**aptのときと同じしくみ（権限分離）**です（[apt-permission-denied-qa.md の4章](apt-permission-denied-qa.md#4-q1-aptを実行したら自動で-_apt-に降格させられたのか)）。

```text
 sudo tcpdump -w ~/step1.pcap
   │
   ├─ ① ネットワークの準備（パケットを受け取れるようにする） …… root
   │
   │   ここで、root から tcpdump ユーザーに切り替わる
   │
   └─ ② ファイルを作って保存する ……………………………………… tcpdump ユーザー
                                                              │
                                                              ▼
                                               /home/kajih/ （kajih 本人しか入れない）
                                                → Permission denied
```

| 処理 | 実行するユーザー | 理由 |
|---|---|---|
| ① ネットワークの準備 | `root` | パケットを受け取るには、root の権限が必要 |
| ② ファイルの保存 | `tcpdump` | ネットワークから来るデータは信用できない。万一、悪意のあるパケットでtcpdumpの不具合を突かれても、被害を小さくするため |

Debianのtcpdumpは、①が終わると、`tcpdump` という権限の弱い専用ユーザーに切り替わります。`tcpdump` ユーザーは「その他」にあたるので、本人しか入れない `/home/kajih`（`drwx------`）に書き込めません。

#### 対処法

| 対処法 | コマンド | 保存したファイルの持ち主 |
|---|---|---|
| A. `/tmp` に保存する（この手順書の方法） | `sudo tcpdump ... -w /tmp/captures/step1.pcap` | `tcpdump` |
| B. `-Z` で、切り替え先を自分のユーザーにする | `sudo tcpdump -Z $USER ... -w ~/step1.pcap` | 自分（`kajih`） |

**対処法Bの例**

```bash
mkdir -p ~/captures
sudo tcpdump -Z $USER -i <LAN_IF> -n --print -w ~/captures/step1.pcap 'udp port 9067 or udp port 9068'
```

| 部分 | 意味 |
|---|---|
| `-Z ユーザー名` | ネットワークの準備が終わったら、このユーザーに切り替える |
| `$USER` | 自分のユーザー名が入った環境変数。`sudo` が実行される**前に**、シェルが `kajih` に置き換える |

> **対処法Bのよいところ**
>
> - ホームディレクトリに保存できるので、**再起動しても消えません**
> - ファイルの持ち主が自分（`kajih`）になるので、あとで削除やコピーがしやすくなります

---

## 3. 方法2：画面の表示をテキストファイルに保存する

### 3-1. 保存する

```bash
mkdir -p /tmp/captures
sudo tcpdump -i <LAN_IF> -n -X -l 'udp port 9067 or udp port 9068' | tee /tmp/captures/step1.txt
```

| 部分 | 意味 |
|---|---|
| `-X` | パケットの中身を16進数と文字で表示する |
| `-l` | 1行ずつすぐに出力する。付けないと、まとめて出力されるので、画面への表示が遅れる |
| `\| tee ファイル名` | 画面に表示しながら、同じ内容をファイルにも保存する |

```text
 tcpdump ──▶ tee ──┬──▶ 画面に表示
                   └──▶ /tmp/captures/step1.txt に保存
```

止めるときは **`Ctrl + C`** を押します。

> このとき、ファイルを書いているのは tcpdump ではなく、**自分の権限で動いている `tee`** です。そのため、ホームディレクトリ（`~/step1.txt`）にもそのまま保存できます。

### 3-2. 保存したファイルを読む

```bash
cat /tmp/captures/step1.txt
```

テキストなので、`cat` や `nano` で読んだり、ドキュメントにそのまま貼ったりできます。

> **テキストで保存するときの注意**
>
> 画面に出た情報しか残りません。例えば `-X` を付けずに保存すると、あとからパケットの中身を見ることはできません。あとで詳しく調べる可能性があるなら、方法1（pcap）で保存してください。

---

## 4. 手元のPCにコピーしてWiresharkで見る

pcapファイルは、**Wireshark**（パケットを画面で見やすく表示するツール）で開けます。VMにはWiresharkを入れていないので、手元のPCにコピーしてから開きます。

### 4-1. 手元のPCにコピーする

**手元のPCで実行**します。`<Routerへの接続先>` は、手元のPCからRouter VMに入るときの接続先（`~/.ssh/config` のホスト名など）に読み替えてください。

**Client VMで保存した場合**（Router VMを踏み台にする）

```bash
scp -J <Routerへの接続先> kajih@10.0.0.10:/tmp/captures/step1.pcap .
```

**Router VMで保存した場合**

```bash
scp <Routerへの接続先>:/tmp/captures/step1.pcap .
```

> 最後の `.` は「今いるフォルダ」という意味です。scpの書き方は [scp-qa.md](scp-qa.md) を参照してください。

### 4-2. このリポジトリでの置き場所

このリポジトリには、キャプチャファイルを置くための `captures/` フォルダがあります。`.gitignore` で除外しているので、pcapファイルを置いても**Gitにはコミットされません**。

```bash
scp -J <Routerへの接続先> kajih@10.0.0.10:/tmp/captures/step1.pcap ~/Developer/dhcp-lab/captures/
```

### 4-3. Wiresharkで開く

1. 手元のPCにWiresharkをインストールする（<https://www.wireshark.org/>）
2. Wiresharkを起動し、「ファイル」→「開く」で `step1.pcap` を選ぶ

Wiresharkでは、パケットを1つ選ぶと、Ethernet・IP・UDPなどの層ごとに中身が分かれて表示されます。tcpdumpの `-X` よりも、どの部分が何を表しているかがわかりやすくなります。

---

## 5. トラブルシューティング

| 症状 | 原因と対処 |
|---|---|
| `Permission denied` で保存できない | ホームディレクトリに保存しようとしている。`/tmp/captures` に保存するか、`-Z $USER` を付ける（[2-5](#2-5-注意ホームディレクトリに直接保存すると失敗する)） |
| `-w` を付けたら、画面に何も表示されない | `-w` は画面に表示しないのが普通の動き。表示もしたいときは `--print` を付ける |
| `-r` で読んだら、パケットが1つもない | キャプチャ中に通信が流れていなかった。tcpdumpを**先に**起動してから、nc や socat で送る |
| テキストファイルへの保存が遅れる・空のまま | `-l` を付け忘れている |
| 再起動したら `/tmp/captures` のファイルが消えた | Debian 13 では `/tmp` は再起動で空になる。残したいファイルは、手元のPCにコピーするか、`-Z $USER` でホームディレクトリに保存する |
| `scp` で `/tmp/captures/step1.pcap` をコピーできない | ファイルの持ち主が `tcpdump` で、自分が読めない可能性がある。`ls -l /tmp/captures` で確認し、読めなければ `sudo chmod 644 /tmp/captures/step1.pcap` を実行する |

---

## 6. まとめ

- **`-w ファイル名`** でpcap形式で保存し、**`-r ファイル名`** で読み直す
- **`--print`** を付けると、保存しながら画面にも表示できる
- ホームディレクトリに直接保存すると、tcpdumpが専用ユーザーに切り替わるため失敗することがある
  - **`/tmp` に保存する**か、**`-Z $USER`** を付ける
- テキストで残したいなら、**`-l`** を付けて **`| tee ファイル名`** で保存する
- pcapファイルは、手元のPCにコピーして**Wireshark**で開ける。リポジトリの `captures/` に置いても、Gitにはコミットされない
