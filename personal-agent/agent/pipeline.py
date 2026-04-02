from agent.classifier import MessageClassifier, Classification
from agent.rules import RulesEngine
from agent.actions import ActionExecutor

class AgentPipeline:
    """
    Orchestrates the full flow:
    ingest → classify → apply rules → execute action
    """

    def __init__(self):
        self.classifier = MessageClassifier()
        self.rules = RulesEngine()
        self.actions = ActionExecutor()

    def process_message(self, message_text: str):
        # Step 1: initial classification
        initial = self.classifier.classify(message_text)

        # Step 2: apply rules
        final = self.rules.apply_rules(message_text, initial)

        # Step 3: execute action
        self._execute(final, message_text)

    def _execute(self, classification, message_text):
        if classification == "legit_recruiter":
            self.actions.handle_legit_recruiter(message_text)
        elif classification == "low_level_recruiter":
            self.actions.handle_low_level_recruiter(message_text)
        elif classification == "geo_violation":
            self.actions.handle_geo_violation(message_text)
        elif classification == "role_mismatch":
            self.actions.handle_role_mismatch(message_text)
        elif classification == "high_quality_match":
            self.actions.handle_high_quality_match(message_text)
        else:
            print("[UNKNOWN] No action taken.")
