# Performance Benchmark: Legacy vs. Next-Gen Math Engine

A side-by-side architectural benchmark executing identical slot game math loops (10,000,000 simulated spins) on identical hardware. By eliminating dynamic array allocations, reducing runtime state evaluation overhead, and optimizing the execution path, throughput increased **65x**.

```text
Legacy Engine:     4,641 spins/sec  ████░░░░░░░░░░░░░░░░
Next-Gen Engine: 304,853 spins/sec  ████████████████████
                                    🚀 65.6x Performance Gain
```

<p float="left">
  <img src="https://raw.githubusercontent.com/BonBonSlick/php_game_engine_guide/refs/heads/main/legacy_3x3.png" width="45%" />
  <img src="https://raw.githubusercontent.com/BonBonSlick/php_game_engine_guide/refs/heads/main/new_3x3.png" width="45%" />
</p>
