# ZIO-learn

ZIO のドキュメントにあるコードを貼り付けて試す、ローカル用の Scala プレイグラウンドです。

## 公開イメージで起動

公開先: [ghcr.io/raydive/zio-learn](https://github.com/raydive/zio-learn/pkgs/container/zio-learn)

```sh
docker run --rm --init -p 127.0.0.1:8080:8080 ghcr.io/raydive/zio-learn:latest
```

[http://localhost:8080](http://localhost:8080) にアクセスしてください。停止は `Ctrl+C` です。

## 使い方

1. サンプルを選ぶか、コードをエディタに貼り付けます。
2. Scala のバージョンを選びます。`zio` / `zio-streams` / `zio-json` はあらかじめ追加されています。
3. **Run** または `⌘ + Enter` / `Ctrl + Enter` で実行します。

入力するコードは、`ZIOAppDefault` を継承する object を1つ定義し、`run` から試したい effect を返す形にしてください。object 名は自由です。

```scala
import zio._

object Main extends ZIOAppDefault {
  def run =
    for {
      _ <- Console.printLine("Hello, ZIO!")
      _ <- Console.printLine("ドキュメントのコードをここで試せます。")
    } yield ()
}
```

コードと Scala バージョンはブラウザに自動保存されます。サンプルを選ぶと現在のコードを置き換えます。

画面下部の「生成される Scala CLI ファイルを見る」から生成内容を確認し、`playground.sc` をダウンロードできます。

## Docker ビルド

リポジトリのルートで実行してください。ビルドにはインターネット接続が必要です。

```sh
docker build -t zio-learn:local .
docker run --rm --init -p 127.0.0.1:8080:8080 zio-learn:local
```

[http://localhost:8080](http://localhost:8080) にアクセスしてください。停止は `Ctrl+C` です。

Compose を使う場合:

```sh
docker compose up --build
```

停止は `Ctrl+C`、コンテナの削除は `docker compose down` です。
