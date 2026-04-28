from pathlib import Path
from qiskit import QuantumCircuit, transpile
from qiskit_aer import AerSimulator

_simulator = AerSimulator()
_CIRCUIT_PATH = Path(__file__).parent / "circuits" / "entropy.qasm"


def _require_positive_bits(num_bits: int) -> None:
    if num_bits < 1:
        raise ValueError("num_bits must be at least 1")


def _load_circuit() -> QuantumCircuit:
    return QuantumCircuit.from_qasm_file(str(_CIRCUIT_PATH))


def get_entropy_bits(num_bits: int) -> list[bool]:
    """Run the quantum circuit to produce num_bits random bits."""
    _require_positive_bits(num_bits)
    shots_needed = max(1, (num_bits + 3) // 4)  # 4 bits per shot

    qc = _load_circuit()
    transpiled = transpile(qc, _simulator)
    job = _simulator.run(transpiled, shots=shots_needed)
    result = job.result()
    counts = result.get_counts()

    bits: list[bool] = []
    for bitstring, freq in sorted(counts.items()):
        for _ in range(freq):
            # bitstring is MSB-first in Qiskit; reverse to get LSB-first
            bits.extend(ch == "1" for ch in reversed(bitstring))

    return bits[:num_bits]


def get_entropy_delta(num_bits: int = 8) -> float:
    """Return a normalized delta value in [-1, 1] derived from quantum bits."""
    _require_positive_bits(num_bits)
    bits = get_entropy_bits(num_bits)
    value = sum(int(b) * (2**i) for i, b in enumerate(bits))
    max_val = (2**num_bits) - 1
    return (value / max_val) * 2.0 - 1.0  # normalize to [-1, 1]
