use rand::Rng;
use rand::SeedableRng;
use rand_chacha::ChaCha8Rng;

use crate::clock::ClockState;
use crate::clock::now_unix;
use crate::config::AppConfig;

/// Base severity per scenario index (0 unused, 1-12 = Tegmark's scenarios).
const BASE_SEVERITY: [f32; 13] = [
    0.00, // 0: unused
    0.10, // 1: Libertarian Utopia
    0.20, // 2: Benevolent Dictator
    0.15, // 3: Egalitarian Utopia
    0.40, // 4: Gatekeeper
    0.50, // 5: Protector God
    0.55, // 6: Enslaved God
    0.70, // 7: Conquerors
    0.60, // 8: Descendants
    0.65, // 9: Zookeeper
    0.75, // 10: 1984
    0.45, // 11: Revert
    1.00, // 12: Self-Destruction
];

/// Run Monte Carlo simulation to compute a probabilistic ClockState.
///
/// `llm_weights` — per-scenario LLM confidence values (length 13, index = ScenarioID).
/// `entropy_delta` — quantum entropy value from the quantum-sim service ([-1.0, 1.0]).
pub fn simulate(llm_weights: &[f32], entropy_delta: f64, cfg: &AppConfig) -> ClockState {
    let iterations = cfg.weights_cfg.clock.monte_carlo_iterations as usize;
    let base = cfg.weights_cfg.clock.base_minutes_to_midnight;
    let floor = cfg.weights_cfg.clock.floor_minutes;

    // Seed deterministically from the quantum entropy value
    let seed = (entropy_delta.abs() * u64::MAX as f64) as u64;
    let mut rng = ChaCha8Rng::seed_from_u64(seed);

    // Effective weight = LLM confidence × base severity
    let effective: Vec<f32> = (0..13usize)
        .map(|i| {
            let conf = if i < llm_weights.len() { llm_weights[i] } else { 0.0 };
            (conf * BASE_SEVERITY[i]).clamp(0.0, 1.0)
        })
        .collect();

    let total_threat: f32 = effective.iter().sum::<f32>().min(0.95);
    let noise_scale = (entropy_delta.abs() as f32 * 0.1).max(0.01);

    // Collect samples
    let mut samples: Vec<f32> = Vec::with_capacity(iterations);
    for _ in 0..iterations {
        let perturbed: f32 = (1..13)
            .map(|i| {
                let noise: f32 = rng.gen_range(-noise_scale..noise_scale);
                (effective[i] + noise).clamp(0.0, 1.0)
            })
            .sum::<f32>()
            .min(0.95);

        let minutes = (base * (1.0 - perturbed)).max(floor);
        samples.push(minutes);
    }

    let mean = samples.iter().sum::<f32>() / samples.len() as f32;
    let variance =
        samples.iter().map(|s| (s - mean).powi(2)).sum::<f32>() / samples.len() as f32;
    let sigma = variance.sqrt();

    // Dominant scenario: highest effective weight (skip index 0)
    let dominant = effective
        .iter()
        .enumerate()
        .skip(1)
        .max_by(|a, b| a.1.partial_cmp(b.1).unwrap_or(std::cmp::Ordering::Equal))
        .map(|(i, _)| i as u8)
        .unwrap_or(0);

    let dominant_confidence = if dominant > 0 {
        effective[dominant as usize]
    } else {
        0.0
    };

    ClockState {
        minutes_to_midnight: mean.max(floor),
        sigma,
        dominant_scenario: dominant,
        scenario_confidence: dominant_confidence,
        scenario_weights: effective,
        computed_at: now_unix(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn cfg() -> AppConfig {
        AppConfig::for_test().expect("test config")
    }

    #[test]
    fn valid_range() {
        let state = simulate(&vec![0.0f32; 13], 0.5, &cfg());
        assert!(state.minutes_to_midnight > 0.0);
        assert!(state.minutes_to_midnight <= 60.0);
    }

    #[test]
    fn high_threat_reduces_minutes() {
        let mut high = vec![0.0f32; 13];
        high[12] = 0.9; // self-destruction confidence
        let low = vec![0.0f32; 13];

        let c = cfg();
        let dangerous = simulate(&high, 0.0, &c);
        let safe = simulate(&low, 0.0, &c);

        assert!(
            dangerous.minutes_to_midnight < safe.minutes_to_midnight,
            "high threat {} should be < safe {}",
            dangerous.minutes_to_midnight,
            safe.minutes_to_midnight
        );
    }

    #[test]
    fn floor_respected() {
        let max = vec![1.0f32; 13];
        let state = simulate(&max, 1.0, &cfg());
        assert!(
            state.minutes_to_midnight >= 0.5,
            "floor violated: {}",
            state.minutes_to_midnight
        );
    }

    #[test]
    fn sigma_non_negative() {
        let state = simulate(&vec![0.1f32; 13], 0.3, &cfg());
        assert!(state.sigma >= 0.0);
    }

    #[test]
    fn dominant_scenario_in_range() {
        let mut w = vec![0.0f32; 13];
        w[7] = 0.8; // conquerors
        let state = simulate(&w, 0.5, &cfg());
        assert_eq!(state.dominant_scenario, 7);
    }

    #[test]
    fn weights_len_13() {
        let state = simulate(&vec![0.2f32; 13], 0.1, &cfg());
        assert_eq!(state.scenario_weights.len(), 13);
    }
}
