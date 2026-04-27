from agent.pipeline import AgentPipeline

pipeline = AgentPipeline()

emails = [
    {
        "sender": "dikshasyncrony@gmail.com",
        "subject": "Full time__ AWS Cloud Ops SRE__ New York, NY",
        "body": "Hi, kindly share updated resume for immediate joiner..."
    },
    {
        "sender": "legit.recruiter@company.com",
        "subject": "Senior SRE role - remote",
        "body": "Hi Chris, we have a senior SRE role supporting AWS and Kubernetes..."
    }
]

for email in emails:
    print("----")
    print(pipeline.process_email(email))
