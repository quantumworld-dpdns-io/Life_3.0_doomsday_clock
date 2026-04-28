FROM python:3.12-slim AS base
WORKDIR /app

# Qiskit Aer requires OpenMP runtime
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir uv

COPY pyproject.toml .
RUN uv pip install --system --no-cache .

COPY app/ ./app/

EXPOSE 8002 50052
CMD ["uvicorn", "app.api:app", "--host", "0.0.0.0", "--port", "8002"]
