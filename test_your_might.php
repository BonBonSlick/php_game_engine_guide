<?php

declare(strict_types=1);

// CHECKLIST
// - xDebug OFF
// - JIT ON
// - PHP LATEST

// Scale factor for integer-based currency computations to eliminate float precision issues
const MULTIPLIER_SCALE    = 100;
const TECHNICAL_SYMBOL_ID = 255;

// ============================================================================
// DATA LOCALITY & PRECOMPUTATIONS BLOCK (Executed once at startup)
// ============================================================================

// Symbol ID to Character mapping for rapid string-based combo key generation
$gameSymbols = [0 => 'S', 1 => 'A', 2 => 'B', 3 => 'C', 4 => 'W', 5 => 'D'];

// Static reels configuration data (strips mapped as nested arrays)
$reelStrips = [
    [4,0,1,3,2,5,5,1,4,0,2,3,5,1,1,0,4,2,3,5,0,4,1,2,3,5,0,4,2,1,3,5,0,2,4,1,3,5,0,2,1,4,3,5,0,1,2,4,3,5,0,1,4,2,3,5,0,4,1,2,3,5,0,4,2,1,3,5,0,2,4,1,3,5,0,2,1,4,3,5,0,1,2,4,3,5,0,1,4,2,3,5,0,4,1,2,3,5,0,4],
    [1,5,2,0,3,4,4,2,1,5,3,0,4,2,2,5,1,3,0,4,5,1,2,3,0,4,5,1,3,2,0,4,5,3,1,2,0,4,5,3,2,1,0,4,5,2,3,1,0,4,5,2,1,3,0,4,5,1,2,3,0,4,5,1,3,2,0,4,5,3,1,2,0,4,5,3,2,1,0,4,5,2,3,1,0,4,5,2,1,3,0,4,5,1,2,3,0,4,5,1],
    [3,2,0,4,5,1,1,0,3,2,5,4,1,0,0,2,3,5,4,1,2,3,0,5,4,1,2,3,5,0,4,1,2,5,3,0,4,1,2,5,0,3,4,1,2,0,5,3,4,1,2,0,3,5,4,1,2,3,0,5,4,1,2,3,5,0,4,1,2,5,3,0,4,1,2,5,0,3,4,1,2,0,5,3,4,1,2,0,3,5,4,1,2,3,0,5,4,1,2,3]
];

$gameFieldReelsHeights = [3, 3, 3];
$totalReels            = count($gameFieldReelsHeights);

// Pre-allocated flat 1D array representing a 3x3 game field matrix (Data Locality approach)
$generatedGameField = [
    TECHNICAL_SYMBOL_ID, TECHNICAL_SYMBOL_ID, TECHNICAL_SYMBOL_ID,
    TECHNICAL_SYMBOL_ID, TECHNICAL_SYMBOL_ID, TECHNICAL_SYMBOL_ID,
    TECHNICAL_SYMBOL_ID, TECHNICAL_SYMBOL_ID, TECHNICAL_SYMBOL_ID,
];

$betCoins       = 1;
$totalScaledWin = 0;

// Precomputed Win Map Lookup Table (LUT). Converts complex Wild logic checks into O(1) isset operations
$possibleCombinationsPayoutsMap = [
    'AAA' => 10,  'WAA' => 10,  'AWA' => 10,  'WWA' => 10,  'AAW' => 10,  'WAW' => 10,  'AWW' => 10,  'WWW' => 30,
    'BBB' => 50,  'WBB' => 50,  'BWB' => 50,  'WWB' => 50,  'BBW' => 50,  'WBW' => 50,  'BWW' => 50,
    'CCC' => 100, 'WCC' => 100, 'CWC' => 100, 'WWC' => 100, 'CCW' => 100, 'WCW' => 100, 'CWW' => 100,
];

// Flat indices mapping horizontal paylines on the 1D matrix layout
$winLines          = [[0, 1, 2], [3, 4, 5], [6, 7, 8]];
$reelStripsLengths = array_map('count', $reelStrips);
$totalIterations   = 1000000;

// High-precision benchmark entry point measurement
$startTime = microtime(true);

// ============================================================================
// PERFORMANCE HOT PATH: MONTE CARLO EMULATION LOOP
// ============================================================================
for ($spinCount = 1; $spinCount <= $totalIterations; $spinCount++) {

    // Step 1: Spin Simulation & Flat Grid Population
    for ($reelID = 0; $reelID < $totalReels; $reelID++) {
        $reelStrip   = $reelStrips[$reelID];
        $stripLength = $gameFieldReelsHeights[$reelID];

        // Inline RNG window selector preventing index out of bounds exception
        $startPos = mt_rand(0, $reelStripsLengths[$reelID] - (1 + $stripLength));

        // Directly writing into pre-allocated memory layout bypassing intermediate temporary arrays
        for ($position = 0; $position < $stripLength; $position++) {
            $flatIndex                           = $position * $totalReels + $reelID;
            $generatedGameField[(int)$flatIndex] = $reelStrip[$startPos + $position];
        }
    }

    // Step 2: Unrolled Payline Evaluation & Payout Lookup
    foreach ($winLines as $positions) {
        $winComboKey = '';
        foreach ($positions as $pos) {
            $winComboKey .= $gameSymbols[$generatedGameField[$pos]];
        }

        // O(1) hash map evaluation for the generated payout configuration key
        if (isset($possibleCombinationsPayoutsMap[$winComboKey])) {
            $totalScaledWin += $betCoins * $possibleCombinationsPayoutsMap[$winComboKey];
        }
    }
}

// High-precision benchmark exit point measurement
$endTime = microtime(true);

// ============================================================================
// METRICS & METRIC POST-PROCESSING OUTPUT BLOCK
// ============================================================================
$executionTime       = $endTime - $startTime;
$iterationsPerSecond = $executionTime > 0 ? ($totalIterations / $executionTime) : 0;
$rtp                 = $totalScaledWin / ($totalIterations * $betCoins) / MULTIPLIER_SCALE;

echo 'RTP: ' . $rtp . PHP_EOL;
echo 'Execution Time: ' . round($executionTime, 4) . ' seconds' . PHP_EOL;
echo 'Iterations per second (RPS): ' . number_format($iterationsPerSecond, 0, '.', ' ') . PHP_EOL;
