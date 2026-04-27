import email
from email.header import decode_header
import imapclient
import imaplib
import os
from dotenv import load_dotenv

load_dotenv()

class EmailIngest:
    """
    Gmail IMAP ingestion (read-write).
    Fetches unread emails and returns normalized message objects.
    """

    def __init__(self):
        self.user = os.getenv("GMAIL_USER")
        self.password = os.getenv("GMAIL_APP_PASSWORD")
        self.server = "imap.gmail.com"

        # Read-write IMAP connection for deleting messages
        self.mail = imaplib.IMAP4_SSL(self.server)
        self.mail.login(self.user, self.password)
        self.mail.select("INBOX", readonly=False)

    def _decode(self, value):
        if not value:
            return ""
        decoded, charset = decode_header(value)[0]
        if isinstance(decoded, bytes):
            return decoded.decode(charset or "utf-8", errors="ignore")
        return decoded

    def fetch_messages(self):
        messages = []

        # Use imapclient for easier message fetching
        with imapclient.IMAPClient(self.server, ssl=True) as client:
            client.login(self.user, self.password)
            client.select_folder("INBOX", readonly=True)

            # Fetch unread messages
            uids = client.search(["UNSEEN"])

            for uid in uids:
                raw = client.fetch(uid, ["RFC822"])[uid][b"RFC822"]
                msg = email.message_from_bytes(raw)

                sender = self._decode(msg.get("From"))
                subject = self._decode(msg.get("Subject"))

                # Extract body (text/plain preferred)
                body = ""
                if msg.is_multipart():
                    for part in msg.walk():
                        if part.get_content_type() == "text/plain":
                            body = part.get_payload(decode=True).decode("utf-8", errors="ignore")
                            break
                else:
                    body = msg.get_payload(decode=True).decode("utf-8", errors="ignore")

                messages.append({
                    "uid": uid,          # <-- REQUIRED FOR DELETE
                    "sender": sender,
                    "subject": subject,
                    "body": body
                })

        return messages

    def archive_message(self, uid):
        """Archive a message without deleting it."""
        self.mail.uid("STORE", str(uid), "+X-GM-LABELS", "(\\Archive)")

    def delete_message(self, uid):
        """Mark a message as deleted and expunge it."""
        self.mail.uid("STORE", str(uid), "+FLAGS", r"(\Deleted)")
        self.mail.expunge()
