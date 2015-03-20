package wannaup.services

import scala.concurrent._
import scala.concurrent.ExecutionContext.Implicits.global
import reactivemongo.bson._
import reactivemongo.api._
import wannaup.db.Mongo
import wannaup.models._

/**
 *
 */
class ThreadService(val mailService: MailService) {
  // import implicit format BSON <--> Thread
  import wannaup.models.Threads._

  /**
   * Manage an incoming email, retrieve id and then search thread linked to email
   * @param email the incoming email
   * @return
   */
  def manage(email: String): Unit = {
    val id = email.replace("-reply@inbound.domain.com", "")
    val message = Message(from = "", to = Some(""), body = "")
    Threads.c.find(BSONDocument("_id" -> BSONObjectID(id))).one[Thread].map {
      case Some(thread) =>
        val updatedThread = thread.copy(messages = thread.messages :+ message)
        Threads.c.save(updatedThread).map { lastError =>
          //TODO: log error in case
        }
      case None => //TODO: log error man!
    }
  }

  /**
   * create a new thread, upon creation postman sends a mail containing the message to the to email address
   * setting the sender as the from mail address and the reply-to field to the email address of the mail node
   * (inbound.yourdomain.com).
   * @param owner of this thread
   * @param message to sent
   */
  def create(owner: User, message: Message): Future[Thread] = {
    val thread = Thread(owner = owner, messages = List(message))
    Threads.c.save(thread).map { lastError =>
      val email = Email(
        subject = "",
        html = message.body,
        text = message.body,
        from = message.from,
        to = message.to.get,
        replyTo = s"${thread.id}-reply@qualcosa.com")
      mailService.send(email)
      thread
    }
  }

  /**
   * return detail of a thread identified with id
   * @param id of the thread
   */
  def get(id: String): Future[Option[Thread]] = {
    Threads.c.find(BSONDocument("_id" -> BSONObjectID(id))).one[Thread]
  }

  /**
   * return all threads owned from a user
   * @param userId the owner of threads
   * @param limit
   * @param skip
   */
  def get(userId: String, limit: Int = 100, skip: Int = 0): Future[List[Thread]] = {
    Threads.c.find(BSONDocument("owner.id" -> userId)).options(QueryOpts(skip)).cursor[Thread].toList(limit)
  }

  /**
   * reply with a new message
   * @param threadId where reply
   * @param msg to reply with
   */
  def reply(threadId: String, message: Message): Future[Option[Thread]] = {
    Threads.c.find(BSONDocument("_id" -> BSONObjectID(threadId))).one[Thread].flatMap {
      case Some(thread) =>
        val newThread = thread.copy(messages = thread.messages :+ message)
        Threads.c.update(BSONDocument("_id" -> BSONObjectID(threadId)), thread, upsert = false, multi = false).map { lastError =>
          val email = Email(
            subject = "",
            html = message.body,
            text = message.body,
            from = message.from,
            to = message.to.get,
            replyTo = s"${thread.id}-reply@qualcosa.com")
          mailService.send(email)
          Some(thread)
        }
      case None => Future.successful(None)
    }
  }

}