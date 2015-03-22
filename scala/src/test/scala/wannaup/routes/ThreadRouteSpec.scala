package wannaup.routes

import scala.concurrent._
import scala.concurrent.duration._
import org.specs2.mutable._
import org.specs2.time.NoTimeConversions
import spray.testkit.Specs2RouteTest
import spray.http._
import StatusCodes._
import play.api.libs.json._
import reactivemongo.bson._
import wannaup.services._
import wannaup.models._
import wannaup.marshallers.ThreadMarshaller._
import testdata._

/**
 *
 */
class ThreadRouteSpec extends Specification with Specs2RouteTest
  with NoTimeConversions
  with GeneralSettings {

  import Threads._

  // Create service to test with mock or stub injected
  val mandrillService = new MandrillService(MandrillSettings(key = "mandrill.key"))
  val threadRoute = new ThreadRoute {
    val actorRefFactory = system
    val threadService = new ThreadService(mandrillService)
  }

  // tests for Thread service
  "ThreadRoute" should {

    "create a new thread when POST a message to /threads" in {
      val authHeader = HttpHeaders.`Authorization`(BasicHttpCredentials(UserData.user1.id, "doesn't matter man"))
      Post("/threads", MessageData.msg0) ~> addHeader(authHeader) ~> threadRoute.route ~> check {
        response.status should be(StatusCodes.OK)
        val respThread = responseAs[Thread]
        val dbThread = Await.result(Threads.c.find(BSONDocument("_id" -> BSONObjectID(respThread.id))).one[Thread], 5.seconds)
        responseAs[Thread] must be equalTo (dbThread.get)
      }
    }

    "reply in a thread when POST a new message to /threads/:theardId/reply" in {
      val authHeader = HttpHeaders.`Authorization`(BasicHttpCredentials(ThreadData.thread0.owner.id, "doesn't matter man"))
      Await.result(Threads.c.save(ThreadData.thread0), 5.seconds)
      val threadId = ThreadData.thread0.id
      Post(s"/threads/$threadId/reply", MessageData.msg0) ~> addHeader(authHeader) ~> threadRoute.route ~> check {
        response.status should be(StatusCodes.OK)
        responseAs[Thread] must be equalTo (ThreadData.thread0)
      }
    }

    "return a detail of a thread when GET thread detail to /threads/:threadId" in {
      val authHeader = HttpHeaders.`Authorization`(BasicHttpCredentials(ThreadData.thread0.owner.id, "doesn't matter man"))
      Await.result(Threads.c.save(ThreadData.thread0), 5.seconds)
      val threadsId = ThreadData.thread0.id
      Get(s"/threads/$threadsId") ~> addHeader(authHeader) ~> threadRoute.route ~> check {
        response.status should be(StatusCodes.OK)
        responseAs[Thread] must be equalTo (ThreadData.thread0)
      }
    }

    "return all threads of a user when GET threads to /threads" in DropDatabaseBefore {
      val threadOfUser = ThreadData.threads.filter(_.owner.id == ThreadData.thread0.owner.id)
      val authHeader = HttpHeaders.`Authorization`(BasicHttpCredentials(ThreadData.thread0.owner.id, "doesn't matter man"))
      val futureThreads = ThreadData.threads.map { Threads.c.save(_) }
      Await.result(Future.sequence(futureThreads), 5.seconds)
      Get("/threads") ~> addHeader(authHeader) ~> threadRoute.route ~> check {
        response.status should be(StatusCodes.OK)
        responseAs[List[Thread]].length must be equalTo (ThreadData.threads.length)
        responseAs[List[Thread]].toSet must equalTo(threadOfUser.toSet)
      }
    }

    "return 401 if GET threads to /threads with bad credentials" in {
      Get("/threads") ~> threadRoute.sealRoute(threadRoute.route) ~> check {
        response.status should be(StatusCodes.Unauthorized)
      }
    }
  }

}