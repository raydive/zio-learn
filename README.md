# ZIO-learn

ZIO のドキュメントにあるコードを貼り付けて試す、ローカル用の小さな Scala プレイグラウンドです。

ブラウザのエディタ → Go サーバー → 依存定義付きの一時ファイル → `scala-cli run` という構成です。画面には標準出力・標準エラー・コンパイルエラーを順次表示します。フロントエンドは HTML / CSS / JavaScript、サーバーは Go 標準ライブラリだけで動作します。

## Docker で起動

```sh
docker compose up --build
```

[http://localhost:8080](http://localhost:8080) にアクセスしてください。停止は `Ctrl+C`、コンテナの削除は `docker compose down` です。

初回ビルドでは JDK・Scala CLI・Scala コンパイラ・設定済みの ZIO ライブラリを取得し、キャッシュをイメージに含めます。事前取得は Dockerfile が `docker/cache-scala.sh` を実行して行い、Go サーバーは関与しません。ビルドにはインターネット接続が必要です。エディタで依存を追加した場合などは、実行時にもダウンロードが発生します。

イメージを直接使う場合:

```sh
docker build -t zio-learn:local .
docker run --rm --init -p 127.0.0.1:8080:8080 --memory=2g --cpus=2 --pids-limit=256 zio-learn:local
```

コンテナには Go サーバー・画面・JDK 21・Scala CLI 1.18.0 が含まれます。ホスト側への Go / Scala / Java のインストールは不要です。Scala CLI は JVM 版を使用するため、同じ Dockerfile で amd64 / arm64 を扱えます。

## GitHub Container Registry

GitHub Actions の `CI` ワークフローで `make check` に成功した後、`ghcr.io/raydive/zio-learn` へ `linux/amd64` と `linux/arm64` のイメージを公開します。

| 実行条件 | イメージタグ |
| --- | --- |
| `main` への push / `main` からの手動実行 | `latest`、`sha-<コミットSHA>` |
| `v1.2.3` などのバージョンタグの push | `1.2.3`、`sha-<コミットSHA>` |
| Pull Request | 検証のみ。公開しません。 |

イメージは `ubuntu-24.04`（amd64）と `ubuntu-24.04-arm`（arm64）の標準ランナーで並列にビルドします。Go のビルドと JVM / Scala CLI による依存の事前取得を各アーキテクチャでネイティブ実行し、QEMU は使いません。Dockerfile は `BUILDPLATFORM` と `GOOS` / `GOARCH` にも対応しているため、ローカルからの Go のクロスコンパイルも可能です。

各ビルドはタグを付けずに digest で GHCR へ送信し、その digest をジョブ間で受け渡します。両方が成功した後、公開ジョブがマニフェストを作成して `latest` / バージョン / SHA タグを付け、amd64 と arm64 が含まれることを検証します。片方のビルドが失敗した場合、公開タグは更新しません。公開用イメージへのキャッシュ汚染を避けるため、Actions のキャッシュからビルドレイヤーを復元しません。

公開後はローカルでビルドせずに起動できます。

```sh
docker run --rm --init -p 127.0.0.1:8080:8080 ghcr.io/raydive/zio-learn:latest
```

公開先はワークフロー内の `github.repository` から決まります。認証は自動発行される `GITHUB_TOKEN` を使い、公開ジョブにだけ `packages: write` を付与しています。追加のPAT登録は不要です。初回作成時のパッケージは private のため、誰でも取得できるようにする場合は GitHub の Packages 設定で visibility を public に変更してください。private のまま使う場合は取得前に認証が必要です。[GitHub の Container registry ドキュメント](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)も参照してください。

## 使い方

1. サンプルを選ぶか、コードをエディタに貼り付けます。
2. Scala のバージョンを選びます。`zio` / `zio-streams` / `zio-json` は常に追加されます。
3. **Run** または `⌘ + Enter` / `Ctrl + Enter` で実行します。
4. **Stop** でコンパイラ・実行中の JVM ごと停止できます。

コード・Scala バージョンはブラウザの localStorage に保存します。サンプルを選ぶと現在のコードを置き換えます。

エディタと生成ファイルのプレビューは Scala の構文を色分けします。Prism 1.30.0 を jsDelivr CDN から読み込みます（[配信元](https://www.jsdelivr.com/package/npm/prismjs)）。バージョンを固定し、SRI で取得したファイルを検証します。色分けにはインターネット接続が必要で、読み込めない場合は通常のテキスト表示になります。入力・選択・コピー・スクロールは通常のテキストエリアとして操作できます。

Scala は `3.3.7` / `3.9.0` / `3.10.0` / `2.13.18` から選べます。初期値は `3.3.7` です。選択は実行・生成ファイル・ダウンロードに反映され、Docker ビルド時には全バージョンで依存をキャッシュします。対応範囲は [Scala CLI の公式ドキュメント](https://scala-cli.virtuslab.org/docs/reference/scala-versions/)も参照してください。

### Application を入力して実行

エディタには `object Main extends ZIOAppDefault` を含む Application を入力します。実行形式の選択はありません。サーバーが依存定義・入力コード・Application の object の `main(Array.empty[String])` の呼び出しを含む `playground.sc` に変換し、`scala-cli run playground.sc --server=false --jvm system` で実行します。

ドキュメントの effect は `Main` の中で定義し、`run` から返してください。値を確認したいときは `Console.printLine` などで表示します。`Unsafe.unsafe` や `Runtime.default.unsafe.run` を自分で呼ぶ必要はありません。[公式の ZIOApp の説明](https://zio.dev/reference/core/zioapp/)も参照してください。

サンプル一覧の **TPromise (ZIOAppDefault)** は、[TPromise の公式ドキュメント](https://zio.dev/reference/stm/tpromise/)の `poll`・`await`・`fail` をこの形式で試せます。[ソース](internal/assets/examples/tpromise.scala)では、サンプルの effect を `run` で順に実行して結果を表示します。

```text
poll = Some(Right(0))
await = 0
fail + poll = Some(Left(failed))
```

入力した Application の定義はそのまま生成ファイルに入ります。`ZIOAppDefault` を継承する object は1つにしてください。名前は `Main`・`MyApp` など自由です。対応する `main(...)` の呼び出しはサーバーが追加するため、エディタへの記述は不要です。標準入力には対応していないため、`readLine` などの入力を求めるコードは変更が必要です。

### 生成するファイル

たとえば Application を入力すると、次の `playground.sc` を生成します。

```scala
//> using scala 3.3.7
//> using dep dev.zio::zio:2.1.27
//> using dep dev.zio::zio-streams:2.1.27
//> using dep dev.zio::zio-json:1.1.0

import zio._

object Main extends ZIOAppDefault {
  def run = Console.printLine("Hello, ZIO!")
}

Main.main(Array.empty[String])
```

画面下部で生成内容を確認し、同じ `playground.sc` をダウンロードできます。依存は [Scala CLI の using directives](https://scala-cli.virtuslab.org/docs/reference/directives/) で定義しています。追加ライブラリはエディタの先頭に `//> using dep ...` を記述できます。

Scala の初期値 (`scalaVersion`)・選択肢 (`scalaVersions`) と常に追加する依存ライブラリは [config.json](internal/assets/config.json) で管理しています。選択肢を追加・変更したらイメージを再ビルドしてください。貼り付けるコードの構文は選んだ Scala バージョンに合わせてください。

## ローカル開発

Go 1.27.2、Scala CLI、JDK が必要です。CI は `go.mod` の Go バージョンを読み込み、Docker のビルド環境も Go 1.27.2 に揃えています。

```sh
make run
```

デフォルトは `127.0.0.1:8080` です。`LISTEN_ADDR` で待ち受けアドレスを指定できます。実行コマンドは `scala-cli` に固定しており、PATH 上に存在することが前提です（Docker イメージに同梱）。画面も Go バイナリに埋め込むため、変更後はサーバーを再起動します。

```sh
make tools  # 固定バージョンの golangci-lint を .bin に導入
make fmt    # gofmt / goimports で整形
make check  # lint・整形チェック・race 検査付きテスト・ビルド
```

`make lint` / `make test` / `make build` で個別にも実行できます。ビルドしたバイナリは `.bin/zio-learn` に出力します。`make run` は `go run ./cmd/zio-learn` を実行します。

golangci-lint は `v2.15.0` に固定しています。`make tools` では公式のインストーラーとバイナリを取得するため、初回のみネットワーク接続と curl が必要です。実行時依存には追加しません。標準の検査に加えて、HTTP レスポンスの閉じ忘れ・エラーの扱い・不要な変換などを検査します。GitHub Actions でも同じ `make check` を実行します。

## ディレクトリ構成

[参考記事](https://zenn.dev/kinopippi/articles/fd6a35a9e42145) と、その出典の [Go 公式ガイド](https://go.dev/doc/modules/layout#server-project) に沿って、起動処理を `cmd`、内部実装を `internal` に置きます。

```text
cmd/zio-learn/main.go        # 環境変数・シグナル・起動時エラーの処理
internal/
  app/                      # 設定の読み込み・各層の組み立て・HTTP の起動と終了
  handler/                  # HTTP 受付・入力検証の呼び出し・JSON / NDJSON 応答
  transpiler/               # 入力検証・起動対象の検出・Scala CLI スクリプトの生成
  playground/               # Scala CLI の実行・停止・出力制限・実行イベントの配信
  assets/
    embed.go                # バイナリに埋め込むリソースの定義
    config.json             # Scala バージョンと固定の依存ライブラリ
    examples/               # Scala のサンプル
    web/                    # HTML / CSS / JavaScript
.golangci.yml               # lint と整形の設定
.github/workflows/ci.yml    # CI のチェック
docker/cache-scala.sh       # Docker ビルド時のコンパイラ・依存の事前取得
Makefile                    # ローカルと CI で共通のコマンド
go.mod
```

HTTP handler が transpiler の検証・変換を呼び出し、その生成結果を playground に渡します。transpiler と playground は HTTP に依存せず、お互いにも依存しません。app が設定とサービスを組み立て、HTTP サーバーと実行サービスの終了を連携させます。

playground は Scala CLI の終了待ちを goroutine で行い、`started`・`output`・`finished` のイベントを読み取り専用 channel で返します。handler がこれを NDJSON に変換します。channel のバッファは16件で、送信待ちは context のキャンセルで解除されます。停止・タイムアウト・出力上限・接続切断ではプロセスグループを終了させ、一時ファイルを削除します。シャットダウン開始後は新規実行を受け付けません。

## 実行範囲

自分で信頼できるコードを試すためのローカル専用ツールです。コードはサーバーと同じコンテナ・同じユーザーで動くため、信頼できないコードの実行やインターネットへの公開には使わないでください。Compose のポートはホストのループバックアドレスに限定しています。ホストのディレクトリや Docker ソケットをマウントする必要はありません。

同時実行はサーバー全体で1件、コンパイル込みで120秒、出力は1 MiB、リクエストは256 KiBまでです。コンテナは非 root ユーザーで起動します。実行ファイルは毎回一時ディレクトリに作成し、実行終了後に削除します。

