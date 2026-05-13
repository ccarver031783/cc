import re

from agent.classifier import Classification

# US states + DC (lowercase for membership checks)
_US_STATES = {
    "al", "ak", "az", "ar", "ca", "co", "ct", "de", "fl", "ga", "hi", "id", "il", "in",
    "ia", "ks", "ky", "la", "me", "md", "ma", "mi", "mn", "ms", "mo", "mt", "ne", "nv",
    "nh", "nj", "nm", "ny", "nc", "nd", "oh", "ok", "or", "pa", "ri", "sc", "sd", "tn",
    "tx", "ut", "vt", "va", "wa", "wv", "wi", "wy", "dc",
}

# Job posts usually use "City, ST" with both letters capitalized (avoids ", in the" false positives).
_STATE_AFTER_COMMA = re.compile(r",\s*([A-Z]{2})\b")

# Spelled-out state after a comma, e.g. "{city}, Ohio" (broken templates still imply geography).
_NAMED_STATE_AFTER_COMMA = re.compile(
    r",\s*(?:alabama|alaska|arizona|arkansas|california|colorado|connecticut|delaware|"
    r"florida|georgia|hawaii|idaho|illinois|indiana|iowa|kansas|kentucky|louisiana|maine|"
    r"maryland|massachusetts|michigan|minnesota|mississippi|missouri|montana|nebraska|nevada|"
    r"new\s+hampshire|new\s+jersey|new\s+mexico|new\s+york|north\s+carolina|north\s+dakota|"
    r"ohio|oklahoma|oregon|pennsylvania|rhode\s+island|south\s+carolina|south\s+dakota|"
    r"tennessee|texas|utah|vermont|virginia|washington|west\s+virginia|wisconsin|wyoming|"
    r"district\s+of\s+columbia)\b",
    re.IGNORECASE,
)

# "Columbus OH" / "Wilmington DE" (no comma before ST). Min 4-letter "city" to reduce false matches.
_CITY_SPACE_ST = re.compile(r"\b([A-Z][a-z]{3,})\s+([A-Z]{2})\b")

_CITY_WORD_BLACKLIST = frozenset(
    {
        "join", "open", "type", "team", "from", "mail", "june", "july", "page", "click",
        "need", "your", "this", "that", "with", "have", "will", "desk", "thank", "please",
        "transaction", "building", "order", "reference", "follow", "unsubscribe", "account",
    }
)

# First N chars: job location is usually early; long footers add "City, ST" mailing addresses.
_LOCATION_SCAN_CHARS = 8000


class RulesEngine:
    """
    Applies hard-coded rules (geography, role fit, recruiter quality)
    to refine or override the classifier's decision.
    """

    GEO_BAD = ["arlington", "reston", "herndon", "alexandria", "virginia"]
    ROLE_GOOD = [
        "site reliability engineer",
        "sre",
        "platform engineer",
        "devsecops",
        "cloud engineer",
        "devops",
        "dev ops",
    ]

    ACCEPTABLE_MD_HYBRID_HINTS = [
        "anne arundel",
        "howard county",
        "howard co",
        "prince george",
        "prince george's",
        "pg county",
        "pg counties",
    ]

    # Matches the word "remote" anywhere (covers "Fully Remote", "100% Remote", "Remote -", etc.).
    _REMOTE_WORD = re.compile(r"\bremote\b", re.IGNORECASE)
    _REMOTE_NEGATED = re.compile(
        r"\bnot\s+remote\b|\bno\s+remote\b|\bnon[- ]?remote\b|\bwithout\s+remote\b",
        re.IGNORECASE,
    )

    _REMOTE_ONLY = re.compile(
        r"\b(100\s*%|100%)\s*remote\b|\bfully\s*remote\b|\bcompletely\s*remote\b|"
        r"\bremote\s*[-–]?\s*only\b|\b100\s*%\s*remote\b|\ball\s*remote\b",
        re.IGNORECASE,
    )

    # Local / onsite-only roles are not satisfied by generic "remote" wording.
    _LOCAL_OR_ONSITE_ONLY = re.compile(
        r"\b(need\s+only\s+local|only\s+local\b|local\s+and\s+w2|w2\s+candidates\s+only|"
        r"onsite\s+only|on-?site\s+only|must\s+be\s+local|local\s+candidates\s+only|"
        r"\d+\s*days?\s+onsite)\b",
        re.IGNORECASE,
    )

    def apply_rules(self, message_text: str, initial_classification: Classification) -> Classification:
        text = message_text.lower()

        # 0. System notifications bypass all rules
        if initial_classification == Classification.SYSTEM_NOTIFICATION:
            return Classification.SYSTEM_NOTIFICATION

        # 1. Geography violation overrides everything except system notifications
        if any(loc in text for loc in self.GEO_BAD):
            return Classification.GEO_VIOLATION

        # 1b. Explicit US job site (City, ST) → geography unless remote / acceptable MD corridor
        if self._location_outside_acceptable_corridor(message_text, text):
            return Classification.GEO_VIOLATION

        # 2. Role mismatch overrides everything except geo + system notifications
        if not any(role in text for role in self.ROLE_GOOD):
            return Classification.ROLE_MISMATCH

        # 3. Amazon Connect spam stays spam
        if initial_classification == Classification.AMAZON_CONNECT_SPAM:
            return Classification.AMAZON_CONNECT_SPAM

        # 4. Otherwise trust the classifier
        return initial_classification

    def _location_outside_acceptable_corridor(self, raw_text: str, lower_text: str) -> bool:
        scan_raw = raw_text[:_LOCATION_SCAN_CHARS]
        scan_lower = lower_text[:_LOCATION_SCAN_CHARS]

        if any(h in lower_text for h in self.ACCEPTABLE_MD_HYBRID_HINTS):
            return False

        if self._LOCAL_OR_ONSITE_ONLY.search(lower_text):
            return True

        if self._REMOTE_WORD.search(raw_text) and not self._REMOTE_NEGATED.search(lower_text):
            return False
        if self._REMOTE_ONLY.search(raw_text) or self._REMOTE_ONLY.search(lower_text):
            return False

        if _NAMED_STATE_AFTER_COMMA.search(scan_lower):
            return True

        for m in _STATE_AFTER_COMMA.finditer(scan_raw):
            code = m.group(1).lower()
            if code in _US_STATES:
                return True

        for m in _CITY_SPACE_ST.finditer(scan_raw):
            city, code = m.group(1), m.group(2)
            if city.lower() in _CITY_WORD_BLACKLIST:
                continue
            if code.lower() in _US_STATES:
                return True

        return False
