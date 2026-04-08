from agent.pipeline import AgentPipeline

pipeline = AgentPipeline()

msg = "Hi Chris, I have a senior SRE role in Baltimore..."
print(pipeline.process_message(msg))
