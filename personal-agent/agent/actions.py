class ActionExecutor:
    """
    Executes the correct action based on the final classification.
    """

    def handle_legit_recruiter(self, message):
        return """
Good day,

Thanks for reaching out. Could you please share the job description and details on the environment, expectations, and compensation?
That will help me determine if this role is a fit.
Thank you, and have a great day!
""".strip()

    def handle_low_level_recruiter(self, message):
        return """
Good day,

Thanks for thinking of me. This particular role is not the right fit for now, but please keep me in mind for senior SRE or platform modernization roles.
Thank you, and have a great day!
""".strip()

    def handle_geo_violation(self, message):
        return """
Good day,

I am currently focused on remote roles or hybrid positions in Anne Arundel, Howard, or Prince George’s Counties. With the location requirement, this one is not a fit at this time.
Thank you, and have a great day!
""".strip()

    def handle_role_mismatch(self, message):
        return """
Good day,

I appreciate you reaching out. This role is not aligned with my background as a senior member of the SRE, DevSecOps, or platform engineering teams. Therefore, I will respectfully decline this opportunity at this time.
Thank you again for considering me, and have a great day!
""".strip()

    def handle_high_quality_match(self, message):
        return """
Good day,

This looks aligned with my background in AWS, Kubernetes, and platform modernization. 
Could you please send over the job description, scope of work, team structure, and compensation range?
I will be happy to take a closer look.
Thank you so much for your time and consideration. Have a great day!
""".strip()

    def handle_amazon_connect_decline(self, message):
        return """Good day,

I appreciate the outreach. I’m not pursuing Amazon Connect or other narrowly scoped AWS service roles. My work is focused on senior‑level SRE, DevSecOps, and platform engineering across broader cloud ecosystems, and I’m not looking to silo into a single AWS product.

Please remove me from any Amazon Connect‑related outreach going forward.

Thank you for your time.
""".strip()

    def handle_system_notification(self, message):
        return None

    ACTIONS = {
        "legit_recruiter": handle_legit_recruiter,
        "low_level_recruiter": handle_low_level_recruiter,
        "geo_violation": handle_geo_violation,
        "role_mismatch": handle_role_mismatch,
        "high_quality_match": handle_high_quality_match,
        "amazon_connect_spam": handle_amazon_connect_decline,
        "system_notification": handle_system_notification,
    }
