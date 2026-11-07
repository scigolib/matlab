// Package main - Verification script for v7.3 read/write
//
// This script verifies that:
// 1. Writer can create valid v7.3 files
// 2. Reader can parse files created by writer
// 3. Data integrity is preserved (no corruption)
//
// Usage, from the repository root: go run ./scripts/verify-roundtrip
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/scigolib/matlab"
	"github.com/scigolib/matlab/types"
)

func main() {
	fmt.Println("🔴 CRITICAL TEST: v7.3 Write/Read Round-Trip Verification")
	fmt.Println("=========================================================")
	fmt.Println()

	// Test data: simple 1D array of doubles
	testData := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	testVar := &types.Variable{
		Name:       "test_data",
		Dimensions: []int{5},
		DataType:   types.Double,
		Data:       testData,
	}

	// Temporary file for testing
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "test_roundtrip_v73.mat")
	defer os.Remove(testFile) //nolint:errcheck // Cleanup temporary test file

	fmt.Println("📝 Step 1: Write test data to v7.3 file")
	fmt.Printf("   File: %s\n", testFile)
	fmt.Printf("   Data: %v\n\n", testData)

	// Step 1: Write using new writer
	writer, err := matlab.Create(testFile, matlab.Version73)
	if err != nil {
		fmt.Printf("❌ FAILED: Create() error: %v\n", err)
		os.Exit(1)
	}

	err = writer.WriteVariable(testVar)
	if err != nil {
		fmt.Printf("❌ FAILED: WriteVariable() error: %v\n", err)
		os.Exit(1)
	}

	err = writer.Close()
	if err != nil {
		fmt.Printf("❌ FAILED: Close() error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Step 1 PASSED: File written successfully")
	fmt.Println()

	// Step 2: Read back using existing reader
	fmt.Println("📖 Step 2: Read back the written file")

	file, err := os.Open(testFile)
	if err != nil {
		fmt.Printf("❌ FAILED: Cannot open file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close() //nolint:errcheck // Test script, cleanup on exit

	matFile, err := matlab.Open(file)
	if err != nil {
		fmt.Printf("❌ FAILED: Open() error: %v\n", err)
		fmt.Println("\n⚠️  READER BUG: Cannot parse file created by writer!")
		os.Exit(1)
	}

	fmt.Println("✅ Step 2 PASSED: File parsed successfully")
	fmt.Println()

	// Step 3: Verify data integrity
	fmt.Println("🔍 Step 3: Verify data integrity")

	if len(matFile.Variables) != 1 {
		fmt.Printf("❌ FAILED: Expected 1 variable, got %d\n", len(matFile.Variables))
		os.Exit(1)
	}

	readVar := matFile.Variables[0]

	// Check variable name
	if readVar.Name != testVar.Name {
		fmt.Printf("❌ FAILED: Variable name mismatch\n")
		fmt.Printf("   Expected: %s\n", testVar.Name)
		fmt.Printf("   Got: %s\n", readVar.Name)
		os.Exit(1)
	}

	// Check data type
	if readVar.DataType != testVar.DataType {
		fmt.Printf("❌ FAILED: Data type mismatch\n")
		fmt.Printf("   Expected: %v\n", testVar.DataType)
		fmt.Printf("   Got: %v\n", readVar.DataType)
		os.Exit(1)
	}

	// Check dimensions
	if len(readVar.Dimensions) != len(testVar.Dimensions) {
		fmt.Printf("❌ FAILED: Dimensions length mismatch\n")
		fmt.Printf("   Expected: %v\n", testVar.Dimensions)
		fmt.Printf("   Got: %v\n", readVar.Dimensions)
		os.Exit(1)
	}

	for i := range testVar.Dimensions {
		if readVar.Dimensions[i] != testVar.Dimensions[i] {
			fmt.Printf("❌ FAILED: Dimension[%d] mismatch\n", i)
			fmt.Printf("   Expected: %d\n", testVar.Dimensions[i])
			fmt.Printf("   Got: %d\n", readVar.Dimensions[i])
			os.Exit(1)
		}
	}

	// Check data values
	readData, ok := readVar.Data.([]float64)
	if !ok {
		fmt.Printf("❌ FAILED: Data type assertion failed\n")
		fmt.Printf("   Expected: []float64\n")
		fmt.Printf("   Got: %T\n", readVar.Data)
		os.Exit(1)
	}

	if len(readData) != len(testData) {
		fmt.Printf("❌ FAILED: Data length mismatch\n")
		fmt.Printf("   Expected: %d\n", len(testData))
		fmt.Printf("   Got: %d\n", len(readData))
		os.Exit(1)
	}

	for i := range testData {
		if readData[i] != testData[i] {
			fmt.Printf("❌ FAILED: Data[%d] mismatch\n", i)
			fmt.Printf("   Expected: %f\n", testData[i])
			fmt.Printf("   Got: %f\n", readData[i])
			os.Exit(1)
		}
	}

	fmt.Println("✅ Step 3 PASSED: Data integrity verified")
	fmt.Println()

	// Summary
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("✅ ALL TESTS PASSED!")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("\n✨ v7.3 Read/Write Round-Trip Works Correctly! ✨")
	fmt.Println("\nVerified:")
	fmt.Println("  ✓ Writer creates valid v7.3 files")
	fmt.Println("  ✓ Reader can parse written files")
	fmt.Println("  ✓ Data integrity preserved")
	fmt.Println("  ✓ Variable metadata preserved")
	fmt.Println("\n🎉 Round-trip verification successful!")
}
