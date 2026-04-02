import os
from enum import Enum
from openai import OpenAI
from dotenv import load_dotenv

load_dotenv()

class Classification(Enum):
    LEGIT_RECRUITER = "legit_recruiter"
    LOW_LEVEL_RECRUITER = "low_level_recruiter"
    GEO_VIOLATION = "geo_violation"
    ROLE_MISMATCH = "role_mismatch"
    HIGH_QUALITY_MATCH = "high_quality_match"


class MessageClassifier:
    """
    Uses an LLM to classify recruiter messages into one of the five categories.
    """

    def __init__(self):
        self.client = OpenAI(api_key=os.getenv("OPENAI_API_KEY"))

    def classify(self, message_text: str) -> Classification:
        prompt = f"""
You are a classifier for recruiter messages. 
Your job is to categorize the message into EXACTLY one of the following:

- legit_recruiter
- low_level_recruiter
- geo_violation
- role_mismatch
- high_quality_match

Definitions:
- legit_recruiter: coherent, real job, real recruiter, no red flags.
- low_level_recruiter: broken English, keyword scraping, vague client, spammy.
- geo_violation: mentions Arlington, Reston, Herndon, Alexandria, or Virginia.
- role_mismatch: role is not SRE/DevSecOps/Platform/Cloud at senior level.
- high_quality_match: senior-level SRE/Platform/DevSecOps role, remote or hybrid in AA/Howard/PG.

Message to classify:
\"\"\"{message_text}\"\"\"

Respond ONLY with the category name.
"""

        try:
            response = self.client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[{"role": "user", "content": prompt}],
                temperature=0
            )

            label = response.choices[0].message.content.strip().lower()

            # Validate output
            for c in Classification:
                if c.value == label:
                    return c

            # Fallback
            return Classification.LOW_LEVEL_RECRUITER

        except Exception as e:
            print(f"[Classifier Error] {e}")
            return Classification.LOW_LEVEL_RECRUITER
