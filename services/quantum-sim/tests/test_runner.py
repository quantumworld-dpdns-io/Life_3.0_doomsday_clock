import pytest
from app.runner import get_entropy_bits, get_entropy_delta


def test_get_entropy_bits_count():
    bits = get_entropy_bits(8)
    assert len(bits) == 8


def test_get_entropy_bits_are_booleans():
    bits = get_entropy_bits(4)
    assert all(isinstance(b, bool) for b in bits)


def test_get_entropy_bits_not_all_same():
    # P(all same) = 2^-31 with 32 bits — effectively impossible for a working circuit
    bits = get_entropy_bits(32)
    assert not (all(bits) or all(not b for b in bits)), (
        "All bits identical — circuit may be broken"
    )


def test_entropy_delta_in_range():
    delta = get_entropy_delta(8)
    assert -1.0 <= delta <= 1.0


def test_entropy_distribution_uniform():
    """Chi-squared test: bits should be approximately 50/50."""
    from scipy.stats import chisquare

    bits = get_entropy_bits(256)
    ones = sum(bits)
    zeros = len(bits) - ones
    _, p_value = chisquare([ones, zeros])
    assert p_value > 0.001, (
        f"Entropy not uniform: {ones} ones / {zeros} zeros (p={p_value:.4f})"
    )


def test_large_entropy_request():
    bits = get_entropy_bits(100)
    assert len(bits) == 100


def test_entropy_bits_max_size():
    bits = get_entropy_bits(1024)
    assert len(bits) == 1024


def test_entropy_bits_single():
    bits = get_entropy_bits(1)
    assert len(bits) == 1
    assert isinstance(bits[0], bool)
