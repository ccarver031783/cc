import email
import imapclient
import imaplib
import os
import smtplib
from email.header import decode_header
from email.message import EmailMessage

from dotenv import load_dotenv

load_dotenv()

class EmailIngest:
    """
    Gmail IMAP ingestion (read-write).
    Fetches messages from INBOX via IMAP; see fetch_messages for selection options.
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

    @staticmethod
    def _sanitize_header(value):
        """RFC 5322 headers cannot contain raw newlines; Gmail sometimes folds them in."""
        if not value:
            return None
        cleaned = " ".join(str(value).split())
        return cleaned or None

    def fetch_messages(self, unseen_only=True, limit=None):
        """
        Load messages from INBOX.

        When unseen_only is True (default), only UNSEEN messages are considered;
        if limit is set, only the newest `limit` of those (by UID) are fetched.

        When unseen_only is False, the newest `limit` messages in the folder are
        fetched (read + unread). If limit is omitted, defaults to 100 to avoid
        pulling the entire mailbox.
        """
        messages = []

        # Use imapclient for easier message fetching
        with imapclient.IMAPClient(self.server, ssl=True) as client:
            client.login(self.user, self.password)
            client.select_folder("INBOX", readonly=True)

            if unseen_only:
                uids = sorted(client.search(["UNSEEN"]))
                if limit is not None and len(uids) > limit:
                    uids = uids[-limit:]
            else:
                eff_limit = limit if limit is not None else 100
                all_uids = sorted(client.search(["ALL"]))
                uids = all_uids[-eff_limit:] if len(all_uids) > eff_limit else all_uids

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

                message_id = self._sanitize_header(msg.get("Message-ID")) or ""
                references = self._sanitize_header(msg.get("References")) or ""

                messages.append({
                    "uid": uid,          # <-- REQUIRED FOR DELETE / mark read
                    "sender": sender,
                    "subject": subject,
                    "body": body,
                    "message_id": message_id,
                    "references": references,
                })

        return messages

    def mark_as_read(self, uid):
        """Set the \\Seen flag on a message by UID."""
        self.mail.uid("STORE", str(uid), "+FLAGS", r"(\Seen)")

    def send_text_reply(self, to_addr, subject, body, *, in_reply_to=None, references=None):
        """
        Send a plain-text email via Gmail SMTP (same credentials as IMAP).
        """
        msg = EmailMessage()
        msg["Subject"] = self._sanitize_header(subject) or ""
        msg["From"] = self.user
        msg["To"] = to_addr
        msg.set_content(body, subtype="plain", charset="utf-8")
        in_reply_to = self._sanitize_header(in_reply_to)
        references = self._sanitize_header(references)
        if in_reply_to:
            msg["In-Reply-To"] = in_reply_to
        ref_line = (references or "").strip()
        if in_reply_to:
            if ref_line and in_reply_to not in ref_line:
                ref_line = f"{ref_line} {in_reply_to}"
            elif not ref_line:
                ref_line = in_reply_to
        ref_line = self._sanitize_header(ref_line)
        if ref_line:
            msg["References"] = ref_line

        with smtplib.SMTP("smtp.gmail.com", 587, timeout=60) as smtp:
            smtp.starttls()
            smtp.login(self.user, self.password)
            smtp.send_message(msg)

    def archive_message(self, uid):
        """Archive a message without deleting it."""
        self.mail.uid("STORE", str(uid), "+X-GM-LABELS", "(\\Archive)")

    def delete_message(self, uid):
        """Mark a message as deleted and expunge it."""
        self.mail.uid("STORE", str(uid), "+FLAGS", r"(\Deleted)")
        self.mail.expunge()
