import threading

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from app.runner import get_entropy_bits, get_entropy_delta

app = FastAPI(title="Life 3.0 Quantum Entropy Service", version="0.1.0")


@app.on_event("startup")
async def _start_grpc() -> None:
    from app.grpc_server import serve

    t = threading.Thread(target=serve, daemon=True, name="grpc-entropy")
    t.start()


@app.get("/health")
async def health():
    return {"status": "ok", "service": "quantum-sim"}


class EntropyResponse(BaseModel):
    bits: list[bool]
    delta: float
    num_bits: int


@app.get("/entropy/{num_bits}", response_model=EntropyResponse)
async def get_entropy(num_bits: int = 8):
    clamped = max(1, min(num_bits, 1024))
    bits = get_entropy_bits(clamped)
    delta = get_entropy_delta()
    return EntropyResponse(bits=bits, delta=delta, num_bits=clamped)


@app.get("/entropy/delta")
async def get_delta():
    return {"delta": get_entropy_delta(16)}
