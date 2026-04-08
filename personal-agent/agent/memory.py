import json
import os

MEMORY_PATH = os.path.join(os.path.dirname(__file__), "..", "memory", "recruiters.json")

class MemoryManager:
    def __init__(self):
        self._load()

    def _load(self):
        with open(MEMORY_PATH, "r") as f:
            self.data = json.load(f)

    def _save(self):
        with open(MEMORY_PATH, "w") as f:
            json.dump(self.data, f, indent=4)

    def add_to_blacklist(self, recruiter_email):
        if recruiter_email not in self.data["blacklist"]:
            self.data["blacklist"].append(recruiter_email)
            self._save()

    def add_to_whitelist(self, recruiter_email):
        if recruiter_email not in self.data["whitelist"]:
            self.data["whitelist"].append(recruiter_email)
            self._save()

    def is_blacklisted(self, recruiter_email):
        return recruiter_email in self.data["blacklist"]

    def is_whitelisted(self, recruiter_email):
        return recruiter_email in self.data["whitelist"]