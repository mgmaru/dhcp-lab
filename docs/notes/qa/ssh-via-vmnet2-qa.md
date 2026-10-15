# Q&A：Router VMからClient VMへのSSHはどこを通っているのか

[client-vm-offline-package-install.md](../troubleshooting/client-vm-offline-package-install.md) を読んで出た質問と、その回答をまとめたメモです。

外出先からRouter VMを踏み台にしてClient VMにSSHしているとき、通信がどの経路を通っているのか、`$SSH_CONNECTION` の結果をどう読むのかを扱います。

関連ドキュメント：

- [vm-environment.md](../../vm-environment.md)：VM環境・隔離LANの構成
- [step1-udp-socket-qa.md](step1-udp-socket-qa.md)：エフェメラルポートとDHCPのポート番号（Q2）
- [step1-udp-tools-runbook.md](../runbooks/step1-udp-tools-runbook.md)：ツールでUDP通信を確認する手順書

## 目次

- [Q1. Router VMからClient VMにSSHできるのは、VMnet2を通っているからか？](#q1-router-vmからclient-vmにsshできるのはvmnet2を通っているからか)
- [Q2. `$SSH_CONNECTION` の結果はどう読むのか？](#q2-ssh_connection-の結果はどう読むのか)
- [わかったこと（自分の環境）](#わかったこと自分の環境)
- [まとめ](#まとめ)

---

## Q1. Router VMからClient VMにSSHできるのは、VMnet2を通っているからか？

### 結論

**はい、そのとおりです。**

Client VMのNICは、VMnet2側の1枚だけです。Client VMに届く経路はVMnet2しかないので、Router VMからClient VMへのSSHは必ずVMnet2を通っています。

### 2つの区間に分けて見る

外出先からClient VMに入るまでの通信は、2つの区間に分かれています。

```text
 手元のPC ──── 区間① ────▶ Router VM ──── 区間② ────▶ Client VM
           インターネット側              VMnet2（隔離LAN）
           （WAN NIC）                  （LAN NIC）
```

| 区間 | 通る経路 | Router VMのどのNICを使うか |
|---|---|---|
| ① 手元のPC → Router VM | インターネット側 | WAN NIC |
| ② Router VM → Client VM | **VMnet2** | LAN NIC（VMnet2側） |

### Router VMが踏み台になれる理由

Router VMはNICを**2枚**持っていて、両方のネットワークにつながっています。

```text
                Internet
                   ▲
                   │ WAN NIC
          ┌────────┴────────┐
          │    Router VM    │  ← 両方のネットワークにつながっている
          └────────┬────────┘
                   │ LAN NIC
                   ▼
                VMnet2
                   │
          ┌────────┴────────┐
          │    Client VM    │  ← VMnet2にしかつながっていない
          └─────────────────┘
```

| VM | 外（インターネット側）から届くか | VMnet2から届くか |
|---|---|---|
| Router VM | 届く | 届く |
| Client VM | **届かない** | 届く |

両方から届くのはRouter VMだけです。そのため、外からClient VMに入るには、Router VMを**中継地点（踏み台）**にする必要があります。

### 「隔離LAN」の意味

VMnet2は「隔離LAN」と呼んでいますが、**中のVM同士は自由に通信できます**。隔離されているのは、VMnet2と外の世界の間だけです。

```text
 ┌──────────── VMnet2（仮想のスイッチ）────────────┐
 │                                                │
 │   Router VM（LAN NIC） ◀────────▶ Client VM     │  ← 中同士はつながる
 │                                                │
 └────────────────────────────────────────────────┘
                    ✖ 外（インターネット・家庭内LAN）へは出られない
```

VMnet2は、Router VMとClient VMのNICが挿さった**仮想のスイッチ**だと考えるとわかりやすいです。この1つの構成から、次の2つのことが起こっています。

| 起こったこと | 理由 |
|---|---|
| Router VMからClient VMにSSHできる | 2台が同じスイッチ（VMnet2）につながっているから |
| Client VMで `apt install` が失敗した | そのスイッチから外へ出る口がないから |

### Router VMはどうやってNICを選んでいるのか

Router VMは、宛先のIPアドレスを見て、どのNICから出すかを決めています。この判断に使う表を**ルーティングテーブル**と呼びます。

```text
宛先が 10.0.0.x（VMnet2のネットワーク） → LAN NIC から出す
それ以外（インターネットなど）           → WAN NIC から出す（default）
```

ルーティングテーブルは `ip route` で表示できます。

### 自分の環境で確かめる方法

**Router VM**：Client VMへの通信がどのNICから出るかを表示します。

```bash
ip route get 10.0.0.10
```

出力例：

```text
10.0.0.10 dev ens34 src 10.0.0.1 ...
```

| 部分 | 意味 |
|---|---|
| `dev ens34` | VMnet2側のNIC（LAN NIC）から出る。NICの名前は環境によって違う |
| `src 10.0.0.1` | 送信元はRouter VMのVMnet2側のIP |

**Client VM**：今のSSHがどこから来ているかを表示します。

```bash
echo $SSH_CONNECTION
```

結果の読み方は、次のQ2で説明します。

---

## Q2. `$SSH_CONNECTION` の結果はどう読むのか？

> Client VMで `echo $SSH_CONNECTION` を実行したら、`10.0.0.1 40166 10.0.0.10 22` と出た。これは「10.0.0.1 がSSHクライアントで、10.0.0.10 の22番ポートで待っているSSHサーバーに接続しに来た」という意味でよいか？

### 実行結果

```text
$ echo $SSH_CONNECTION
10.0.0.1 40166 10.0.0.10 22
```

### 結論

**はい、その認識で合っています。**

`$SSH_CONNECTION` は、SSHサーバーがログインしたときに設定する環境変数です。今のSSH接続について、4つの値を「接続元 → 接続先」の順に並べています。

```text
10.0.0.1   40166   10.0.0.10   22
└─接続元IP─┘└接続元ポート┘└─接続先IP─┘└接続先ポート┘
  Router VM   OSが自動で選んだ   Client VM   SSHサーバー
```

| 値 | 意味 |
|---|---|
| `10.0.0.1` | 接続元のIP。Router VMのVMnet2側のIP |
| `40166` | 接続元のポート。**OSが自動で選んだエフェメラルポート** |
| `10.0.0.10` | 接続先のIP。Client VMのVMnet2側のIP |
| `22` | 接続先のポート。SSHサーバー（sshd）が待ち受ける**決まったポート** |

図にすると次のようになります。

```text
 Router VM (10.0.0.1)                          Client VM (10.0.0.10)
┌───────────────────────┐                     ┌───────────────────────┐
│ 接続元                 │                     │ 接続先                 │
│ ポート 40166            │ ───── VMnet2 ─────▶ │ ポート 22（sshd）       │
│ （OSが自動で選んだ）     │       TCP           │ （決まったポート）       │
└───────────────────────┘                     └───────────────────────┘
```

### これまでの話とのつながり①：エフェメラルポートの実例

[step1-udp-socket-qa.md のQ2](step1-udp-socket-qa.md#q2-dhcpではクライアント側のポートも68番に決まっているのか) で、「普通はサーバー側のポートが固定で、クライアント側のポートはOSが自動で決める（エフェメラルポート）」という話をしました。今回の結果は、まさにその実例です。

| | サーバー側のポート | クライアント側のポート |
|---|---|---|
| SSH（今回の結果） | 22番（固定） | 40166番（OSが自動で選んだ） |
| DHCP | 67番（固定） | **68番（固定）** |

> `40166` は毎回同じとは限りません。SSHでつなぎ直すと、別の番号になることがあります。一方、DHCPはクライアント側も68番に固定されています。並べてみると、DHCPが特別なルールで動いていることがよくわかります。

### これまでの話とのつながり②：VMnet2を通っていることの確認

接続元が `10.0.0.1`（Router VMのVMnet2側のIP）なので、Q1で説明したとおり、Router VMから**VMnet2を通って**Client VMに来ていることが確かめられました。

### 補足：「SSHクライアント」が厳密にはどこで動いているか

SSHでの入り方によって、SSHクライアント（`ssh` コマンド）が動いている場所が少し違います。

| 入り方 | `ssh` コマンドが動いている場所 | `10.0.0.1` からの接続を作っているもの |
|---|---|---|
| Router VMにログインしてから `ssh 10.0.0.10` を実行 | Router VM | Router VMの `ssh` コマンド |
| 手元のPCで `ssh -J <Router> 10.0.0.10` を実行（ProxyJump） | **手元のPC** | Router VMのSSHサーバーが、手元のPCに代わって接続を中継している |

```text
【Router VMで ssh を実行した場合】
 手元のPC ──SSH──▶ Router VM [ssh コマンド] ──SSH──▶ Client VM

【手元のPCで ssh -J を実行した場合】
 手元のPC [ssh コマンド] ══ Router VMの中を通るトンネル ══▶ Client VM
                           （Router VMのsshdが 10.0.0.1 から接続を中継）
```

どちらの入り方でも、Client VMから見た接続元は `10.0.0.1` です。そのため、「10.0.0.1 から 10.0.0.10 の22番に接続してきた」という理解で問題ありません。

---

## わかったこと（自分の環境）

今回の確認で、次のことがわかりました。

| 項目 | 結果 |
|---|---|
| Router VMのVMnet2側のIP | `10.0.0.1` |
| Client VMのVMnet2側のIP | `10.0.0.10` |
| SSHの経路 | Router VM → VMnet2 → Client VM |

### 手順書への影響

IPは手順書の想定（Router `10.0.0.1` / Client `10.0.0.10`）と同じです。

- [手順書](../runbooks/step1-udp-tools-runbook.md)の「**0-3. 固定IPを設定する**」は**飛ばしてかまいません**
- このIPはSSHでも使っているので、**付け直したり消したりしないでください**。SSHが切れて、外出先から戻れなくなります

---

## まとめ

- Client VMはVMnet2にしかつながっていないので、Router VMからClient VMへのSSHは**必ずVMnet2を通る**
- Router VMはWAN側とVMnet2側の両方につながっているので、外からClient VMに入るための**踏み台**になれる
- VMnet2は「外に出られない仮想のスイッチ」。中のVM同士は通信できるが、外へは出られない
  - だからSSHは通るのに、`apt install` は失敗した
- `$SSH_CONNECTION` は「**接続元IP 接続元ポート 接続先IP 接続先ポート**」の順に表示される
- SSHのクライアント側のポート（例：`40166`）は**エフェメラルポート**。DHCPのクライアント側（68番固定）とは違う
- 自分の環境のIPは手順書の想定どおりなので、手順書の0-3は飛ばしてよい
