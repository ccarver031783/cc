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

## 🧪 Tests

Basic unit tests live in `/tests`.

---

## 📄 License

MIT License (or your choice)
