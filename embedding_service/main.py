"""
Tiny embedding microservice. Wraps the SAME model used at ingest time
(ingest/vectors/load.py) so query-time embeddings are comparable to the
vectors already stored in Postgres -- using a different model here would
silently make similarity search meaningless, even though nothing would
error.

Run with: uvicorn main:app --host 0.0.0.0 --port 8001
"""

from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer

MODEL_NAME = "all-MiniLM-L6-v2"  # must match ingest/vectors/load.py exactly

app = FastAPI()
model = SentenceTransformer(MODEL_NAME)


class EmbedRequest(BaseModel):
    text: str


class EmbedResponse(BaseModel):
    embedding: list[float]


@app.post("/embed", response_model=EmbedResponse)
def embed(req: EmbedRequest) -> EmbedResponse:
    vector = model.encode(req.text).tolist()
    return EmbedResponse(embedding=vector)


@app.get("/health")
def health() -> dict:
    return {"status": "ok", "model": MODEL_NAME}