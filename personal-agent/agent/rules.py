class RulesEngine:
    """
    Applies hard-coded rules (geography, role fit, recruiter quality)
    to refine or override the classifier's decision.
    """

    GEO_BAD = ["arlington", "reston", "herndon", "alexandria", "virginia"]
    GEO_GOOD = ["anne arundel", "howard", "prince george"]

    ROLE_GOOD = [
        "site reliability engineer",
        "sre",
        "platform engineer",
        "devsecops",
        "cloud engineer"
    ]

    def apply_rules(self, message_text: str, initial_classification):
        text = message_text.lower()

        # Geography violation
        if any(loc in text for loc in self.GEO_BAD):
            return "geo_violation"

        # Role mismatch
        if not any(role in text for role in self.ROLE_GOOD):
            return "role_mismatch"

        # Otherwise trust the classifier for now
        return initial_classification.value
