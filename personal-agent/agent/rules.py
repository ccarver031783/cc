# agent/rules.py

from agent.classifier import Classification

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

    def apply_rules(self, message_text: str, initial_classification: Classification) -> str:
        text = message_text.lower()

        # 1. Geography violation overrides everything
        if any(loc in text for loc in self.GEO_BAD):
            return Classification.GEO_VIOLATION.value

        # 2. Role mismatch overrides everything except geo
        if not any(role in text for role in self.ROLE_GOOD):
            return Classification.ROLE_MISMATCH.value

        # 3. Otherwise trust the classifier
        return initial_classification.value
