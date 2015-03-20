package testdata

import scala.concurrent._
import scala.concurrent.duration._
import org.specs2.mutable._
import org.specs2.mutable._
import org.specs2.specification.BeforeEach
import spray.testkit._

import wannaup.db.Database

trait GeneralSettings extends BeforeAfter {
  this: Specs2RouteTest =>

  implicit val routeTestTimeout = RouteTestTimeout(5.seconds)

  val db = wannaup.db.Database.db
  def before = {
    val a = Await.result(db.drop(), 10.seconds)
  }
  def after = {
    val a = Await.result(db.drop(), 10.seconds)
    system.shutdown()
    //    db.connection.actorSystem.shutdown()
  }
}

object DropDatabaseBefore extends BeforeEach {
  import scala.concurrent.ExecutionContext.Implicits.global
  def before = {
    val a = Await.result(Database.db.drop(), 10.seconds)
  }
}