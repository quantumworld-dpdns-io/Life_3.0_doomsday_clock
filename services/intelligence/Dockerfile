FROM python:3.12-slim AS base
WORKDIR /app

RUN pip install --no-cache-dir uv

COPY pyproject.toml .
RUN uv pip install --system --no-cache .

COPY app/ ./app/

EXPOSE 8001 50051

CMD ["uvicorn", "app.api:app", "--host", "0.0.0.0", "--port", "8001"]
