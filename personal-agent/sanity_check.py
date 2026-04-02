from openai import OpenAI
from dotenv import load_dotenv
import os

print("Loading environment...")
load_dotenv()

print("Key loaded?", "OPENAI_API_KEY" in os.environ)

client = OpenAI()

print("Calling API...")

resp = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello from my agent!"}]
)

print("Response:", resp.choices[0].message.content)
