"""gRPC server implementing EntropyService from entropy.proto.

Falls back gracefully if proto stubs have not been generated yet
(run 'make proto' to generate them).
"""
import grpc
from concurrent import futures

from app.runner import get_entropy_bits, get_entropy_delta

GRPC_PORT = 50052


def serve() -> None:
    try:
        from app.proto import entropy_pb2, entropy_pb2_grpc  # type: ignore

        class EntropyServicer(entropy_pb2_grpc.EntropyServiceServicer):
            def GetEntropy(self, request, context):
                num_bits = max(1, min(request.num_bits, 1024))
                bits = get_entropy_bits(num_bits)
                delta = get_entropy_delta()
                return entropy_pb2.EntropyResponse(bits=bits, delta=delta)

        server = grpc.server(futures.ThreadPoolExecutor(max_workers=4))
        entropy_pb2_grpc.add_EntropyServiceServicer_to_server(EntropyServicer(), server)
        server.add_insecure_port(f"[::]:{GRPC_PORT}")
        server.start()
        print(f"gRPC EntropyService listening on :{GRPC_PORT}")
        server.wait_for_termination()

    except ImportError:
        print("Proto stubs not generated yet — run 'make proto' first.")
        print(f"gRPC EntropyService would listen on :{GRPC_PORT}")


if __name__ == "__main__":
    serve()
