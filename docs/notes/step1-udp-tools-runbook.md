# STEP 1 手順書：ツールでUDP通信を確認する

コードを書かずに、`nc`・`socat`・`tcpdump` を使って Client VM と Router VM の間でUDPを送り合い、パケットを観察する手順です。

UDPソケットやポート番号の考え方は [step1-udp-socket-qa.md](step1-udp-socket-qa.md) を参照してください。

## この手順のゴール

- [ ] Client → Server にUDPでデータが届く
- [ ] Server → Client にUDPでデータが返る
- [ ] `tcpdump` でUDPパケットの送信元・宛先を確認できる
- [ ] ブロードキャストで送ったデータをServerが受け取れる

## 全体像

```text
 Router VM（Server役）                      Client VM（Client役）
 LAN: 10.0.0.1/24                           LAN: 10.0.0.10/24
┌───────────────────────────┐              ┌───────────────────────────┐
│ ターミナルA：tcpdump で観察  │              │                           │
│ ターミナルB：9067番で待ち受け │◀── UDP ────▶│ ターミナルC：9068番から送信  │
└───────────────────────────┘    VMnet2    └───────────────────────────┘
```

| 項目 | 値 |
|---|---|
| Server（Router VM） | `10.0.0.1`、UDP **9067**番 |
| Client（Client VM） | `10.0.0.10`、UDP **9068**番 |
| ネットワーク | VMnet2（隔離LAN） |

> 本番のDHCPは67番・68番を使いますが、STEP 1では sudo が不要で、既存のDHCPクライアントともぶつからない 9067番・9068番を使います。

---

## 0. 事前準備（両方のVM）

### 0-1. ツールをインストールする

**Router VM と Client VM の両方**で実行します。

```bash
sudo apt update
sudo apt install -y netcat-openbsd socat tcpdump
```

> **Client VMで `Temporary failure resolving 'deb.debian.org'` と表示されて失敗する場合**
>
> Client VMは隔離LAN（VMnet2）にしかつながっていないので、インターネットに出られません。[client-vm-offline-package-install.md](client-vm-offline-package-install.md) の手順でインストールしてください。

確認：
```bash
nc -h 2>&1 | head -1     # OpenBSD netcat と表示されればOK
socat -V | head -1
tcpdump --version | head -1
```

### 0-2. LAN側のインターフェース名を調べる

**両方のVM**で実行します。

```bash
ip -br addr
```

出力例：
```text
lo      UNKNOWN  127.0.0.1/8
ens33   UP       192.168.xx.xx/24     ← NAT側など（VMnet2ではない）
ens34   UP                            ← VMnet2側（LAN）
```

VMnet2につながっているNICの名前をメモしておきます。以降は `<LAN_IF>` と書くので、自分の環境の名前（例：`ens34`）に読み替えてください。

| VM | `<LAN_IF>` |
|---|---|
| Router VM | （ここに記入） |
| Client VM | （ここに記入） |

> どれがVMnet2側かわからないときは、VMwareの「仮想マシン設定 → ネットワークアダプタ」で VMnet2 に割り当てたアダプタを確認し、MACアドレス（`ip link` で表示）を照らし合わせます。

### 0-3. 固定IPを設定する

`ip addr add` で付けたIPは**一時的**なもので、再起動すると消えます。

> **SSHで作業しているときは注意**
>
> Router VMを踏み台にしてClient VMにSSHしている場合、VMnet2側にはすでにIPが付いています。先に `ip -br addr` で確認してください。すでにIPが付いていれば、この手順は飛ばし、そのIPで以降の `10.0.0.1` / `10.0.0.10` を読み替えます。SSHで使っているNICのIPを変えたり消したりすると、接続が切れて戻れなくなります。詳しくは [client-vm-offline-package-install.md の7章](client-vm-offline-package-install.md#7-sshで作業するときの注意) を参照してください。

**Router VM**
```bash
sudo ip link set <LAN_IF> up
sudo ip addr add 10.0.0.1/24 dev <LAN_IF>
```

**Client VM**
```bash
sudo ip link set <LAN_IF> up
sudo ip addr add 10.0.0.10/24 dev <LAN_IF>
```

確認（両方のVM）：
```bash
ip -br addr show <LAN_IF>
```

> すでに同じIPが付いている場合は `RTNETLINK answers: File exists` と表示されます。問題ありません。

### 0-4. 疎通を確認する

**Client VM**
```bash
ping -c 3 10.0.0.1
```

`0% packet loss` ならOKです。通らない場合は [トラブルシューティング](#トラブルシューティング) を見てください。

---

## 1. パケット観察を始める（tcpdump）

**Router VM のターミナルA**で実行し、**開いたままにしておきます**。

```bash
sudo tcpdump -i <LAN_IF> -n -X 'udp port 9067 or udp port 9068'
```

| オプション | 意味 |
|---|---|
| `-i <LAN_IF>` | 観察するインターフェース |
| `-n` | IPやポートを名前に変換せず、数字のまま表示する |
| `-X` | パケットの中身を16進数と文字で表示する |
| `'udp port ...'` | 9067番か9068番のUDPだけを表示する |

`listening on <LAN_IF> ...` と表示されたら待機中です。

---

## 2. ユニキャストで送り合う（nc）

### 2-1. Serverを起動する

**Router VM のターミナルB**で実行します。

```bash
nc -u -l 9067
```

| オプション | 意味 |
|---|---|
| `-u` | UDPを使う（付けないとTCPになる） |
| `-l` | 待ち受けモード（Server役） |
| `9067` | 待ち受けるポート番号 |

何も表示されず止まったように見えますが、それで正常（受信待ち）です。

### 2-2. Clientから送る（Client → Server）

**Client VM のターミナルC**で実行します。

```bash
nc -u -p 9068 10.0.0.1 9067
```

| オプション | 意味 |
|---|---|
| `-u` | UDPを使う |
| `-p 9068` | 送信元ポートを9068番に固定する（付けないとOSが自動で選ぶ） |
| `10.0.0.1 9067` | 宛先のIPとポート |

`hello` と入力して **Enter** を押します。

**期待する結果**
- ターミナルB（Server）に `hello` と表示される
- ターミナルA（tcpdump）に次のような行が出る

```text
IP 10.0.0.10.9068 > 10.0.0.1.9067: UDP, length 6
```

> `length 6` は `hello` の5文字と改行1文字を足した数です。

- [ ] Client → Server を確認できた

### 2-3. Serverから返す（Server → Client）

**ターミナルB（Server）**で `hi` と入力して **Enter** を押します。

**期待する結果**
- ターミナルC（Client）に `hi` と表示される
- ターミナルA（tcpdump）に次のような行が出る

```text
IP 10.0.0.1.9067 > 10.0.0.10.9068: UDP, length 3
```

- [ ] Server → Client を確認できた

### 2-4. 結果から読み取ること

```text
 Client → Server:  10.0.0.10.9068  >  10.0.0.1.9067
 Server → Client:  10.0.0.1.9067   >  10.0.0.10.9068
                   └─ 送信元 ─┘       └─ 宛先 ─┘
```

- 行きと帰りで、送信元と宛先の「IP.ポート」が**入れ替わっている**
- Serverは返事の宛先を指定していない。最初に届いたパケットの送信元を覚えて、そこへ返している（`recvfrom` → `sendto` と同じ動き）
- データより前にハンドシェイクのパケット（TCPのSYNなど）が**出ていない**。UDPには接続がない

### 2-5. 終了する

ターミナルBとCで **Ctrl + C** を押します。ターミナルA（tcpdump）は次の手順でも使うので、そのままにしておきます。

> `nc -u -l` は、最初にデータを送ってきた相手とだけやりとりを続けます。別の相手で試すときは、Serverを起動し直してください。

---

## 3. ブロードキャストで送る（socat）

DHCPDISCOVERと同じように、宛先を全員（`255.255.255.255`）にして送ります。

### 3-1. Serverを起動する

**Router VM のターミナルB**で実行します。

```bash
socat -u UDP-RECV:9067 -
```

| 指定 | 意味 |
|---|---|
| `-u` | 一方向（受信して画面に出すだけ） |
| `UDP-RECV:9067` | UDP 9067番で受信する |
| `-` | 受信したデータを画面（標準出力）に出す |

### 3-2. ブロードキャストを送る

**Client VM のターミナルC**で実行します。

```bash
echo "who is server?" | socat - UDP-DATAGRAM:255.255.255.255:9067,broadcast,sourceport=9068
```

| 指定 | 意味 |
|---|---|
| `UDP-DATAGRAM:255.255.255.255:9067` | 宛先は全員、ポートは9067番 |
| `broadcast` | ソケットの `SO_BROADCAST` を有効にする |
| `sourceport=9068` | 送信元ポートを9068番にする |

**期待する結果**
- ターミナルB（Server）に `who is server?` と表示される
- ターミナルA（tcpdump）に次のような行が出る

```text
IP 10.0.0.10.9068 > 255.255.255.255.9067: UDP, length 15
```

- [ ] ブロードキャストをServerが受信できた

### 3-3. 試してみる：`broadcast` を外すとどうなるか

```bash
echo "no broadcast option" | socat - UDP-DATAGRAM:255.255.255.255:9067,sourceport=9068
```

**期待する結果**：`Permission denied` などのエラーになり、送信できない。

ブロードキャストを送るには、ソケットで `SO_BROADCAST` を明示的に有効にする必要があります。自分でコードを書くときも同じです。

### 3-4. 終了する

ターミナルBで **Ctrl + C** を押します。ターミナルA（tcpdump）も **Ctrl + C** で止めます。

---

## 4. 後片付け（必要な場合のみ）

固定IPはVMを再起動すれば消えます。すぐに外したいときは次を実行します。

**Router VM**
```bash
sudo ip addr del 10.0.0.1/24 dev <LAN_IF>
```

**Client VM**
```bash
sudo ip addr del 10.0.0.10/24 dev <LAN_IF>
```

> 次の STEP 2 でも固定IPのまま進めるので、通常は外さなくて構いません。

---

## トラブルシューティング

| 症状 | 確認すること |
|---|---|
| `ping` が通らない | ・`ip -br addr` で両方のVMにIPが付いているか<br>・NICが `UP` になっているか<br>・両方のVMのNICが VMnet2 に割り当てられているか（VMwareの設定） |
| `ping` は通るが、ncのデータが届かない | ・Serverを先に起動しているか<br>・ポート番号（9067/9068）や `-u` の付け忘れがないか<br>・ファイアウォール（`sudo nft list ruleset`）でUDPが遮断されていないか |
| tcpdumpに何も表示されない | ・`-i` に指定したのがVMnet2側のNICか<br>・フィルタのポート番号が合っているか |
| ブロードキャストが届かない | ・Client VMに複数のNICがあると、`255.255.255.255` 宛てが VMnet2 ではないNIC（デフォルトルート側）から出ることがある<br>・その場合は宛先を `10.0.0.255`（VMnet2のサブネット宛て）に変えて試す |
| `nc: invalid option` などのエラー | ・`netcat-traditional` が入っていてオプションが違う可能性がある<br>・`sudo apt install netcat-openbsd` を実行する |
| `Address already in use` | ・前に起動したncやsocatが残っている<br>・`ss -ulnp` で9067番・9068番を使っているプロセスを確認して止める |

---

## 完了チェック

- [ ] Client → Server：`hello` がServerに届いた
- [ ] Server → Client：`hi` がClientに届いた
- [ ] tcpdumpで、行きと帰りの送信元・宛先が入れ替わっていることを確認した
- [ ] tcpdumpで、UDPにはハンドシェイクがないことを確認した
- [ ] ブロードキャスト（`255.255.255.255`）がServerに届いた
- [ ] `broadcast` オプションがないと送信できないことを確認した

## 次にやること

同じ動き（9067番で待ち受けて返事をするServer、9068番から送るClient）を、`playground/udp-echo/` にコードで書いてみます。
