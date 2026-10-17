import zio._

object Main extends ZIOAppDefault {
  def run =
    for {
      _ <- Console.printLine("Hello, ZIO!")
      _ <- Console.printLine("ドキュメントのコードをここで試せます。")
    } yield ()
}
