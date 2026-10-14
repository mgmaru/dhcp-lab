# 自作DHCPサーバー・クライアント 実行環境構成

## 1. このドキュメントの目的

このドキュメントでは、自作DHCPサーバー・DHCPクライアントを安全に実験するためのVM環境を整理します。

目的は次の3つです。

- DHCPのブロードキャスト通信を安全に観察する
- 家庭内LANなどの実ネットワークへ影響を与えない
- サーバー・クライアント間の通信を分かりやすく切り分ける

今回のDHCP演習では、**VMware上の隔離された仮想ネットワーク内で実行する**ことを前提とします。

---

# 2. 全体構成

基本構成は次の通りです。

```text
+------------------------------------------------------+
| Windows 11 Host                                     |
|                                                      |
|   VMware Workstation Pro                            |
|                                                      |
|   +----------------------+                          |
|   | Debian Router VM     |                          |
|   |                      |                          |
|   | 自作 DHCP Server     |                          |
|   |                      |                          |
|   | LAN NIC              |                          |
|   +----------+-----------+                          |
|              |                                      |
|              | VMnet2                               |
|              | 学習用・隔離LAN                      |
|              |                                      |
|   +----------+-----------+                          |
|   | Debian Client VM     |                          |
|   |                      |                          |
|   | 自作 DHCP Client     |                          |
|   |                      |                          |
|   | LAN NIC              |                          |
|   +----------------------+                          |
|                                                      |
+------------------------------------------------------+
```

ポイントは、Router VMとClient VMの間に**DHCP実験専用の仮想LAN**を作ることです。

---

# 3. 各マシンの役割

## Windows Host

WindowsはVMwareを動かすホストOSです。

今回のDHCP通信そのものは、基本的にWindows上では実行しません。

役割は、

- VMwareの起動
- Router VM / Client VMの管理
- VMnetの設定
- 必要に応じてパケット観察

です。

---

## Debian Router VM

Router VMでは、

**自作DHCPサーバー**

を実行します。

```text
Debian Router VM
      |
      +-- DHCP Server Program
      |
      +-- LAN Interface
              |
              v
            VMnet2
```

最終的には、

```text
DHCPDISCOVER
      ↓
DHCPOFFER
      ↓
DHCPREQUEST
      ↓
DHCPACK
```

のServer側処理を担当します。

---

## Debian Client VM

Client VMでは、

**自作DHCPクライアント**

を実行します。

```text
Debian Client VM
      |
      +-- DHCP Client Program
      |
      +-- LAN Interface
              |
              v
            VMnet2
```

Client側では、

1. DHCPDISCOVERを送信
2. DHCPOFFERを受信
3. DHCPREQUESTを送信
4. DHCPACKを受信

という処理を行います。

---

# 4. DHCP通信の経路

通信の流れをVM環境に当てはめると次のようになります。

```text
Debian Client VM
自作 DHCP Client
        |
        | DHCPDISCOVER
        | Broadcast
        v
=========================
        VMnet2
=========================
        |
        v
Debian Router VM
自作 DHCP Server
        |
        | DHCPOFFER
        v
=========================
        VMnet2
=========================
        |
        v
Debian Client VM
```

その後、

```text
Client
   |
   | DHCPREQUEST
   v
Server
   |
   | DHCPACK
   v
Client
```

と通信します。

---

# 5. なぜVMを使うのか

DHCPは通常のWebアプリケーションのような通信とは少し性質が違います。

DHCP Clientは最初、

```text
自分のIPアドレス
→ まだ持っていない

DHCP ServerのIPアドレス
→ まだ知らない
```

という状態です。

そのためDHCPDISCOVERでは、

**ブロードキャスト**

を利用します。

概念的には、

```text
Client
  |
  | 「DHCP Serverはいますか？」
  |
  +----> 同じLAN上の端末全体
```

という通信になります。

---

# 6. 実ネットワークで実験しない理由

例えば家庭内LANで自作DHCP Serverを起動すると、

```text
                    +--> PC
                    |
自作 DHCP Server ---+--> Smartphone
                    |
                    +--> TV
                    |
                    +--> Tablet
```

のように、意図していない端末へDHCP応答を返してしまう可能性があります。

本来の家庭用ルーターもDHCP Serverとして動いているため、

```text
Client
   |
   | DHCPDISCOVER
   |
   +----> 家庭用Router DHCP Server
   |
   +----> 自作 DHCP Server
```

となる可能性があります。

これはネットワークトラブルの原因になります。

そのため、

**DHCP実験は隔離されたVMネットワーク内で行う**

のが安全です。

---

# 7. VMnetの役割

VMwareでは、仮想マシン同士を接続するために仮想ネットワークを作れます。

今回の構成では、例として、

```text
VMnet2
```

をDHCP学習専用LANとして利用します。

```text
Router VM
    |
    |
  VMnet2
    |
    |
Client VM
```

このVMnetには、DHCP実験に関係するVMだけを接続します。

---

# 8. VMnet側のDHCP機能に注意

ここは非常に重要です。

VMwareの仮想ネットワークには、VMware自身のDHCP Serverを動かせる機能があります。

もしVMware DHCPが有効なままだと、

```text
Client VM
   |
   | DHCPDISCOVER
   |
   +--------> 自作 DHCP Server
   |
   +--------> VMware DHCP Server
```

という状態になる可能性があります。

すると、

```text
DHCPOFFER #1
自作 Server

DHCPOFFER #2
VMware DHCP
```

のように複数のOfferが届き、実験結果が分かりにくくなります。

---

# 9. 推奨するVMnet設定

学習用VMnetは、

```text
DHCP実験専用
```

にします。

推奨イメージ：

```text
VMnet2

用途:
  DHCP学習用LAN

接続:
  Debian Router VM
  Debian Client VM

VMware DHCP:
  OFF

実家庭LAN:
  接続しない
```

重要なのは、

**自作DHCP ServerだけがDHCP応答を返す環境にすること**

です。

---

# 10. Router VMのNIC構成

Router VMには、将来的には2枚のNICを持たせてもよいです。

```text
                Debian Router VM
              +-------------------+
              |                   |
Internet <----| WAN NIC           |
              |                   |
VMnet2 <------| LAN NIC           |
              |                   |
              +-------------------+
```

DHCP Serverは、

**LAN側NIC**

で待ち受けます。

今回のDHCP学習だけであれば、まずLAN側の通信だけに集中して構いません。

---

# 11. Client VMのNIC構成

Client VMは、DHCP実験用LANに接続します。

```text
Debian Client VM
      |
      | LAN NIC
      |
      v
    VMnet2
```

DHCP Clientは、このインターフェースから、

```text
DHCPDISCOVER
```

を送信します。

---

# 12. 最初は固定IPでもよい

いきなり、

```text
ClientにIPなし
↓
DHCPDISCOVER
```

から始めると、トラブル時の切り分けが難しくなる場合があります。

そのため、最初のUDP通信確認では一時的に固定IPを設定しても構いません。

例えば、

```text
Router LAN
10.0.0.1/24

Client LAN
10.0.0.10/24
```

として、

```text
Client → UDP → Server
Server → UDP → Client
```

が動くことを確認します。

---

# 13. 学習環境の段階的な変更

おすすめの進め方は次の通りです。

```text
STEP 1
固定IPあり

Router : 10.0.0.1
Client : 10.0.0.10

        ↓

UDP通信確認

        ↓

STEP 2
DHCP形式のメッセージを送信

        ↓

DISCOVER / OFFER / REQUEST / ACKを確認

        ↓

STEP 3
Client側の固定IPを外す

        ↓

本来のDHCP通信へ

        ↓

STEP 4
DHCPACKで受け取った情報を表示

        ↓

STEP 5
DHCPACKの情報をOSへ反映
```

---

# 14. 最初の実験状態

最初は次の状態でも十分です。

```text
+------------------------+
| Debian Router VM       |
|                        |
| LAN: 10.0.0.1/24       |
|                        |
| 自作 DHCP Server       |
+-----------+------------+
            |
            |
          VMnet2
            |
            |
+-----------+------------+
| Debian Client VM       |
|                        |
| LAN: 10.0.0.10/24      |
|                        |
| 自作 DHCP Client       |
+------------------------+
```

この段階では、

**DHCPの正式な動作を再現することよりUDP通信を確認すること**

を優先します。

---

# 15. DHCP本番実験時

DORAを実装できたら、Client側の固定IPを外します。

```text
+------------------------+
| Debian Router VM       |
|                        |
| LAN: 10.0.0.1/24       |
|                        |
| 自作 DHCP Server       |
+-----------+------------+
            |
            |
          VMnet2
            |
            |
+-----------+------------+
| Debian Client VM       |
|                        |
| LAN: IPなし            |
|                        |
| 自作 DHCP Client       |
+------------------------+
```

この状態から、

```text
Client
   |
   | DHCPDISCOVER
   v
Server
```

を開始します。

---

# 16. DHCP取得後の最終イメージ

最終的には、

```text
Before

Client
IP: none


       DHCPDISCOVER
             |
             v

Server
Pool:
10.0.0.100 - 10.0.0.150


       DHCPOFFER
       DHCPREQUEST
       DHCPACK
             |
             v


After

Client
IP: 10.0.0.100/24
Gateway: 10.0.0.1
DNS: ...
Lease: ...
```

という状態を目指します。

---

# 17. OSへ反映する前の段階

最初はDHCPACKを受信しても、

Linuxのネットワーク設定は変更しません。

```text
DHCPACK
   |
   v
自作 DHCP Client
   |
   v
Terminalへ表示

Assigned IP:
Subnet Mask:
Gateway:
DNS:
Lease Time:
```

これによって、

```text
DHCP Protocol
```

と、

```text
Linux Network Configuration
```

を分けて学習できます。

---

# 18. 最終段階

最終的には、

```text
DHCPACK
   |
   v
自作 DHCP Client
   |
   +--> IP Address
   |
   +--> Prefix / Subnet Mask
   |
   +--> Default Route
   |
   +--> DNS
   |
   v
Linux Network Interface
```

まで実装します。

最終環境：

```text
+--------------------------------------------------+
| Debian Router VM                                 |
|                                                  |
| LAN: 10.0.0.1/24                                 |
|                                                  |
| +----------------------+                         |
| | 自作 DHCP Server     |                         |
| +----------+-----------+                         |
+------------|-------------------------------------+
             |
             | DHCP
             |
========== VMnet2 ==================================
             |
             |
+------------|-------------------------------------+
| Debian Client VM                                 |
|                                                  |
| +----------+-----------+                         |
| | 自作 DHCP Client     |                         |
| +----------+-----------+                         |
|            |                                     |
|            v                                     |
| Linux Interface                                  |
| 10.0.0.x/24                                      |
|                                                  |
+--------------------------------------------------+
```

---

# 19. パケット観察

DHCPはパケットキャプチャと非常に相性がよい学習題材です。

例えば、

```text
Client Program
     |
     | DISCOVER
     v
Linux Kernel
     |
     v
Virtual NIC
     |
     v
VMnet2
     |
     v
Virtual NIC
     |
     v
Linux Kernel
     |
     v
Server Program
```

という流れをパケットキャプチャで確認できます。

利用候補：

- `tcpdump`
- Wireshark

---

# 20. 観察する項目

パケットを確認するときは、まず次の情報を見ます。

```text
Ethernet
├── Source MAC
└── Destination MAC

IP
├── Source IP
└── Destination IP

UDP
├── Source Port
└── Destination Port

DHCP
├── Message Type
├── Transaction ID
├── Client MAC
├── Offered IP
├── Requested IP
└── Options
```

プログラムで作った値と、実際にネットワークへ流れた値を比較すると理解が深まります。

---

# 21. トラブルシューティングの考え方

例えばClientがOfferを受信できない場合、

```text
ClientはDISCOVERを作れたか？
        |
        v
DISCOVERを送信できたか？
        |
        v
VMnet2へ流れたか？
        |
        v
Router VMへ届いたか？
        |
        v
Serverが受信したか？
        |
        v
Serverが解析できたか？
        |
        v
OFFERを作れたか？
        |
        v
OFFERを送信できたか？
        |
        v
Clientまで戻ったか？
```

というように、通信経路を1段ずつ確認します。

---

# 22. この環境で分離できる責務

VMを分けることで、

```text
DHCP Client
```

と、

```text
DHCP Server
```

を独立したプログラムとして考えられます。

```text
Client VM                         Router VM

Application                       Application
DHCP Client                       DHCP Server
     |                                |
     v                                v
UDP                              UDP
     |                                |
     v                                v
IP                               IP
     |                                |
     +---------- VMnet2 --------------+
```

実際のネットワーク構成に近い形で実験できるのが大きなメリットです。

---

# 23. 開発場所と実行場所は分けてもよい

プログラムのコードを書く場所と実行する場所は、必ずしも同じである必要はありません。

例えば、

```text
Mac / Windows
     |
     | SSH / Remote Development
     v
Debian VM
     |
     v
Build / Run
```

という構成でも構いません。

重要なのは、

**DHCP Server / Clientの実際のプロセスをVM内で動かすこと**

です。

---

# 24. 推奨構成まとめ

| 項目 | 推奨 |
|---|---|
| 仮想化 | VMware Workstation Pro |
| DHCP Server | Debian Router VM |
| DHCP Client | Debian Client VM |
| 実験用LAN | 専用VMnet |
| VMware DHCP | OFF |
| 家庭内LANへの接続 | 原則しない |
| 初期検証 | 固定IPあり |
| DHCP本番検証 | Client固定IPなし |
| パケット確認 | tcpdump / Wireshark |
| 最初のゴール | DHCPACK内容の表示 |
| 最終ゴール | Linux Interfaceへ設定反映 |

---

# 25. 安全チェックリスト

実験前に確認します。

```text
[ ] DHCP実験専用VMnetを使用している

[ ] Router VMとClient VMだけを接続している

[ ] VMwareのDHCP Serverを無効にしている

[ ] 自作DHCP Serverが家庭内LANへ接続されていない

[ ] 使用するネットワークインターフェースを確認した

[ ] Server側とClient側の実行場所を確認した

[ ] tcpdump等で通信を観察できる状態にした
```

---

# 26. 今回の環境の考え方

今回の環境で最も重要なのは、

```text
「VMを使うこと」
```

そのものではありません。

本質は、

```text
DHCP実験用ネットワークを
実ネットワークから隔離すること
```

です。

VMwareを利用すると、

```text
安全性
+
再現性
+
観察しやすさ
+
壊しても戻しやすい
```

というメリットがあります。

そのため、自作DHCPの学習環境として非常に適しています。

---

# 27. 最終構成図

```text
                        Windows 11 Host
                              |
                              |
                    VMware Workstation Pro
                              |
          +-------------------+-------------------+
          |                                       |
          |                                       |
+---------v-----------+                 +---------v-----------+
| Debian Router VM    |                 | Debian Client VM    |
|                     |                 |                     |
| 自作 DHCP Server    |                 | 自作 DHCP Client    |
|                     |                 |                     |
| LAN Interface       |                 | LAN Interface       |
| 10.0.0.1/24         |                 | 最初はIPなし        |
+---------+-----------+                 +----------+----------+
          |                                        |
          |                                        |
          +--------------- VMnet2 ----------------+
                   DHCP実験専用 仮想LAN
                   VMware DHCP: OFF
                   実LANから隔離
```

この構成の中で、

```text
DISCOVER
   ↓
OFFER
   ↓
REQUEST
   ↓
ACK
```

を一つずつ実装・観察していきます。
