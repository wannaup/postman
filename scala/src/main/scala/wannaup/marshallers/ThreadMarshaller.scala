package wannaup.marshallers

import spray.httpx.PlayJsonSupport

/**
 *
 */
object ThreadMarshaller extends PlayJsonSupport {
  implicit val threadRestFormat = wannaup.formats.ThreadFormats.rest
  implicit val messageRestFormat = wannaup.formats.MessageFormats.rest
}