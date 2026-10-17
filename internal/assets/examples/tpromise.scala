import zio._
import zio.stm._

// Based on https://zio.dev/reference/stm/tpromise/
object MyApp extends ZIOAppDefault {
  val tPromiseOptionValue: UIO[Option[Either[String, Int]]] = for {
    tPromise <- TPromise.make[String, Int].commit
    _        <- tPromise.succeed(0).commit
    res      <- tPromise.poll.commit
  } yield res

  val tPromiseValue: IO[String, Int] = for {
    tPromise <- TPromise.make[String, Int].commit
    _        <- tPromise.succeed(0).commit
    res      <- tPromise.await.commit
  } yield res

  val tPromiseFail: UIO[TPromise[String, Int]] = for {
    tPromise <- TPromise.make[String, Int].commit
    _        <- tPromise.fail("failed").commit
  } yield tPromise

  def run = for {
    pollResult    <- tPromiseOptionValue
    awaitResult   <- tPromiseValue
    failedPromise <- tPromiseFail
    failedResult  <- failedPromise.poll.commit
    _             <- Console.printLine(s"poll = $pollResult")
    _             <- Console.printLine(s"await = $awaitResult")
    _             <- Console.printLine(s"fail + poll = $failedResult")
  } yield ()
}
