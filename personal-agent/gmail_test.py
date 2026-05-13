import argparse
import re
import smtplib
from collections import Counter
from email.utils import parseaddr

from agent.email_ingest import EmailIngest
from agent.pipeline import AgentPipeline

# From-addresses that match recruiter subject heuristics but are not direct recruiter outreach.
RECRUITER_SKIP_SENDERS = frozenset(
    {
        "noreply@glassdoor.com",
        "noreply@iacconf.com",
        "newsletters-noreply@linkedin.com",
        "crew@morningbrew.com",
        "notifications@calendly.com",
        "support@email.fractionalpowerhouse.com",
    }
)

# Skip entire domain (newsletters, schedulers, non-staffing senders).
RECRUITER_SKIP_DOMAINS = frozenset(
    {
        "morningbrew.com",
        "iacconf.com",
        "calendly.com",
        "fractionalpowerhouse.com",
        "eh-solve.com",
        "luma-mail.com",
    }
)

RECRUITER_PATTERNS = [
    r"urgent requirement",
    r"immediate requirement",
    r"hot requirement",
    r"requirement \|",
    r"job (opportunity|opening|role)",
    r"\bposition\s*[:]",
    r"\brole\s*[:]",
    r"\blocation\s*[-:]",
    r"\bhiring\s+senior\b",
    r"\bjob\s+senior\b",
    r"\bis\s+shared\s+with\s+you\b",
    r"\bh1b\s+sponsorship\b",
    r"\bcontract\b",
    r"\bfulltime\b",
    r"\bfull\s*time\b",
    r"\bremote[- ]",
    r"\blooking\s+for\b",
    r"we have an opening",
    r"\bclient\s+.{0,40}\blooking\b",
    # Body-heavy recruiter templates (subject alone often misses these)
    r"need\s+only\s+local",
    r"exciting job opportunity",
    r"\bdevops\s+engineer\b",
    r"\baws\s+devops\b",
    r"employment\s+type\s*:\s*contract",
    r"hiring\s+of\s+\w+\s+engineer",
]


def classify_recruiter(email):
    """
    Return a reason string if the message matches recruiter heuristics, else None.
    """
    _, from_addr = parseaddr(email.get("sender") or "")
    if from_addr:
        addr_lower = from_addr.lower()
        if addr_lower in RECRUITER_SKIP_SENDERS:
            return None
        dom = addr_lower.rsplit("@", 1)[-1] if "@" in addr_lower else ""
        if dom in RECRUITER_SKIP_DOMAINS or dom.endswith(".luma-mail.com"):
            return None

    subject = email["subject"].lower()
    sender = email["sender"].lower()

    # LinkedIn digest alerts: not direct recruiter mail for this script.
    if "jobalerts-noreply@linkedin.com" in sender:
        return None

    if "amazon connect" in subject:
        return "amazon_connect"

    if "linkedin" in sender and "job" in subject:
        return "linkedin_job"

    body = (email.get("body") or "")[:20000]
    haystack = f"{email.get('subject') or ''}\n{body}".lower()

    for pattern in RECRUITER_PATTERNS:
        if re.search(pattern, haystack):
            return "subject_pattern"

    return None


def looks_like_recruiter(email):
    return classify_recruiter(email) is not None


def reply_to_address(sender_header):
    _, addr = parseaddr(sender_header)
    if addr:
        return addr
    m = re.search(r"<([^>\s]+)>", sender_header or "")
    return m.group(1).strip() if m else ""


def reply_subject_line(original_subject):
    o = (original_subject or "").strip()
    if o.lower().startswith("re:"):
        return o
    return f"Re: {o}" if o else "Re:"


def maybe_send_and_mark_read(ingest, email, response, interactive):
    """
    If interactive, prompt before SMTP send and/or marking \\Seen.
    When not interactive, do nothing (caller only prints).
    """
    if not interactive:
        return

    uid = email.get("uid")
    if uid is None:
        print("No IMAP UID on this message — cannot mark read or send a threaded reply.\n")
        return

    to_addr = reply_to_address(email["sender"])
    if response:
        print(
            "\nReview the proposed reply above. "
            "It will be sent from your GMAIL_USER address to the message's From address."
        )
        ans = input("Send this reply and mark the message as read? [y/N]: ").strip().lower()
        if ans != "y":
            print("Skipped: no send, message left as-is.\n")
            return
        if not to_addr:
            print("Could not parse a recipient address from From; not sending.\n")
            return
        try:
            ingest.send_text_reply(
                to_addr,
                reply_subject_line(email.get("subject")),
                response,
                in_reply_to=(email.get("message_id") or None) or None,
                references=(email.get("references") or None) or None,
            )
            print("Reply sent.")
        except (OSError, smtplib.SMTPException) as exc:
            print(f"Send failed ({type(exc).__name__}): {exc}\n")
            return
    else:
        ans = input(
            "No reply template for this classification. Mark the message as read anyway? [y/N]: "
        ).strip().lower()
        if ans != "y":
            print("Skipped: message left as-is.\n")
            return

    try:
        ingest.mark_as_read(uid)
        print("Message marked as read.\n")
    except OSError as exc:
        print(f"Mark-as-read failed ({type(exc).__name__}): {exc}\n")


def print_summary(batch_size, recruiter_count, reason_counts):
    if batch_size:
        pct = 100.0 * recruiter_count / batch_size
        line = (
            f"Scanned {batch_size} message(s); "
            f"{recruiter_count} recruiter-like ({pct:.1f}%)."
        )
    else:
        line = "Scanned 0 message(s); 0 recruiter-like."
    print(line)
    if reason_counts:
        parts = ", ".join(f"{k}={v}" for k, v in sorted(reason_counts.items()))
        print(f"Match breakdown: {parts}")


def main():
    parser = argparse.ArgumentParser(description="Fetch Gmail and run recruiter filter / agent.")
    parser.add_argument(
        "--include-read",
        action="store_true",
        help="Include read mail: fetch the newest messages in INBOX (not only UNSEEN).",
    )
    parser.add_argument(
        "--limit",
        type=int,
        default=None,
        metavar="N",
        help="With UNSEEN only: cap to the N newest unread. With --include-read: "
        "fetch the N newest messages (default 100 if omitted).",
    )
    parser.add_argument(
        "--interactive-replies",
        action="store_true",
        help="After each draft reply, prompt before sending via SMTP and marking the message read. "
        "Disables auto-delete (Amazon) and auto-archive (LinkedIn) until you confirm sends yourself.",
    )
    args = parser.parse_args()

    if args.include_read:
        fetch_limit = args.limit if args.limit is not None else 100
        unseen_only = False
    else:
        fetch_limit = args.limit
        unseen_only = True

    ingest = EmailIngest()
    batch = ingest.fetch_messages(unseen_only=unseen_only, limit=fetch_limit)

    reason_counts = Counter()
    recruiter_emails = []
    for email in batch:
        reason = classify_recruiter(email)
        if reason:
            recruiter_emails.append(email)
            reason_counts[reason] += 1

    batch_size = len(batch)
    recruiter_count = len(recruiter_emails)

    print("--- Summary ---")
    print_summary(batch_size, recruiter_count, reason_counts)
    print()

    pipeline = AgentPipeline()
    apply_effects = not args.interactive_replies

    for email in recruiter_emails:
        print("----- EMAIL -----")
        print(f"From: {email['sender']}")
        print(f"Subject: {email['subject']}")
        print("\n--- AGENT RESPONSE ---")
        classification, response = pipeline.process_email(
            email, apply_mailbox_side_effects=apply_effects
        )
        print(f"(classification: {classification.value})")
        print(response if response is not None else "(no reply template)")
        print("\n")
        maybe_send_and_mark_read(ingest, email, response, args.interactive_replies)

    print("--- Summary (end) ---")
    print_summary(batch_size, recruiter_count, reason_counts)


if __name__ == "__main__":
    main()
