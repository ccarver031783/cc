from agent.classifier import Classification

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
        "cloud engineer"
    ]

    def apply_rules(self, message_text: str, initial_classification: Classification) -> Classification:
        text = message_text.lower()

        # 0. System notifications bypass all rules
        if initial_classification == Classification.SYSTEM_NOTIFICATION:
            return Classification.SYSTEM_NOTIFICATION

        # 1. Geography violation overrides everything except system notifications
        if any(loc in text for loc in self.GEO_BAD):
            return Classification.GEO_VIOLATION

        # 2. Role mismatch overrides everything except geo + system notifications
        if not any(role in text for role in self.ROLE_GOOD):
            return Classification.ROLE_MISMATCH

        # 3. Amazon Connect spam stays spam
        if initial_classification == Classification.AMAZON_CONNECT_SPAM:
            return Classification.AMAZON_CONNECT_SPAM

        # 4. Otherwise trust the classifier
        return initial_classification
