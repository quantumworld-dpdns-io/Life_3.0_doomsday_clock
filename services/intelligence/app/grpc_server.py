"""
gRPC server for the Intelligence Service.

Proto-generated stubs (scenario_pb2, scenario_pb2_grpc) are generated from
shared/proto/scenario.proto via `make proto`. Until stubs exist the server
registers a no-op fallback so the FastAPI app can still start.
"""
import asyncio
import time

import grpc
from grpc import aio

try:
    from generated import scenario_pb2, scenario_pb2_grpc  # type: ignore
    _STUBS_AVAILABLE = True
except ImportError:
    _STUBS_AVAILABLE = False

from app.config import settings
from app.db import get_pool


async def _fetch_latest_signals(limit: int = 50) -> list[dict]:
    pool = await get_pool()
    rows = await pool.fetch(
        """
        SELECT s.scenario, s.confidence, s.reasoning, a.url, s.created_at
        FROM scenario_signals s
        JOIN raw_articles a ON a.id = s.article_id
        ORDER BY s.created_at DESC
        LIMIT $1
        """,
        limit,
    )
    return [dict(r) for r in rows]


async def _fetch_aggregate(window_days: int = 7) -> list[float]:
    pool = await get_pool()
    rows = await pool.fetch(
        """
        SELECT scenario, AVG(confidence) AS avg_confidence
        FROM scenario_signals
        WHERE created_at > NOW() - ($1 * INTERVAL '1 day')
        GROUP BY scenario
        ORDER BY scenario
        """,
        window_days,
    )
    weights: list[float] = [0.0] * 13
    for r in rows:
        weights[r["scenario"]] = float(r["avg_confidence"])
    return weights


if _STUBS_AVAILABLE:
    class ScenarioSignalServicer(scenario_pb2_grpc.ScenarioSignalServiceServicer):  # type: ignore
        async def GetLatestSignals(self, request, context):
            signals = await _fetch_latest_signals(request.limit or 50)
            response_signals = []
            for s in signals:
                response_signals.append(scenario_pb2.ScenarioSignal(  # type: ignore
                    scenario=s["scenario"],
                    confidence=s["confidence"],
                    source_url=s["url"],
                    timestamp=int(s["created_at"].timestamp()) if s["created_at"] else int(time.time()),
                ))
            return scenario_pb2.GetLatestSignalsResponse(signals=response_signals)  # type: ignore

        async def GetAggregate(self, request, context):
            weights = await _fetch_aggregate(request.window_days or 7)
            return scenario_pb2.AggregateResponse(weights=weights)  # type: ignore

        async def StreamSignals(self, request, context):
            while context.is_active():
                signals = await _fetch_latest_signals(10)
                for s in signals:
                    yield scenario_pb2.ScenarioSignal(  # type: ignore
                        scenario=s["scenario"],
                        confidence=s["confidence"],
                        source_url=s["url"],
                        timestamp=int(s["created_at"].timestamp()) if s["created_at"] else int(time.time()),
                    )
                await asyncio.sleep(30)

    async def serve():
        server = aio.server()
        scenario_pb2_grpc.add_ScenarioSignalServiceServicer_to_server(  # type: ignore
            ScenarioSignalServicer(), server
        )
        server.add_insecure_port(f"[::]:{settings.grpc_port}")
        await server.start()
        print(f"gRPC server started on port {settings.grpc_port}")
        await server.wait_for_termination()

else:
    async def serve():
        print(
            "WARNING: gRPC stubs not found. Run `make proto` to generate them. "
            "gRPC server will not start."
        )
