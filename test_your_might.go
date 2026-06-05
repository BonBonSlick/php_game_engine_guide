package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	BET_MULTIPLIER_SCALE uint8  = 100
	TECHNICAL_SYMBOL_ID  uint8  = math.MaxUint8
	TOTAL_ITERATIONS     uint32 = 100000000 // 100M total iterations for Monte Carlo simulation
)

func main() {
	// Static game configurations (3x3 grid base model)
	reelStrips := [][]uint8{
		{4, 0, 1, 3, 2, 5, 5, 1, 4, 0, 2, 3, 5, 1, 1, 0, 4, 2, 3, 5, 0, 4, 1, 2, 3, 5, 0, 4, 2, 1, 3, 5, 0, 2, 4, 1, 3, 5, 0, 2, 1, 4, 3, 5, 0, 1, 2, 4, 3, 5, 0, 1, 4, 2, 3, 5, 0, 4, 1, 2, 3, 5, 0, 4, 2, 1, 3, 5, 0, 2, 4, 1, 3, 5, 0, 2, 1, 4, 3, 5, 0, 1, 2, 4, 3, 5, 0, 1, 4, 2, 3, 5, 0, 4, 1, 2, 3, 5, 0, 4},
		{1, 5, 2, 0, 3, 4, 4, 2, 1, 5, 3, 0, 4, 2, 2, 5, 1, 3, 0, 4, 5, 1, 2, 3, 0, 4, 5, 1, 3, 2, 0, 4, 5, 3, 1, 2, 0, 4, 5, 3, 2, 1, 0, 4, 5, 2, 3, 1, 0, 4, 5, 2, 1, 3, 0, 4, 5, 1, 2, 3, 0, 4, 5, 1, 3, 2, 0, 4, 5, 3, 1, 2, 0, 4, 5, 3, 2, 1, 0, 4, 5, 2, 3, 1, 0, 4, 5, 2, 1, 3, 0, 4, 5, 1, 2, 3, 0, 4, 5, 1},
		{3, 2, 0, 4, 5, 1, 1, 0, 3, 2, 5, 4, 1, 0, 0, 2, 3, 5, 4, 1, 2, 3, 0, 5, 4, 1, 2, 3, 5, 0, 4, 1, 2, 5, 3, 0, 4, 1, 2, 5, 0, 3, 4, 1, 2, 0, 5, 3, 4, 1, 2, 0, 3, 5, 4, 1, 2, 3, 0, 5, 4, 1, 2, 3, 5, 0, 4, 1, 2, 5, 3, 0, 4, 1, 2, 5, 0, 3, 4, 1, 2, 0, 5, 3, 4, 1, 2, 0, 3, 5, 4, 1, 2, 3, 0, 5, 4, 1, 2, 3},
	}
	winLines := [][]uint8{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}}
	gameFieldReelsHeights := []uint8{3, 3, 3}

	// Utilize available physical threads
	cores := runtime.NumCPU()
	var betCoins uint8 = 1

	// Precomputed and pre-allocated metadata to eliminate runtime overhead
	reelStripsLengths := []uint32{
		uint32(len(reelStrips[0])),
		uint32(len(reelStrips[1])),
		uint32(len(reelStrips[2])),
	}

	payoutConfig := map[string]uint16{
		"AAA": 10, "WAA": 10, "AWA": 10, "WWA": 10, "AAW": 10, "WAW": 10, "AWW": 10, "WWW": 30,
		"BBB": 50, "WBB": 50, "BWB": 50, "WWB": 50, "BBW": 50, "WBW": 50, "BWW": 50,
		"CCC": 100, "WCC": 100, "CWC": 100, "WWC": 100, "CCW": 100, "WCW": 100, "CWW": 100,
	}

	symbolIDsByName := map[byte]uint8{'S': 0, 'A': 1, 'B': 2, 'C': 3, 'W': 4, 'D': 5}
	symbolCount := len(symbolIDsByName) // Total number of unique symbols (e.g., 6)

	// 3D Look-Up Table (LUT) initialized during engine bootstrapping
	PayOutLookUpTable := make([][][]uint16, symbolCount)
	for i := 0; i < symbolCount; i++ {
		PayOutLookUpTable[i] = make([][]uint16, symbolCount)
		for j := 0; j < symbolCount; j++ {
			PayOutLookUpTable[i][j] = make([]uint16, symbolCount)
		}
	}

	for combinationKey, betMultiplier := range payoutConfig {
		res1 := symbolIDsByName[combinationKey[0]]
		res2 := symbolIDsByName[combinationKey[1]]
		res3 := symbolIDsByName[combinationKey[2]]

		PayOutLookUpTable[res1][res2][res3] = betMultiplier
	}
	var totalReels uint8 = uint8(len(gameFieldReelsHeights))

	// Async workloads orchestration
	iterationsPerWorker := TOTAL_ITERATIONS / uint32(cores)
	var wg sync.WaitGroup
	winsChannel := make(chan uint64, cores)

	startTime := time.Now()

	// High-performance non-blocking seed generation
	baseServerSeed := uint64(time.Now().UnixNano())
	baseClientSeed := uint64(123456789)

	for coreNumber := 0; coreNumber < cores; coreNumber++ {
		wg.Add(1)

		go func(coreNumber int) {
			defer wg.Done()

			serverSeed := baseServerSeed + uint64(coreNumber)
			clientSeed := baseClientSeed * uint64(coreNumber+1)

			winsChannel <- RunSimulationWorker(
				serverSeed,
				clientSeed,
				iterationsPerWorker,
				reelStrips,
				gameFieldReelsHeights,
				totalReels,
				reelStripsLengths,
				winLines,
				PayOutLookUpTable,
				betCoins,
			)
		}(coreNumber)
	}

	// Wait for processing lanes to complete, then close data pipeline
	wg.Wait()
	close(winsChannel)

	var totalScaledWin uint64
	for win := range winsChannel {
		totalScaledWin += win
	}

	executionTime := time.Since(startTime)
	iterationsPerSecond := float64(TOTAL_ITERATIONS) / executionTime.Seconds()
	rtp := float64(totalScaledWin) / float64(TOTAL_ITERATIONS*uint32(betCoins)) / float64(BET_MULTIPLIER_SCALE)

	// Output Formatting
	fmt.Printf("RTP: %g\n", rtp)
	fmt.Printf("Execution Time: %.4f seconds\n", executionTime.Seconds())
	fmt.Printf("Iterations per second (RPS): %s\n", formatWithSpaces(int(iterationsPerSecond)))
}

// RunSimulationWorker executes pure mathematical simulations on isolated threads.
// All data structures are optimized for Data Locality to ensure L1/L2 Cache hits.
func RunSimulationWorker(
	serverSeed, clientSeed uint64,
	iterationsPerCore uint32,
	reelStrips [][]uint8,
	gameFieldReelsHeights []uint8,
	totalReels uint8,
	reelStripsLengths []uint32,
	winLines [][]uint8,
	payoutLookUpTable [][][]uint16,
	betCoins uint8,
) uint64 {
	// Local thread-safe pseudo-random number generator (v2 PCG)
	localRNG := rand.New(rand.NewPCG(serverSeed, clientSeed))

	var totalGameFieldCells uint8
	for _, height := range gameFieldReelsHeights {
		totalGameFieldCells += uint8(height)
	}

	var localTotalScaledWin uint64

	// Allocation outside the loop guarantees zero heap allocations on the hot path (CPU Stack bound)
	localGeneratedGameField := make([]uint8, totalGameFieldCells)
	for i := range localGeneratedGameField {
		localGeneratedGameField[i] = TECHNICAL_SYMBOL_ID
	}

	// HOT PATH: Avoids dynamic memory operations; safely overrides the preallocated 1D matrix
	for spinCount := 0; spinCount < int(iterationsPerCore); spinCount++ {

		// Step 1: Spin Simulation & Flat Grid Population (Row-Major Order Mapping)
		for reelID := 0; reelID < int(totalReels); reelID++ {
			reelStripLength := gameFieldReelsHeights[reelID]
			// Fixed boundary math to guarantee access to the very last element of the strip
			maximalPositionOnReelStrip := reelStripsLengths[reelID] - uint32(reelStripLength)

			reelStripStartPosition := localRNG.IntN(int(maximalPositionOnReelStrip) + 1)
			reelStrip := reelStrips[reelID]

			for position := 0; position < int(reelStripLength); position++ {
				gameFieldPosition := position*int(totalReels) + reelID
				localGeneratedGameField[gameFieldPosition] = uint8(reelStrip[reelStripStartPosition+position])
			}
		}

		// Step 2: Unrolled Payline Evaluation via Cache-Friendly Array Lookup
		// Loops use explicit length check allowing the Go compiler to perform all (Bounds Check Elimination)
		for winLineID := 0; winLineID < len(winLines); winLineID++ {
			// Static unrolling assuming 3-symbol evaluations for absolute performance
			firstSymbol := localGeneratedGameField[winLines[winLineID][0]]
			secondSymbol := localGeneratedGameField[winLines[winLineID][1]]
			thirdSymbol := localGeneratedGameField[winLines[winLineID][2]]

			payout := payoutLookUpTable[firstSymbol][secondSymbol][thirdSymbol]
			if payout > 0 {
				localTotalScaledWin += uint64(betCoins) * uint64(payout)
			}
		}
	}

	return localTotalScaledWin
}

// Helper function mimicking PHP's number_format space separator for CLI readability
func formatWithSpaces(number int) string {
	stringNumber := strconv.Itoa(number)
	var response []byte
	length := len(stringNumber)
	for i := 0; i < length; i++ {
		if i > 0 && (length-i)%3 == 0 {
			response = append(response, ' ')
		}
		response = append(response, stringNumber[i])
	}

	return string(response)
}
