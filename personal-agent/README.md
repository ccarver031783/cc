# Personal Recruiter AI Agent

A lightweight, cloud‑friendly AI agent designed to filter recruiter messages, classify them using explicit rules, and automate responses based on geography, role fit, and recruiter quality.

This project is built to run from a phone using GitHub Codespaces and deploy as a tiny container to a free cloud platform (Fly.io, Railway, etc.).

---

## 🎯 Purpose

This agent helps manage inbound recruiter communication from:

- Email (IMAP/API)
- LinkedIn recruiter messages (via email notifications)

It classifies each message into one of five categories:

- **LEGIT_RECRUITER**
- **LOW_LEVEL_RECRUITER**
- **GEO_VIOLATION**
- **ROLE_MISMATCH**
- **HIGH_QUALITY_MATCH**

Each category triggers a specific automated action.

---

## 🧠 Classification Rules

### Geography (hard filters)
Reject automatically if the role is on‑site or hybrid in:

- Arlington  
- Reston  
- Herndon  
- Alexandria  
- Any Virginia location  

Accept if:

- Remote  
- Hybrid in Anne Arundel County  
- Hybrid in Howard County  
- Hybrid in Prince George’s County  

### Role Requirements
Accept only:

- Senior SRE  
- Senior DevSecOps  
- Senior Platform Engineer  
- Senior Cloud Engineer  

Reject:

- Junior roles  
- Unrelated tech stacks  
- Helpdesk/sysadmin  
- Anything outside the desired trajectory  

### Recruiter Requirements
Reject if:

- Recruiter cannot name the client  
- Recruiter cannot specify employment type  
- Message appears to be keyword‑scraping  

---

## 📤 Automated Actions

### LEGIT_RECRUITER
- Notify Chris  
- Leave message unread  
- No auto‑reply  

### LOW_LEVEL_RECRUITER
- Auto‑reply with canned “thanks but no thanks”  
- Archive  

### GEO_VIOLATION
Auto‑reply:

> “Thanks for reaching out. Chris is not pursuing roles in Arlington, Reston, Herndon, Alexandria, or Virginia. If you have remote roles or hybrid roles in Anne Arundel, Howard, or Prince George’s counties, he’d be happy to talk.”

Then archive.

### ROLE_MISMATCH
Auto‑reply:

> “Thanks for reaching out. This role doesn’t align with Chris’s current trajectory, but he appreciates the consideration.”

Then archive.

### HIGH_QUALITY_MATCH
Auto‑reply:

> “Thanks for reaching out — this is Chris’s AI agent. He has been notified of this opportunity and will personally respond within 24 hours.”

Notify Chris immediately.

---

## 📬 Outputs

- **Daily digest** of LEGIT_RECRUITER + HIGH_QUALITY_MATCH  
- **Instant alerts** for HIGH_QUALITY_MATCH  

---

## 🧱 Project Structure

See `/agent`, `/integrations`, and `/config` for modular components.

---

## 🚀 Deployment

This project is designed to run:

- Locally in GitHub Codespaces  
- As a tiny container on Fly.io or Railway (free tier)  

---

## 📦 Requirements

See `requirements.txt` for dependencies.

---

## 📧 Gmail test script (`gmail_test.py`)

Fetch mail from **INBOX** over IMAP, apply recruiter heuristics + the classification pipeline, and print a summary plus draft replies. Configure **`GMAIL_USER`**, **`GMAIL_APP_PASSWORD`**, and **`OPENAI_API_KEY`** in `.env` (same app password is used for SMTP when sending replies).

Run from the repo root:

```bash
cd cc/personal-agent
python3 gmail_test.py --help
```

### Flags

| Flag | Meaning |
|------|--------|
| *(none)* | **Unread only** (`UNSEEN`). No cap: every unread message in INBOX is considered. |
| `--limit N` | **Cap how many messages** are fetched (newest by UID). With unread-only, keeps the **N** newest unread. With `--include-read`, limits to the **N** newest messages in the folder. |
| `--include-read` | Include **read** mail, not just unread: fetches the **newest** messages in INBOX. If you omit `--limit`, defaults to **100** messages so the whole mailbox is never pulled at once. |
| `--interactive-replies` | After each draft, **prompt** before sending a reply via **SMTP** and marking the message **read**. Also **disables** automatic delete (Amazon spam) and archive (LinkedIn-style notifications) for that run until you confirm per message. |

### Examples

```bash
# All current unread in INBOX; print-only (no send, no mark read)
python3 gmail_test.py

# Newest 40 unread only (faster/smaller batch)
python3 gmail_test.py --limit 40

# Newest 150 messages in INBOX (read + unread) — good for a backlog pass
python3 gmail_test.py --include-read --limit 150

# Default include-read window (100 newest messages) if you omit --limit
python3 gmail_test.py --include-read

# Review each reply before sending and marking read (still uses unread-only unless you add --include-read)
python3 gmail_test.py --interactive-replies

# Same, but only the 25 newest unread
python3 gmail_test.py --limit 25 --interactive-replies

# Backlog: recent read+unread, confirm sends interactively
python3 gmail_test.py --include-read --limit 80 --interactive-replies
```

---

## 🧪 Tests

Basic unit tests live in `/tests`.

---

## 📄 License

MIT License (or your choice)
