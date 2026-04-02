class ActionExecutor:
    """
    Executes the correct action based on the final classification.
    For now, these are stubs that print what they *would* do.
    """

    def handle_legit_recruiter(self, message):
        print("[LEGIT] Notify Chris. Leave unread.")

    def handle_low_level_recruiter(self, message):
        print("[LOW LEVEL] Auto-reply with canned decline. Archive.")

    def handle_geo_violation(self, message):
        print("[GEO VIOLATION] Auto-reply with geography boundary message. Archive.")

    def handle_role_mismatch(self, message):
        print("[ROLE MISMATCH] Auto-reply with trajectory mismatch message. Archive.")

    def handle_high_quality_match(self, message):
        print("[HIGH QUALITY] Auto-reply with 'Chris will respond in 24 hours'. Notify Chris.")
