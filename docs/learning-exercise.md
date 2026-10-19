# 自作DHCPサーバー・クライアント 学習演習

## 1. この演習の目的

この演習では、DHCP（Dynamic Host Configuration Protocol）の仕組みを理解するために、簡易的な **DHCPサーバー** と **DHCPクライアント** を自分で実装します。

完成済みライブラリに処理を任せるのではなく、

- DHCPクライアントは何を送るのか
- DHCPサーバーは何を判断するのか
- DHCPメッセージには何が入っているのか
- なぜブロードキャストが必要なのか
- IPアドレスはどのように決まるのか
- DHCPACKを受信した後、OSへどのように反映されるのか

を順番に理解することが目的です。

> この資料には具体的な実装コードは載せません。  
> 処理の流れと設計上のポイントを見ながら、自分で実装してください。

---

# 2. DHCPとは

DHCPは、ネットワークに参加した端末へ自動的にネットワーク設定を配布するためのプロトコルです。

代表的には次の情報を配布できます。

| 情報 | 例 |
|---|---|
| IPアドレス | `10.0.0.100` |
| サブネットマスク | `255.255.255.0` |
| デフォルトゲートウェイ | `10.0.0.1` |
| DNSサーバー | `8.8.8.8` |
| リース時間 | 3600秒 |

DHCPがない場合、各端末にこれらを手作業で設定する必要があります。

---

# 3. DHCPサーバーとクライアントの役割

## DHCPクライアント

ネットワーク設定を「もらう側」です。

例：

- PC
- スマートフォン
- 仮想マシン
- IoT機器

今回の演習では、Debian Client VMがDHCPクライアントの役割を担当します。

## DHCPサーバー

ネットワーク設定を「配る側」です。

主な役割は、

1. クライアントから要求を受け取る
2. 配布可能なIPアドレスを選ぶ
3. クライアントへ候補を提示する
4. クライアントの要求を確認する
5. IPアドレスの利用を許可する

ことです。

今回の演習では、Debian Router VMがDHCPサーバーの役割を担当します。

---

# 4. DHCP通信の基本：DORA

DHCPで最も重要なのが、次の4段階です。

```text
DHCP Client                         DHCP Server
     |                                   |
     | ------ DHCPDISCOVER ------------> |
     |                                   |
     | <------- DHCPOFFER --------------- |
     |                                   |
     | ------- DHCPREQUEST -------------> |
     |                                   |
     | <-------- DHCPACK ---------------- |
     |                                   |
```

頭文字を取って **DORA** と呼ばれます。

| 順番 | メッセージ | 意味 |
|---:|---|---|
| 1 | DHCPDISCOVER | DHCPサーバーを探す |
| 2 | DHCPOFFER | IPアドレス候補を提案する |
| 3 | DHCPREQUEST | 提案されたIPを使いたいと要求する |
| 4 | DHCPACK | サーバーが利用を許可する |

---

# 5. DHCPDISCOVER

## クライアント側の考え方

DHCPクライアントは、ネットワークに接続した直後には自分のIPアドレスを持っていません。

さらに、

> DHCPサーバーがどこに存在するのか

も知りません。

そのため、特定の相手ではなくネットワーク全体へ問い合わせます。

これが **ブロードキャスト** です。

概念的には、

```text
IPを持っていないClient
        |
        | 「DHCPサーバーはいますか？」
        v
LAN全体へブロードキャスト
```

となります。

## クライアントが行う処理

1. DHCPメッセージを作成する
2. メッセージ種別をDHCPDISCOVERにする
3. Transaction IDを生成する
4. 自分のMACアドレスなど必要な情報を格納する
5. DHCPサーバー向けのUDPポートへ送信する
6. DHCPOFFERを待つ

---

# 6. DHCPOFFER

DHCPサーバーがDHCPDISCOVERを受信すると、

> このIPアドレスなら使えます

という候補を返します。

例えば、

```text
Client
  |
  | DHCPDISCOVER
  v
Server

Server
  |
  | 「10.0.0.100 を使えます」
  v
Client
```

## サーバー側の処理

1. DHCPDISCOVERを受信する
2. DHCPメッセージを解析する
3. Transaction IDを確認する
4. クライアントを識別する
5. 使用可能なIPアドレスを選ぶ
6. DHCPOFFERを作成する
7. 候補IPアドレスや各種ネットワーク情報を入れる
8. クライアントへ返す

---

# 7. DHCPREQUEST

DHCPOFFERを受信したクライアントは、

> そのIPアドレスを使いたいです

という要求を送ります。

```text
Server
  |
  | 「10.0.0.100 を使えます」
  v
Client

Client
  |
  | 「10.0.0.100 を使いたいです」
  v
Server
```

## クライアント側の処理

1. DHCPOFFERを受信する
2. メッセージを解析する
3. Transaction IDが自分の通信のものか確認する
4. 提案されたIPアドレスを確認する
5. DHCPREQUESTを作成する
6. 利用したいIPアドレスを指定する
7. DHCPサーバーへ送信する
8. DHCPACKを待つ

---

# 8. DHCPACK

DHCPサーバーはDHCPREQUESTを受信したら、要求されたIPアドレスを本当に貸し出してよいか確認します。

問題がなければDHCPACKを返します。

```text
Client
  |
  | 「10.0.0.100 を使いたいです」
  v
Server

Server
  |
  | 「10.0.0.100 を使ってよいです」
  v
Client
```

## サーバー側の処理

1. DHCPREQUESTを受信する
2. Transaction IDなどを確認する
3. 要求されたIPアドレスを確認する
4. IPアドレスをクライアントへ割り当てる
5. DHCPACKを作成する
6. ネットワーク設定を含める
7. クライアントへ返す

---

# 9. DHCPクライアントがDHCPACKを受信した後

今回の演習では、ここを2段階に分けます。

## 第1段階：受信内容を表示する

最初は、OSのネットワーク設定を変更しません。

DHCPACKを解析して、例えば次の情報を表示できれば成功です。

```text
Assigned IP
Subnet Mask
Default Gateway
DNS Server
Lease Time
DHCP Server
```

この段階では、

```text
DHCPプロトコルを理解する
```

ことに集中します。

## 第2段階：OSのネットワーク設定へ反映する

DORAが完成してから、取得した情報をLinuxのネットワークインターフェースへ設定します。

```text
DHCPACK
   |
   v
DHCP Client Program
   |
   +--> IP Address
   |
   +--> Subnet Mask
   |
   +--> Default Route
   |
   +--> DNS
   |
   v
Linux Network Configuration
```

この段階になると、DHCPだけではなくOSとの連携も学習対象になります。

---

# 10. DHCPで使用するUDPポート

DHCPではUDPを使用します。

| 役割 | UDPポート |
|---|---:|
| DHCP Server | 67 |
| DHCP Client | 68 |

クライアントはまだ通常のIP通信を行える状態ではないため、DHCPでは通常のアプリケーション通信とは少し異なる扱いが必要です。

---

# 11. DHCPメッセージで重要なフィールド

DHCPはBOOTPを基礎にしたメッセージ形式を使用しています。

最初からすべてを完璧に理解する必要はありません。

まずは次の項目を意識してください。

| フィールド | 役割 |
|---|---|
| op | Request / Replyの区別 |
| htype | ハードウェア種別 |
| hlen | ハードウェアアドレス長 |
| xid | Transaction ID |
| flags | ブロードキャスト等の制御 |
| ciaddr | Client IP Address |
| yiaddr | Your IP Address |
| siaddr | Server IP Address |
| giaddr | Relay Agent Address |
| chaddr | Client Hardware Address |
| options | DHCP固有の追加情報 |

---

# 12. Transaction IDの役割

Transaction IDは非常に重要です。

DHCPクライアントはDHCPDISCOVERを送信するときにTransaction IDを作ります。

例えば、

```text
Transaction ID = A
```

として送った場合、

```text
DISCOVER  xid=A
     |
     v
Server

OFFER     xid=A
     |
     v
Client
```

のように同じIDが返ってきます。

これによってクライアントは、

> このDHCPOFFERは自分のDHCPDISCOVERに対する返答だ

と判断できます。

複数のクライアントが同時にDHCP通信をしていても、通信を区別できます。

---

# 13. MACアドレスが重要になる理由

DHCPを開始する時点では、クライアントにはまだIPアドレスがありません。

そのため、

```text
IPアドレスだけでクライアントを識別する
```

ことができません。

そこでMACアドレスなどの情報を使ってクライアントを識別します。

```text
Client A
MAC: AA:AA:AA:AA:AA:AA

Client B
MAC: BB:BB:BB:BB:BB:BB
```

DHCPサーバーは、

```text
MAC A → 10.0.0.100
MAC B → 10.0.0.101
```

のように管理できます。

---

# 14. DHCP Options

DHCPでは、固定部分だけではなく **Options** にさまざまな情報を格納できます。

代表例：

| Option | 意味 |
|---|---|
| DHCP Message Type | DISCOVER / OFFER / REQUEST / ACKなど |
| Subnet Mask | サブネットマスク |
| Router | デフォルトゲートウェイ |
| Domain Name Server | DNSサーバー |
| Requested IP Address | クライアントが要求するIP |
| IP Address Lease Time | リース時間 |
| DHCP Server Identifier | DHCPサーバー識別情報 |

今回の演習では、すべてのOptionを実装する必要はありません。

必要なものから少しずつ増やしてください。

---

# 15. IPアドレスの貸し出し範囲

DHCPサーバーは通常、配布するIPアドレスの範囲を持っています。

例えば、

```text
Network

10.0.0.0/24

Router
10.0.0.1

DHCP Pool
10.0.0.100
    ↓
10.0.0.150
```

この範囲を **DHCP Pool** として管理します。

最初の演習では、もっと単純にして構いません。

例えば、

```text
最初のClient → 10.0.0.100
```

だけでもDORAを理解できます。

その後、

```text
Client A → 10.0.0.100
Client B → 10.0.0.101
Client C → 10.0.0.102
```

のように拡張します。

---

# 16. リース

DHCPで割り当てられたIPアドレスは、通常永久に使えるわけではありません。

一定時間だけ貸し出されます。

これを **Lease（リース）** と呼びます。

例えば、

```text
10.0.0.100

Lease Time
3600 seconds
```

であれば、そのIPアドレスの利用には時間制限があります。

本格的なDHCPクライアントでは、

```text
IP取得
  |
  v
一定時間利用
  |
  v
リース更新
  |
  +--> 成功 → 継続
  |
  +--> 失敗 → 再試行
                 |
                 v
             最終的に破棄
```

という処理が必要になります。

今回の最初の実装では、リース更新までは実装しなくても構いません。

> リース更新の流れ（T1 でユニキャスト、T2 でブロードキャストに切り替える、期限切れで最初からやり直す）は、[dhcp-lease-renew-rebind-qa.md](notes/qa/dhcp-lease-renew-rebind-qa.md) にまとめています。

---

# 17. 演習全体の進め方

この演習では、一度に完全なDHCPを作ろうとしません。

次の順番で進めます。

```text
STEP 1
UDP通信

   ↓

STEP 2
DHCPメッセージの生成・解析

   ↓

STEP 3
DHCPDISCOVER

   ↓

STEP 4
DHCPOFFER

   ↓

STEP 5
DHCPREQUEST

   ↓

STEP 6
DHCPACK

   ↓

STEP 7
DORA完成

   ↓

STEP 8
複数クライアント対応

   ↓

STEP 9
リース管理

   ↓

STEP 10
Linuxインターフェースへ反映
```

---

# 18. STEP 1：UDP通信を理解する

## 目標

DHCPを実装する前に、

```text
Client → Server
Server → Client
```

というUDP通信を成立させます。

この段階ではDHCP形式でなくても構いません。

## 確認すること

- UDPソケットとは何か
- Server側はどのポートで待つのか
- Client側はどのポートを使うのか
- UDPでは接続という概念がTCPとどう違うのか
- ブロードキャストをどう扱うのか

## 完了条件

Clientから送ったデータをServerが受信できる。

Serverから返したデータをClientが受信できる。

---

# 19. STEP 2：DHCPメッセージを表現する

## 目標

DHCPメッセージをプログラム内で扱えるようにします。

考えるべき処理は、

```text
プログラム内部のデータ
        |
        v
バイト列
        |
        v
UDP送信
```

と、

```text
UDP受信
   |
   v
バイト列
   |
   v
プログラム内部のデータ
```

です。

## 学習ポイント

- バイナリプロトコル
- バイト列
- 固定長フィールド
- 可変長フィールド
- エンディアン
- 整数とバイト列の変換
- IPアドレスのバイト表現

## 完了条件

自分で作ったDHCP形式のメッセージを、

```text
生成 → バイト列化 → 復元
```

できる。

---

# 20. STEP 3：DHCPDISCOVERを実装する

## Client

Clientは、

1. Transaction IDを生成する
2. 自分のMACアドレスを取得する
3. DHCPDISCOVERメッセージを作る
4. DHCP Message Typeを設定する
5. ブロードキャストする
6. DHCPOFFERを待つ

## Server

Serverは、

1. UDPポートで待機する
2. パケットを受信する
3. DHCPメッセージとして解析する
4. DHCP Message Typeを確認する
5. DHCPDISCOVERであることを判定する
6. Transaction IDやMACアドレスを表示する

## 完了条件

Server側で、

```text
DHCPDISCOVER received
Transaction ID: ...
Client MAC: ...
```

のような情報を確認できる。

---

# 21. STEP 4：DHCPOFFERを実装する

## Server

ServerはDISCOVERを受け取ったら、

1. 使用可能なIPを選ぶ
2. DHCPOFFERを作成する
3. DISCOVERと同じTransaction IDを設定する
4. 候補IPアドレスを入れる
5. DHCP Server Identifierを入れる
6. 必要なネットワーク情報を入れる
7. Clientへ送る

## Client

Clientは、

1. DHCPOFFERを受信する
2. Transaction IDを確認する
3. OfferされたIPアドレスを確認する
4. DHCP Server Identifierを確認する
5. 内容を表示する

## 完了条件

Clientが、

```text
Offered IP: 10.0.0.100
```

のように候補IPを認識できる。

---

# 22. STEP 5：DHCPREQUESTを実装する

## Client

ClientはOfferを受け取ったら、

1. 使用したいIPを決定する
2. DHCPREQUESTを作成する
3. Requested IP Addressを設定する
4. DHCP Server Identifierを設定する
5. Transaction IDを維持する
6. Serverへ送る

## Server

Serverは、

1. DHCPREQUESTを受信する
2. Transaction IDを確認する
3. Clientを確認する
4. Requested IP Addressを確認する
5. 自分がOfferしたIPか確認する

## 完了条件

Serverが、

```text
Client requested 10.0.0.100
```

と認識できる。

---

# 23. STEP 6：DHCPACKを実装する

## Server

Serverは要求されたIPを確認して問題がなければ、

1. DHCPACKを作成する
2. Transaction IDを維持する
3. 割り当てIPを設定する
4. Subnet Maskを設定する
5. Router情報を設定する
6. DNS情報を設定する
7. Lease Timeを設定する
8. Clientへ送る

## Client

Clientは、

1. DHCPACKを受信する
2. Transaction IDを確認する
3. DHCP Message Typeを確認する
4. Assigned IPを読み取る
5. Subnet Maskを読み取る
6. Routerを読み取る
7. DNSを読み取る
8. Lease Timeを読み取る
9. 内容を表示する

## 完了条件

Client側で、

```text
DHCP configuration acquired

IP Address:
Subnet Mask:
Default Gateway:
DNS:
Lease Time:
```

を正しく表示できる。

---

# 24. STEP 7：DORA全体を完成させる

ここまで完成すると、

```text
Client                                 Server
  |                                      |
  | DISCOVER                             |
  |------------------------------------->|
  |                                      |
  |                              IP選択   |
  |                                      |
  | OFFER                                |
  |<-------------------------------------|
  |                                      |
  | IP候補確認                            |
  |                                      |
  | REQUEST                              |
  |------------------------------------->|
  |                                      |
  |                            要求確認   |
  |                                      |
  | ACK                                  |
  |<-------------------------------------|
  |                                      |
  | 設定情報表示                          |
```

という一連のDHCP通信を自分で実装したことになります。

---

# 25. STEP 8：複数クライアントへの対応

DORAが動いてから実装してください。

## 課題

2台以上のClientが存在した場合、

```text
Client A → 10.0.0.100
Client B → 10.0.0.101
```

のように異なるIPを渡せるようにします。

## 考えること

- Clientを何で識別するか
- 使用済みIPをどう管理するか
- 未使用IPをどう見つけるか
- 同じIPを2台へ配らないためにはどうするか
- Server再起動時に情報をどうするか

---

# 26. STEP 9：リース管理

次にリースを追加します。

Server側では、

```text
Client
MAC Address
Assigned IP
Lease Start
Lease End
```

などを管理できるようにします。

## 考えること

- リースが切れたIPを再利用してよいのはいつか
- 同じClientが再び来たとき同じIPを渡すか
- Clientが更新要求を送ってきたらどうするか
- Serverを再起動したらリース情報をどうするか

---

# 27. STEP 10：LinuxインターフェースへIPを反映する

最後に、

```text
DHCPACK
   |
   v
DHCP Client
   |
   v
Linux Network Interface
```

を実装します。

ここではDHCPプロトコルだけでなく、OSのネットワーク管理も扱います。

## 反映する情報

最低限、

- IPアドレス
- Prefix / Subnet Mask

を反映します。

さらに、

- Default Gateway
- DNS

も反映できると、実用的な構成に近づきます。

## 学習ポイント

- Linuxのネットワークインターフェース
- Kernelが管理するIP設定
- Routing Table
- Default Route
- DNS設定
- 管理者権限
- NetlinkなどのOS API

---

# 28. なぜ最初からOSへIPを設定しないのか

IPアドレスを設定するだけなら、それほど難しい処理ではありません。

しかし、実際のDHCPクライアントとして考えると、

```text
DHCPACK
   |
   +--> IP設定
   |
   +--> Route設定
   |
   +--> DNS設定
   |
   +--> Lease管理
   |
   +--> Lease更新
   |
   +--> 更新失敗処理
   |
   +--> IP削除
```

まで責務が広がります。

そのため、

```text
DHCPプロトコルの理解

と

OSネットワーク設定の理解
```

を分けて学習します。

---

# 29. 推奨する実験環境

今回のようなDHCP実験は、隔離された仮想ネットワークで行うのがおすすめです。

```text
Windows Host
    |
    +-----------------------------+
    | VMware                      |
    |                             |
    |   Debian Router VM          |
    |   DHCP Server               |
    |       |                     |
    |       | LAN                 |
    |       |                     |
    |   Debian Client VM          |
    |   DHCP Client               |
    |                             |
    +-----------------------------+
```

Router VMとClient VMの間に、学習用の専用LANを用意します。

---

# 30. 重要：既存DHCPサーバーを混在させない

DHCP実験では特に注意が必要です。

同じLAN上に、

```text
自作DHCP Server
```

と

```text
VMwareのDHCP Server
```

などが同時に存在すると、

```text
Client
  |
  | DISCOVER
  |
  +----> 自作Server
  |
  +----> VMware DHCP
```

となり、複数のDHCPOFFERが返る可能性があります。

実験用LANでは既存DHCPを無効にし、自作DHCPだけが応答する状態にしてください。

また、自作DHCPを家庭や会社などの実ネットワークへ接続しないでください。

---

# 31. パケットキャプチャを活用する

自分のプログラムだけを見ていると、

```text
送信側がおかしいのか
受信側がおかしいのか
パケットそのものがおかしいのか
```

を切り分けにくいことがあります。

そこで、

- Wireshark
- tcpdump

などのパケットキャプチャツールを利用すると理解しやすくなります。

確認するとよい情報：

- DHCP Message Type
- Transaction ID
- Client MAC Address
- Offered IP
- Requested IP
- Server Identifier
- UDP Source Port
- UDP Destination Port
- Broadcastかどうか

---

# 32. デバッグするときの考え方

DHCP通信が動かない場合、DORA全体を一度に疑わないようにします。

例えば、

```text
DISCOVERは送れたか？
        ↓
Serverまで届いたか？
        ↓
Serverは正しく解析できたか？
        ↓
OFFERを作れたか？
        ↓
OFFERはClientまで届いたか？
        ↓
Clientは解析できたか？
```

というように、1段階ずつ確認します。

これはDHCP以外のネットワークプログラムでも重要な考え方です。

---

# 33. 最初の完成ライン

最初の目標はここです。

```text
Client
  |
  | DISCOVER
  v
Server
  |
  | OFFER
  v
Client
  |
  | REQUEST
  v
Server
  |
  | ACK
  v
Client

IP: 10.0.0.100
Mask: ...
Gateway: ...
DNS: ...
Lease: ...
```

この時点では実際にLinuxへIPを設定しなくても構いません。

ここまでできれば、

- UDP
- Broadcast
- DHCP
- MAC Address
- IP Address
- Transaction ID
- Binary Protocol
- Client / Server
- State Transition

を実際のプログラムを通して理解できます。

---

# 34. 発展課題

基本実装が完成した後は、次の順番で拡張すると理解が深まります。

## 発展1：DHCPNAK

Serverが要求を拒否する場合を実装する。

## 発展2：IP Pool

複数のIPアドレスを管理する。

## 発展3：複数Client

同時に複数クライアントへIPを配る。

## 発展4：Lease

IPアドレスに利用期限を持たせる。

## 発展5：Lease Renewal

Clientからリース更新を行う。

## 発展6：Lease Persistence

Serverを再起動しても割当情報を保持する。

## 発展7：OSへの反映

DHCPACKを受信したらLinuxへIPを設定する。

## 発展8：実際のDHCPクライアントとの通信

自作ServerとLinux標準DHCP Clientを通信させる。

または、

自作Clientと既存DHCP Serverを通信させる。

これができると、自作実装が標準的なDHCPと互換性を持っていることを確認できます。

---

# 35. この演習で学べること

この演習はDHCPだけを学ぶものではありません。

最終的には次の知識がつながります。

```text
Application
    |
    | DHCP
    v
UDP
    |
    v
IP
    |
    v
Ethernet
    |
    v
Network Interface
```

さらにプログラム側では、

```text
DHCP
 |
 +-- Client / Server
 |
 +-- UDP Socket
 |
 +-- Binary Protocol
 |
 +-- Endianness
 |
 +-- State Machine
 |
 +-- Timeout
 |
 +-- Resource Management
 |
 +-- OS Network Configuration
```

を経験できます。

---

# 36. 最終ゴール

最終的には、

```text
Debian Client VM
        |
        | DHCPDISCOVER
        v
Debian Router VM
自作DHCP Server
        |
        | DHCPOFFER
        v
Debian Client VM
        |
        | DHCPREQUEST
        v
Debian Router VM
        |
        | DHCPACK
        v
Debian Client VM
        |
        v
Linux Interface
10.0.0.x/24
        |
        v
通信可能
```

というところまで、自分のコードで動かすことを目標とします。

最初から全部を実装する必要はありません。

**DORAを1段階ずつ完成させ、最後にOSへ接続する。**

この順番で進めることで、DHCPの仕組みとネットワークプログラミングの両方を整理しながら学習できます。
