# WebSocket JSON Sender / WebSocket JSON 送信ツール

English follows Japanese.

## 概要

- 指定した WebSocket URL に接続し、`Name=Value` 形式で渡した引数から JSON を構築して 1 回だけ送信します。
- 送信後は受信した JSON テキストを整形して表示します（整形できない場合はそのまま表示）。
- `ws://`/`wss://` 双方に対応。`-insecure-skip-verify` で TLS 検証をスキップ可能（テスト用途のみ推奨）。

## 前提

- Go 1.21 以降（推奨）

## ビルド

```sh
go build ./...
```

## 使い方

```sh
go run main.go -url ws://localhost -path /ws [-port 8080] [-dial-timeout 10s] [-read-timeout 10s] [-insecure-skip-verify] Name=Value [More=Data]
```

- `-url` (必須): ベース URL（例 `ws://localhost`）
- `-path` (必須): パス（例 `/ws`）
- `-port`: ポート番号を上書きしたい場合に指定
- `-dial-timeout`: 接続確立のタイムアウト
- `-read-timeout`: 送信後の受信待ちタイムアウト（`0` で無期限）
- `-insecure-skip-verify`: `wss://` 利用時にサーバ証明書検証をスキップ（テスト専用）
- 末尾の引数: `Name=Value` 形式で任意個のキー/値を渡すと JSON へまとめて送信

### 実行例

```sh
go run main.go -url ws://localhost -path /chat -port 9000 user=alice action=ping
# wss の例（自己署名の場合のみ -insecure-skip-verify を付与）
# go run main.go -url wss://example.com -path /ws -insecure-skip-verify user=alice action=ping
```

---

## Overview (English)

- Connects to a WebSocket URL, builds JSON from `Name=Value` CLI args, and sends it once right after connect.
- Pretty-prints received JSON text (falls back to raw text if not valid JSON).
- Supports both `ws://` and `wss://`. `-insecure-skip-verify` skips TLS verification (testing only).

## Requirements

- Go 1.21+

## Build

```sh
go build ./...
```

## Usage

```sh
go run main.go -url ws://localhost -path /ws [-port 8080] [-dial-timeout 10s] [-read-timeout 10s] [-insecure-skip-verify] Name=Value [More=Data]
```

- `-url` (required): Base URL, e.g. `ws://localhost`
- `-path` (required): Path, e.g. `/ws`
- `-port`: Override port if needed
- `-dial-timeout`: Timeout when establishing the connection
- `-read-timeout`: Timeout for receiving after send (`0` waits indefinitely)
- `-insecure-skip-verify`: For `wss://`, skip TLS verification (testing only)
- Trailing args: any number of `Name=Value` pairs to merge into the JSON body

### Example

```sh
go run main.go -url ws://localhost -path /chat -port 9000 user=alice action=ping
# wss example (add -insecure-skip-verify only for self-signed certs)
# go run main.go -url wss://example.com -path /ws -insecure-skip-verify user=alice action=ping
```
