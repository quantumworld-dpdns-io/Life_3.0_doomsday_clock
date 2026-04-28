import pytest
from fastapi.testclient import TestClient

from app.api import app

client = TestClient(app)


def test_health():
    r = client.get("/health")
    assert r.status_code == 200
    assert r.json()["status"] == "ok"
    assert r.json()["service"] == "quantum-sim"


def test_entropy_endpoint_default():
    r = client.get("/entropy/8")
    assert r.status_code == 200
    data = r.json()
    assert data["num_bits"] == 8
    assert len(data["bits"]) == 8
    assert all(isinstance(b, bool) for b in data["bits"])
    assert -1.0 <= data["delta"] <= 1.0


def test_entropy_endpoint_32_bits():
    r = client.get("/entropy/32")
    assert r.status_code == 200
    data = r.json()
    assert data["num_bits"] == 32
    assert len(data["bits"]) == 32


def test_entropy_clamped_at_1024():
    r = client.get("/entropy/2000")
    assert r.status_code == 200
    data = r.json()
    assert data["num_bits"] == 1024
    assert len(data["bits"]) == 1024


def test_entropy_clamped_minimum():
    r = client.get("/entropy/0")
    assert r.status_code == 200
    data = r.json()
    assert data["num_bits"] == 1
    assert len(data["bits"]) == 1


def test_delta_endpoint():
    r = client.get("/entropy/delta")
    assert r.status_code == 200
    delta = r.json()["delta"]
    assert -1.0 <= delta <= 1.0
