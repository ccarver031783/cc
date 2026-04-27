from agent.pipeline import AgentPipeline
from agent.email_ingest import EmailIngest
import re

# --- SINGLE, UNIFIED FILTER ---
def looks_like_recruiter(email):
    subject = email["subject"].lower()
    sender = email["sender"].lower()

    # Amazon Connect spam (special case)
    if "amazon connect" in subject:
        return True

    # LinkedIn job alerts (include, but no reply)
    if "linkedin" in sender and "job" in subject:
        return True

    # Recruiter subject patterns
    RECRUITER_PATTERNS = [
        r"urgent requirement",
        r"immediate requirement",
        r"hot requirement",
        r"requirement \|",
        r"job (opportunity|opening|role)",
        r"position[: ]",
        r"role[: ]",
        r"contract",
        r"fulltime",
        r"full time",
        r"remote[- ]",
        r"looking for",
        r"we have an opening",
        r"client.*looking",
    ]

    for pattern in RECRUITER_PATTERNS:
        if re.search(pattern, subject):
            return True

    return False
# --------------------------------

pipeline = AgentPipeline()
ingest = EmailIngest()

emails = [e for e in ingest.fetch_messages() if looks_like_recruiter(e)]

print(f"Fetched {len(emails)} recruiter-like emails\n")

for email in emails:
    print("----- EMAIL -----")
    print(f"From: {email['sender']}")
    print(f"Subject: {email['subject']}")
    print("\n--- AGENT RESPONSE ---")
    print(pipeline.process_email(email))
    print("\n")
