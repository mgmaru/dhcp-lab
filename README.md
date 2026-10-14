# dhcp-lab
A hands-on DHCP implementation project for learning UDP, packet formats, address leasing, and Linux networking.

## ディレクトリ構成

実装言語はまだ決めていないため、言語に依存しない構成にしています。`message` や `handler` などは役割を表すモジュール名です。拡張子やテストの置き場所は、採用した言語の慣習に合わせて調整します。

```text
dhcp-lab/
├── README.md
├── docs/                      # 学習資料・設計メモ
│   ├── learning-exercise.md   # 演習の進め方（STEP 1〜10）
│   ├── vm-environment.md      # VM環境・隔離LANの構成
│   └── notes/                 # STEPごとの学習メモ・ハマりどころ
├── playground/                # 本実装の前に試す使い捨ての実験コード
│   └── udp-echo/              # STEP 1: DHCP形式を使わないUDP送受信の確認
├── dhcp/                      # 共通: DHCPメッセージの表現（ソケット・OSに依存しない）
│   ├── message                # 固定長ヘッダ（op, xid, flags, yiaddr, chaddr ...）の生成・解析
│   ├── options                # Options（TLV形式）の生成・解析
│   └── constants              # ポート番号, Message Type, Magic Cookie など
├── server/                    # DHCPサーバー（Router VMで実行）
│   ├── main                   # エントリポイント・引数処理
│   ├── transport              # UDP 67番での受信と応答の送信
│   ├── handler                # DISCOVER→OFFER, REQUEST→ACK/NAK の判断
│   ├── pool                   # 配布するIPアドレス範囲の管理
│   ├── lease                  # MAC↔IPの割当・リース期限・永続化
│   └── config                 # 設定ファイルの読み込み
├── client/                    # DHCPクライアント（Client VMで実行）
│   ├── main                   # エントリポイント・取得結果の表示
│   ├── transport              # UDP 68番からのブロードキャスト送信・受信・タイムアウト
│   ├── state                  # 状態遷移（INIT → SELECTING → REQUESTING → BOUND）
│   └── netconf                # 取得した設定をLinuxへ反映（IP・Route・DNS）
├── configs/
│   └── server.example         # サーバー設定例（サブネット, プール範囲, Router, DNS, リース時間）
├── scripts/                   # VM上で使う補助スクリプト
│   ├── capture.sh             # tcpdumpでDHCP（UDP 67/68）をキャプチャ
│   └── static-ip.sh           # STEP 1用の固定IPの設定・解除
├── tests/
│   ├── dhcp/                  # メッセージ・Optionsの生成⇔解析
│   ├── server/                # プール・リース割当のロジック
│   └── testdata/              # 実パケットから取り出したバイト列など
└── captures/                  # pcapの保存先（Git管理外）
```

### 設計方針

- **プロトコル層とI/Oを分ける**: `dhcp/` はバイト列とプログラム内部のデータの相互変換だけを担当し、ソケットやOSには触れません。そのため STEP 2 はVMを使わずに単体テストで確認できます。
- **サーバーとクライアントは別プログラムにする**: 実行するVMが異なるため、エントリポイントを分けます。両者が共有するのは `dhcp/` だけです（`server → dhcp ← client`）。
- **「判断」と「送受信」を分ける**: `server/handler` は「受信したメッセージから返すメッセージを決める」処理として書き、`transport` とは切り離します。ネットワークがなくてもDORAの判断ロジックをテストできます。
- **OSへの反映は `client/netconf` に閉じ込める**: Linux専用で root 権限が必要な処理を1か所にまとめます。DHCPACKの内容を表示するだけの第1段階では使いません。
- **実パケットをテストに使う**: tcpdump / Wireshark で取得したDHCPパケットのバイト列を `tests/testdata/` に置き、自作の解析処理と突き合わせます。

### 演習STEPとの対応

| STEP | 内容 | 主に扱うディレクトリ |
|---:|---|---|
| 1 | UDP通信 | `playground/udp-echo/`, `scripts/static-ip.sh` |
| 2 | DHCPメッセージの生成・解析 | `dhcp/`, `tests/dhcp/` |
| 3〜6 | DISCOVER / OFFER / REQUEST / ACK | `server/handler`, `server/transport`, `client/state`, `client/transport` |
| 7 | DORA完成 | `server/`, `client/`, `scripts/capture.sh` |
| 8 | 複数クライアント対応 | `server/pool`, `server/lease` |
| 9 | リース管理 | `server/lease`, `configs/` |
| 10 | Linuxインターフェースへ反映 | `client/netconf` |
| 発展 | NAK / リース更新 / リース永続化 | `server/handler`, `client/state`, `server/lease` |
