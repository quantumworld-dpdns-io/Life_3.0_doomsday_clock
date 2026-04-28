OPENQASM 2.0;
include "qelib1.inc";

qreg q[4];
creg c[4];

// Apply Hadamard to all qubits: creates equal superposition
// Upon measurement, each qubit collapses to 0 or 1 with equal probability
h q[0];
h q[1];
h q[2];
h q[3];

// Add entanglement for more complex correlations
cx q[0], q[1];
cx q[2], q[3];

// Additional Hadamard to break entanglement symmetry
h q[1];
h q[3];

measure q[0] -> c[0];
measure q[1] -> c[1];
measure q[2] -> c[2];
measure q[3] -> c[3];
