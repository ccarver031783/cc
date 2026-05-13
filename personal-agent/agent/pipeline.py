from agent.classifier import MessageClassifier, Classification
from agent.rules import RulesEngine
from agent.actions import ActionExecutor
from agent.memory import MemoryManager
from agent.email_ingest import EmailIngest

class AgentPipeline:
    """
    Orchestrates the full flow:
    ingest → classify → apply rules → execute action
    """

    def __init__(self):
        self.classifier = MessageClassifier()
        self.rules = RulesEngine()
        self.actions = ActionExecutor()
        self.memory = MemoryManager()
        self.ingest = EmailIngest()

    def process_message(self, message_text: str):
        initial = self.classifier.classify(message_text)
        final = self.rules.apply_rules(message_text, initial)
        self.memory.record(message_text, final)
        return final, self._execute(final, message_text)

    def _execute(self, classification, message_text):
        if classification == Classification.LEGIT_RECRUITER:
            return self.actions.handle_legit_recruiter(message_text)
        elif classification == Classification.LOW_LEVEL_RECRUITER:
            return self.actions.handle_low_level_recruiter(message_text)
        elif classification == Classification.GEO_VIOLATION:
            return self.actions.handle_geo_violation(message_text)
        elif classification == Classification.ROLE_MISMATCH:
            return self.actions.handle_role_mismatch(message_text)
        elif classification == Classification.HIGH_QUALITY_MATCH:
            return self.actions.handle_high_quality_match(message_text)
        elif classification == Classification.AMAZON_CONNECT_SPAM:
            return self.actions.handle_amazon_connect_decline(message_text)
        elif classification == Classification.SYSTEM_NOTIFICATION:
            return self.actions.handle_system_notification(message_text)
        else:
            return None

    def process_email(self, email, apply_mailbox_side_effects=True):
        message_text = f"{email.get('subject', '')}\n\n{email.get('body', '')}"
        classification, response = self.process_message(message_text)

        uid = email.get("uid")
        if apply_mailbox_side_effects and uid is not None:
            if classification == Classification.AMAZON_CONNECT_SPAM:
                self.ingest.delete_message(uid)
            if classification == Classification.SYSTEM_NOTIFICATION:
                self.ingest.archive_message(uid)

        return classification, response
