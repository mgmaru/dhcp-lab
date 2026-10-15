# Client VMにパッケージをインストールできないときの対処

STEP 1の手順書（[step1-udp-tools-runbook.md](../runbooks/step1-udp-tools-runbook.md)）の「0-1. ツールをインストールする」で、Client VMの `apt install` が失敗しました。このドキュメントは、そのとき何が起こったのか、原因、対処の手順をまとめた記録です。

## 目次

- [1. 何が起こったのか](#1-何が起こったのか)
- [2. 原因](#2-原因)
- [3. 原因を確かめる方法](#3-原因を確かめる方法)
- [4. 対処法の選び方](#4-対処法の選び方)
- [5. 対処法A：Router VMでダウンロードしてClient VMにコピーする](#5-対処法arouter-vmでダウンロードしてclient-vmにコピーする)
- [6. 対処法B：一時的にNATのネットワークアダプタを追加する](#6-対処法b一時的にnatのネットワークアダプタを追加する)
- [7. SSHで作業するときの注意](#7-sshで作業するときの注意)
- [8. まとめ](#8-まとめ)

---

## 1. 何が起こったのか

### 状況

- 外出先から、Router VMを踏み台にしてClient VMにSSHで入っていた
- Client VMで、手順書のとおりにツールをインストールしようとした

```text
 手元のPC ──SSH──▶ Router VM ──SSH（VMnet2）──▶ Client VM
                   （踏み台）                     ここで apt install を実行
```

### 実行したコマンド

```bash
sudo apt install -y netcat-openbsd socat tcpdump
```

### 出たエラー（抜粋）

```text
Installing:
  netcat-openbsd  socat  tcpdump

Installing dependencies:
  libpcap0.8t64
...
Err:4 http://deb.debian.org/debian trixie/main amd64 tcpdump amd64 4.99.5-2
  Temporary failure resolving 'deb.debian.org'
...
Error: Failed to fetch http://deb.debian.org/debian/pool/main/t/tcpdump/tcpdump_4.99.5-2_amd64.deb  Temporary failure resolving 'deb.debian.org'
Error: Unable to fetch some archives, maybe run apt-get update or try with --fix-missing?
```

### エラーの読み方

| 部分 | 意味 |
|---|---|
| `Installing: ...` | インストールする予定のパッケージ。ここまではClient VMの中だけで処理できている |
| `Ign:` | 取得を試みたが、うまくいかなかった（あとで再試行する） |
| `Err:` | 取得に失敗した |
| `Temporary failure resolving 'deb.debian.org'` | `deb.debian.org` という名前から**IPアドレスを調べられなかった**（名前解決の失敗） |
| `Failed to fetch ...` | 名前が引けないので、パッケージをダウンロードできなかった |

> **パッケージ名やバージョンは表示されたのに、なぜダウンロードできないのか？**
>
> `apt` は、手元にある**パッケージ一覧**（Debianをインストールしたときや、前回 `apt update` を実行したときに取得したもの）を見て、「何を入れるか」を決めます。ここまではインターネットが不要です。実際のファイル（`.deb`）をダウンロードする段階で、初めてインターネットが必要になります。

---

## 2. 原因

**Client VMがインターネットにつながっていないこと**が原因です。

[vm-environment.md](../../vm-environment.md) の構成では、Client VMは**隔離LAN（VMnet2）にしかつながっていません**。VMnet2は、DHCPの実験が家庭内LANなどに影響しないように、わざと外に出られないようにしたネットワークです。

```text
            Internet
               ▲
               │ WAN NIC
 ┌─────────────┴─────────────┐         ┌───────────────────────────┐
 │ Router VM                 │ VMnet2  │ Client VM                 │
 │ インターネットに出られる     │◀───────▶│ VMnet2にしかつながっていない │
 └───────────────────────────┘ （隔離） │ → インターネットに出られない  │
                                       └───────────────────────────┘
```

つまり、これは**設計どおりの動き**で、設定ミスではありません。ただ、パッケージをインストールするときだけは、別の手段でファイルを用意する必要があります。

> Router VMはVMnet2とつながっていますが、ルーター（NAT）としての設定はしていません。そのため、Client VMの通信を外へ中継することはできません。

---

## 3. 原因を確かめる方法

Client VMで次のコマンドを実行します。

```bash
ip -br addr          # どのNICに、どのIPが付いているか
ip route             # インターネットへの出口（default の行）があるか
ping -c 2 8.8.8.8    # インターネット上のIPアドレスに届くか
cat /etc/resolv.conf # 名前解決に使うDNSサーバー
echo $SSH_CONNECTION # SSHがどのIPを使って接続しているか
```

| 確認結果 | 判断 |
|---|---|
| `ip route` に `default via ...` の行がなく、`ping 8.8.8.8` も失敗する | **インターネットへの出口がない**（今回のケース） |
| `ping 8.8.8.8` は成功するが、`ping deb.debian.org` は失敗する | インターネットには出られるが、**名前解決（DNS）だけ失敗**している。`/etc/resolv.conf` を見直す |

> `echo $SSH_CONNECTION` は「接続元IP 接続元ポート 接続先IP 接続先ポート」を表示します。接続先IPを見ると、SSHがどのNICを使っているかわかります。[7章](#7-sshで作業するときの注意)の注意で使います。

---

## 4. 対処法の選び方

| 状況 | おすすめの対処法 |
|---|---|
| 外出先などから**SSHで作業している**（VMwareを操作できない） | **[対処法A](#5-対処法arouter-vmでダウンロードしてclient-vmにコピーする)**：Router VMでダウンロードしてコピー |
| Windows Hostの前にいて、**VMwareを操作できる** | 対処法AかB。どちらでもよい |

> ネットワークアダプタの追加や変更は**VMware（Windows Host）側の操作**です。Client VMの中からはできません。今回は外出先からSSHで作業していたので、対処法Aを選びました。

---

## 5. 対処法A：Router VMでダウンロードしてClient VMにコピーする

今回選んだ方法です。VMwareの設定を変えずに済みます。

### 考え方

インターネットに出られるRouter VMでパッケージのファイル（`.deb`）をダウンロードし、VMnet2を通ってClient VMにコピーしてから、Client VMでインストールします。

```text
 Internet
    │ ① ダウンロード
    ▼
 Router VM ──② コピー（scp、VMnet2経由）──▶ Client VM ──③ インストール
```

### 前提条件

- Router VMがインターネットに出られること
- Router VMとClient VMが、**同じDebianのバージョン**（今回は trixie）で、**同じアーキテクチャ**（今回は amd64）であること
  - 違うと、ダウンロードしたパッケージがClient VMで使えないことがあります
- Router VMとClient VMが、VMnet2でSSH（scp）できること

確認コマンド（両方のVMで実行し、結果が同じであることを確かめます）：
```bash
grep VERSION_CODENAME /etc/os-release   # 例：VERSION_CODENAME=trixie
dpkg --print-architecture               # 例：amd64
```

### 5-1. Router VM：パッケージをダウンロードする

Router VMにSSHした状態で実行します。

```bash
sudo apt update
mkdir -p ~/debs && cd ~/debs
apt download netcat-openbsd socat tcpdump libpcap0.8t64
ls ~/debs
```

| コマンド | 意味 |
|---|---|
| `sudo apt update` | パッケージ一覧を最新にする |
| `mkdir -p ~/debs && cd ~/debs` | ダウンロード先のフォルダを作って移動する |
| `apt download ...` | パッケージのファイル（`.deb`）を、**インストールせずに**今のフォルダへダウンロードする |

次のように `.deb` ファイルが4つ並べば成功です（バージョンは異なる場合があります）。

```text
libpcap0.8t64_1.10.5-2_amd64.deb
netcat-openbsd_1.229-1_amd64.deb
socat_1.8.0.3-1+deb13u1_amd64.deb
tcpdump_4.99.5-2_amd64.deb
```

> **なぜこの4つなのか？**
>
> Client VMのエラー画面で「Installing」と「Installing dependencies」に表示されたのが、この4つだったからです。`libpcap0.8t64` は tcpdump が使うライブラリ（依存パッケージ）です。ほかに必要なライブラリは、Client VMにすでに入っています。
>
> 別のパッケージを入れるときは、Client VMで `sudo apt install <パッケージ名>` を一度実行し、失敗する前に表示される一覧を見れば、何をダウンロードすればよいかわかります。

ついでに、Router VM自身にもツールをインストールしておきます（手順書の0-1はRouter VMでも必要です）。

```bash
sudo apt install -y netcat-openbsd socat tcpdump
```

### 5-2. .debファイルをClient VMにコピーする

コピーの方法は3つあります。自分の環境に合わせて**どれか1つ**を使ってください。

> コピーには `scp`（SSH越しにファイルをコピーするコマンド）を使います。`scp` の書き方は [scp-qa.md](../qa/scp-qa.md) にまとめています。

| 方法 | どこで実行するか | 向いている場合 |
|---|---|---|
| ① Client VMから取ってくる | Client VM | Client VMからRouter VMにパスワードでログインできる |
| ② Router VMから送る | Router VM | Router VMからClient VMにパスワードでログインできる |
| ③ 手元のPCから中継する | 手元のPC | 手元のPCの鍵でしかVMにログインできない |

以下の `<...>` は、自分の環境に合わせて読み替えてください。

| 表記 | 意味 | 例 |
|---|---|---|
| `<Routerのユーザー名>` | Router VMのログインユーザー | `kajih` |
| `<Clientのユーザー名>` | Client VMのログインユーザー | `kajih` |
| `<RouterのVMnet2側IP>` | Router VMのVMnet2側のIP | `10.0.0.1` |
| `<ClientのIP>` | Client VMのVMnet2側のIP（踏み台から接続しているIP） | `10.0.0.10` |
| `<Routerへの接続先>` | 手元のPCからRouter VMに入るときの接続先（`~/.ssh/config` のホスト名など） | `router` |

**方法①：Client VMからRouter VMのファイルを取ってくる**

Client VMで実行します。Router VMのログインパスワードを聞かれます。

```bash
mkdir -p ~/debs
scp <Routerのユーザー名>@<RouterのVMnet2側IP>:~/debs/*.deb ~/debs/
```

**方法②：Router VMからClient VMへ送る**

Router VMで実行します。Client VMのログインパスワードを聞かれます。

```bash
ssh <Clientのユーザー名>@<ClientのIP> 'mkdir -p ~/debs'
scp ~/debs/*.deb <Clientのユーザー名>@<ClientのIP>:~/debs/
```

**方法③：手元のPCから、Router VMを経由してコピーする**

手元のPCで実行します。

```bash
ssh -J <Routerへの接続先> <Clientのユーザー名>@<ClientのIP> 'mkdir -p ~/debs'
scp -3 -o ProxyJump=<Routerへの接続先> \
  <Routerへの接続先>:~/debs/*.deb \
  <Clientのユーザー名>@<ClientのIP>:~/debs/
```

| オプション | 意味 |
|---|---|
| `-J` / `-o ProxyJump=...` | 指定したホストを踏み台にして接続する |
| `-3` | リモート同士のコピーを、手元のPCを中継して行う。VM同士でSSHの認証ができなくても使える |

コピーできたか確認します（Client VMで実行）。

```bash
ls ~/debs
```

4つの `.deb` ファイルが並べばOKです。

### 5-3. Client VM：インストールする

Client VMで実行します。

```bash
mkdir -p /tmp/debs
cp ~/debs/*.deb /tmp/debs/
chmod 755 /tmp/debs
chmod 644 /tmp/debs/*.deb
sudo apt install -y /tmp/debs/*.deb
```

> `apt install` にファイルのパス（`./` や `/` で始まるもの）を渡すと、インターネットからではなく、そのファイルからインストールします。

> **なぜ `/tmp/debs` にコピーし直すのか**
>
> `sudo apt install -y ~/debs/*.deb` のようにホームディレクトリのファイルを直接指定すると、`(13: Permission denied)` で失敗します。aptはファイルを読むときに `_apt` という専用ユーザーに切り替えますが、`_apt` は `/home/kajih` の中に入れないためです。詳しくは [apt-permission-denied-qa.md](apt-permission-denied-qa.md) を参照してください。
>
> `/tmp/debs` にコピーせず、`sudo dpkg -i ~/debs/*.deb` でインストールすることもできます。

### 5-4. インストールを確認する

Client VMで実行します。

```bash
nc -h 2>&1 | head -1      # OpenBSD netcat と表示されればOK
socat -V | head -1
tcpdump --version | head -1
```

3つとも表示されれば完了です。

### 5-5. 後片付け（任意）

ダウンロードしたファイルは、もう使いません。必要なら削除します。

```bash
rm -r ~/debs     # Router VM と Client VM の両方
rm -r /tmp/debs  # Client VM（再起動で消える場合もある）
```

---

## 6. 対処法B：一時的にNATのネットワークアダプタを追加する

Windows Hostの前にいて、VMwareを操作できるときの方法です。インストールしている間だけClient VMをインターネットにつなぎ、終わったら元に戻します。

> **SSHで作業しているときは使わないでください。** ネットワークの設定を変えると、SSHの接続が切れることがあります。

### 6-1. VMwareでNATアダプタを追加する

1. VMwareで Client VM を選び、「仮想マシン設定」を開く
2. 「追加」→「ネットワークアダプタ」を選ぶ
3. 接続を **NAT** にして OK
   - VMを起動したままでも追加できます

### 6-2. Client VMで、追加したNICにIPを付けてインストールする

```bash
ip -br addr                 # 新しく増えたNICの名前を確認（例：ens37）
sudo dhclient <NAT_IF>      # VMwareのNATからIPをもらう
ping -c 2 deb.debian.org    # 名前が引けて、届くか確認
sudo apt update
sudo apt install -y netcat-openbsd socat tcpdump
```

> `dhclient` が見つからないと表示された場合は、`sudo networkctl reconfigure <NAT_IF>` か `sudo nmcli device connect <NAT_IF>` を試してください。Debianのインストール方法によって、使われているネットワーク管理ツールが違います。

### 6-3. インストールが終わったら、NATアダプタを外す

1. Client VMで、もらったIPを返す
   ```bash
   sudo dhclient -r <NAT_IF>
   ```
2. VMwareの「仮想マシン設定」で、追加したネットワークアダプタを**削除**する

> **必ず外してください。** NICが2枚あるままだと、次の問題が起こります。
>
> - 手順書の手順3で、`255.255.255.255` 宛てのブロードキャストがNAT側のNICから出てしまい、Router VMに届かないことがある
> - 「隔離LANで実験する」という前提が崩れる

### 別のやり方：既存のアダプタを一時的にNATへ切り替える

VMwareでVMnet2側アダプタの接続先を **VMnet2 → NAT** に変えてインストールし、終わったら **NAT → VMnet2** に戻す方法もあります。ただし、戻し忘れるとRouter VMと通信できなくなるので注意してください。

---

## 7. SSHで作業するときの注意

今回のように、Router VMを踏み台にしてClient VMにSSHしているときは、次の点に気をつけてください。

> SSHの通信がどの経路を通っているのか、`$SSH_CONNECTION` の読み方などは [ssh-via-vmnet2-qa.md](../qa/ssh-via-vmnet2-qa.md) にまとめています。

### SSHで使っているNICの設定を変えない

SSHの通信は、VMnet2を通っています。そのNICのIPを変えたり消したり、NICを止めたり（`ip link set ... down`）すると、**SSHの接続が切れて、外出先からは戻れなくなります**。

```text
 手元のPC ──▶ Router VM ──VMnet2──▶ Client VM
                           ▲
                           └ ここのIPを変える・消すと、SSHが切れる
```

### 手順書の「0-3. 固定IPを設定する」の前に、今のIPを確認する

すでにSSHでVM同士が通信できているので、VMnet2側にはIPが付いている可能性が高いです。まず `ip -br addr` で確認してください。

| 確認結果 | どうするか |
|---|---|
| すでに `10.0.0.1`（Router）と `10.0.0.10`（Client）が付いている | 0-3は飛ばしてよい |
| 違うIPが付いている | **そのIPのまま**、手順書の `10.0.0.1` / `10.0.0.10` を読み替えて進める |

> SSHで作業しているときに、IPの付け直しはしないでください。どうしても変えたいときは、Windows HostのVMwareコンソールから操作できるときに行います。

---

## 8. まとめ

- **何が起こったか**：Client VMで `apt install` を実行したら、`Temporary failure resolving 'deb.debian.org'` で失敗した
- **原因**：Client VMは隔離LAN（VMnet2）にしかつながっておらず、インターネットに出られない。これは**設計どおり**
- **今回選んだ対処**：インターネットに出られるRouter VMで `apt download` し、`scp` でClient VMにコピーしてから、`/tmp/debs` に置き直して `sudo apt install /tmp/debs/*.deb` でインストールする
- **VMwareを操作できるとき**：一時的にNATアダプタを追加してインストールし、**終わったら必ず外す**
- **SSHで作業しているとき**：SSHで使っているNICのIPや状態を変えない
