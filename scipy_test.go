package matlab

// Tests against real SciPy-generated v5 MAT-files located in testdata/scipy/.
//
// All four files use the big-endian ("IM") byte order and miCOMPRESSED elements
// (zlib-wrapped miMATRIX).  The parser's miCOMPRESSED path is therefore exercised
// for every test in this file.
//
// File origins (MATLAB 7.4 / scipy.io):
//   testdouble_7.4_GLNX86.mat  – 1×9 float64 row vector (multiples of π/4)
//   testcomplex_7.4_GLNX86.mat – 1×9 complex128 row vector (unit-circle points)
//   testmatrix_7.4_GLNX86.mat  – 3×5 matrix stored as uint8 in MATLAB memory
//   inner_outer_tbl_param.mat  – 12-variable scientific dataset (real-world file)

import (
	"math"
	"os"
	"testing"

	"github.com/scigolib/matlab/types"
)

// TestOpen_ScipyDoubleFile verifies reading testdouble_7.4_GLNX86.mat.
//
// The file contains a single variable "testdouble" with nine evenly-spaced
// angles from 0 to 2π (multiples of π/4): [0, π/4, π/2, …, 2π].
func TestOpen_ScipyDoubleFile(t *testing.T) {
	f, err := os.Open("testdata/scipy/testdouble_7.4_GLNX86.mat")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	mat, err := Open(f)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	// --- header ---
	if mat.Version != "5.0" {
		t.Errorf("Version = %q, want %q", mat.Version, "5.0")
	}
	// Big-endian SciPy files use "IM" indicator.
	if mat.Endian != "IM" {
		t.Errorf("Endian = %q, want %q", mat.Endian, "IM")
	}

	// --- variable count ---
	if len(mat.Variables) != 1 {
		t.Fatalf("len(Variables) = %d, want 1", len(mat.Variables))
	}

	v := mat.Variables[0]

	// --- name ---
	if v.Name != "testdouble" {
		t.Errorf("Name = %q, want %q", v.Name, "testdouble")
	}

	// --- type ---
	if v.DataType != types.Double {
		t.Errorf("DataType = %v, want %v", v.DataType, types.Double)
	}

	// --- dimensions: 1×9 row vector ---
	wantDims := []int{1, 9}
	if !dimsEqual(v.Dimensions, wantDims) {
		t.Errorf("Dimensions = %v, want %v", v.Dimensions, wantDims)
	}

	// --- not complex ---
	if v.IsComplex {
		t.Error("IsComplex = true, want false")
	}

	// --- data values: k * π/4 for k = 0..8 ---
	data, ok := v.Data.([]float64)
	if !ok {
		t.Fatalf("Data type = %T, want []float64", v.Data)
	}
	if len(data) != 9 {
		t.Fatalf("len(data) = %d, want 9", len(data))
	}

	wantData := [9]float64{
		0,
		math.Pi / 4,
		math.Pi / 2,
		3 * math.Pi / 4,
		math.Pi,
		5 * math.Pi / 4,
		3 * math.Pi / 2,
		7 * math.Pi / 4,
		2 * math.Pi,
	}
	for i, want := range wantData {
		if math.Abs(data[i]-want) > 1e-15 {
			t.Errorf("data[%d] = %.17g, want %.17g", i, data[i], want)
		}
	}
}

// TestOpen_ScipyComplexFile verifies reading testcomplex_7.4_GLNX86.mat.
//
// The file contains "testcomplex": nine complex128 values that trace the unit
// circle in the complex plane at angles k*π/4 (k = 0..8).  Each element has
// magnitude ≈ 1 (within floating-point precision).
func TestOpen_ScipyComplexFile(t *testing.T) {
	f, err := os.Open("testdata/scipy/testcomplex_7.4_GLNX86.mat")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	mat, err := Open(f)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if len(mat.Variables) != 1 {
		t.Fatalf("len(Variables) = %d, want 1", len(mat.Variables))
	}

	v := mat.Variables[0]

	// --- name ---
	if v.Name != "testcomplex" {
		t.Errorf("Name = %q, want %q", v.Name, "testcomplex")
	}

	// --- type ---
	if v.DataType != types.Double {
		t.Errorf("DataType = %v, want %v", v.DataType, types.Double)
	}

	// --- dimensions: 1×9 row vector ---
	wantDims := []int{1, 9}
	if !dimsEqual(v.Dimensions, wantDims) {
		t.Errorf("Dimensions = %v, want %v", v.Dimensions, wantDims)
	}

	// --- must be marked complex ---
	if !v.IsComplex {
		t.Error("IsComplex = false, want true")
	}

	// --- data must be *types.NumericArray ---
	na, ok := v.Data.(*types.NumericArray)
	if !ok {
		t.Fatalf("Data type = %T, want *types.NumericArray", v.Data)
	}

	realPart, ok := na.Real.([]float64)
	if !ok {
		t.Fatalf("NumericArray.Real type = %T, want []float64", na.Real)
	}
	imagPart, ok := na.Imag.([]float64)
	if !ok {
		t.Fatalf("NumericArray.Imag type = %T, want []float64", na.Imag)
	}

	if len(realPart) != 9 || len(imagPart) != 9 {
		t.Fatalf("len(real)=%d len(imag)=%d, want both 9", len(realPart), len(imagPart))
	}

	// Exact values as produced by SciPy (exp(i*k*π/4) for k=0..8).
	// Tiny floating-point residuals (≈1e-16) at quarter-turn boundaries are expected.
	wantReal := [9]float64{
		1,
		0.7071067811865476,
		6.123233995736766e-17,
		-0.7071067811865475,
		-1,
		-0.7071067811865477,
		-1.8369701987210297e-16,
		0.7071067811865474,
		1,
	}
	wantImag := [9]float64{
		0,
		0.7071067811865475,
		1,
		0.7071067811865476,
		1.2246467991473532e-16,
		-0.7071067811865475,
		-1,
		-0.7071067811865477,
		-2.4492935982947064e-16,
	}
	for i := range wantReal {
		if math.Abs(realPart[i]-wantReal[i]) > 1e-15 {
			t.Errorf("real[%d] = %.17g, want %.17g", i, realPart[i], wantReal[i])
		}
		if math.Abs(imagPart[i]-wantImag[i]) > 1e-15 {
			t.Errorf("imag[%d] = %.17g, want %.17g", i, imagPart[i], wantImag[i])
		}
	}

	// Unit-circle invariant: |z| ≈ 1 for all elements (ignoring floating-point
	// residuals at axis-aligned points).
	for i := range realPart {
		mag := math.Sqrt(realPart[i]*realPart[i] + imagPart[i]*imagPart[i])
		if math.Abs(mag-1.0) > 1e-15 {
			t.Errorf("|z[%d]| = %.17g, want ≈ 1", i, mag)
		}
	}
}

// TestOpen_ScipyMatrixFile verifies reading testmatrix_7.4_GLNX86.mat.
//
// The file contains "testmatrix": a 3×5 matrix.  MATLAB stored the raw bytes
// as uint8 (the in-memory MATLAB storage class), so v.Data is []uint8.
// The 15 elements are laid out column-major:
//
//	col 0: [1, 2, 3]
//	col 1: [2, 0, 0]
//	col 2: [3, 0, 0]
//	col 3: [4, 0, 0]
//	col 4: [5, 0, 0]
func TestOpen_ScipyMatrixFile(t *testing.T) {
	f, err := os.Open("testdata/scipy/testmatrix_7.4_GLNX86.mat")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	mat, err := Open(f)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if len(mat.Variables) != 1 {
		t.Fatalf("len(Variables) = %d, want 1", len(mat.Variables))
	}

	v := mat.Variables[0]

	// --- name ---
	if v.Name != "testmatrix" {
		t.Errorf("Name = %q, want %q", v.Name, "testmatrix")
	}

	// --- type reported by MATLAB header ---
	if v.DataType != types.Double {
		t.Errorf("DataType = %v, want %v", v.DataType, types.Double)
	}

	// --- dimensions: 3×5 ---
	wantDims := []int{3, 5}
	if !dimsEqual(v.Dimensions, wantDims) {
		t.Errorf("Dimensions = %v, want %v", v.Dimensions, wantDims)
	}

	// --- not complex ---
	if v.IsComplex {
		t.Error("IsComplex = true, want false")
	}

	// --- data: 15 bytes column-major ---
	// MATLAB stored integer values ≤ 255 using uint8 storage internally.
	data, ok := v.Data.([]uint8)
	if !ok {
		t.Fatalf("Data type = %T, want []uint8", v.Data)
	}
	if len(data) != 15 {
		t.Fatalf("len(data) = %d, want 15", len(data))
	}

	// Column-major element check: index = col*nrows + row.
	wantMatrix := [3][5]uint8{
		{1, 2, 3, 4, 5}, // row 0
		{2, 0, 0, 0, 0}, // row 1
		{3, 0, 0, 0, 0}, // row 2
	}
	nrows := 3
	for row := 0; row < 3; row++ {
		for col := 0; col < 5; col++ {
			got := data[col*nrows+row]
			want := wantMatrix[row][col]
			if got != want {
				t.Errorf("matrix[%d,%d] = %d, want %d", row, col, got, want)
			}
		}
	}
}

// TestOpen_ScipyMultiVariable verifies reading inner_outer_tbl_param.mat.
//
// This is a real-world scientific dataset with 12 variables of mixed dimensions
// and one uint16-backed scalar ("fs").  The test validates the variable count,
// names, types, and spot-checks scalar and array boundary values via sub-tests.
func TestOpen_ScipyMultiVariable(t *testing.T) {
	f, err := os.Open("testdata/scipy/inner_outer_tbl_param.mat")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	mat, err := Open(f)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	// --- header ---
	if mat.Version != "5.0" {
		t.Errorf("Version = %q, want %q", mat.Version, "5.0")
	}
	if mat.Endian != "IM" {
		t.Errorf("Endian = %q, want %q", mat.Endian, "IM")
	}
	if len(mat.Variables) != 12 {
		t.Fatalf("len(Variables) = %d, want 12", len(mat.Variables))
	}

	t.Run("variable_metadata", func(t *testing.T) {
		checkInnerOuterMetadata(t, mat)
	})
	t.Run("scalar_float64_values", func(t *testing.T) {
		checkInnerOuterScalars(t, mat)
	})
	t.Run("fs_uint16_scalar", func(t *testing.T) {
		checkInnerOuterFS(t, mat)
	})
	t.Run("column_vectors", func(t *testing.T) {
		checkInnerOuterVectors(t, mat)
	})
}

// checkInnerOuterMetadata verifies that every variable has the expected name,
// DataType, dimensions, and IsComplex flag.
func checkInnerOuterMetadata(t *testing.T, mat *MatFile) {
	t.Helper()

	type varSpec struct {
		name     string
		dataType types.DataType
		dims     []int
	}
	wantSpecs := []varSpec{
		{"nu", types.Double, []int{1, 1}},
		{"utau", types.Double, []int{1, 1}},
		{"delta", types.Double, []int{1, 1}},
		{"Uinf", types.Double, []int{1, 1}},
		{"Retau", types.Double, []int{1, 1}},
		{"fs", types.Double, []int{1, 1}}, // uint16 storage, but Double MATLAB class
		{"z2", types.Double, []int{1, 1}},
		{"z1", types.Double, []int{40, 1}},
		{"Ubar1", types.Double, []int{40, 1}},
		{"Ubar2", types.Double, []int{1, 1}},
		{"Urms1", types.Double, []int{40, 1}},
		{"Urms2", types.Double, []int{1, 1}},
	}

	for i, spec := range wantSpecs {
		v := mat.Variables[i]
		if v.Name != spec.name {
			t.Errorf("[%d] Name = %q, want %q", i, v.Name, spec.name)
		}
		if v.DataType != spec.dataType {
			t.Errorf("[%d] %s DataType = %v, want %v", i, spec.name, v.DataType, spec.dataType)
		}
		if !dimsEqual(v.Dimensions, spec.dims) {
			t.Errorf("[%d] %s Dimensions = %v, want %v", i, spec.name, v.Dimensions, spec.dims)
		}
		if v.IsComplex {
			t.Errorf("[%d] %s IsComplex = true, want false", i, spec.name)
		}
		if !mat.HasVariable(spec.name) {
			t.Errorf("HasVariable(%q) = false", spec.name)
		}
	}
	if mat.HasVariable("nonexistent") {
		t.Error("HasVariable(\"nonexistent\") = true, want false")
	}
}

// checkInnerOuterScalars spot-checks the float64-backed scalar variables.
func checkInnerOuterScalars(t *testing.T, mat *MatFile) {
	t.Helper()

	scalars := []struct {
		name string
		want float64
	}{
		{"nu", 1.5319862046916872e-05},
		{"utau", 0.6256057829320717},
		{"delta", 0.36121557206917465},
		{"Uinf", 19.950336041634756},
		{"Retau", 14750.690970945816},
		{"z2", 0.00010603323126626648},
		{"Ubar2", 3.1145624425855467},
		{"Urms2", 1.221748135269584},
	}
	for _, tc := range scalars {
		v := mat.GetVariable(tc.name)
		if v == nil {
			t.Errorf("GetVariable(%q) = nil", tc.name)
			continue
		}
		data, ok := v.Data.([]float64)
		if !ok {
			t.Errorf("%s: Data = %T, want []float64", tc.name, v.Data)
			continue
		}
		if len(data) != 1 {
			t.Errorf("%s: len(data) = %d, want 1", tc.name, len(data))
			continue
		}
		if math.Abs(data[0]-tc.want) > math.Abs(tc.want)*1e-14 {
			t.Errorf("%s: data[0] = %.17g, want %.17g", tc.name, data[0], tc.want)
		}
	}
}

// checkInnerOuterFS verifies "fs": a scalar stored as uint16 with value 20000.
// MATLAB wrote this variable using uint16 storage but tagged it as Double class,
// so the Go Data field is []uint16, not []float64.
func checkInnerOuterFS(t *testing.T, mat *MatFile) {
	t.Helper()

	v := mat.GetVariable("fs")
	if v == nil {
		t.Fatal("GetVariable(\"fs\") = nil")
	}
	data, ok := v.Data.([]uint16)
	if !ok {
		t.Fatalf("fs Data = %T, want []uint16", v.Data)
	}
	if len(data) != 1 || data[0] != 20000 {
		t.Errorf("fs data = %v, want [20000]", data)
	}
}

// checkInnerOuterVectors verifies the three 40-element column vectors by
// checking their lengths and first/last element values.
func checkInnerOuterVectors(t *testing.T, mat *MatFile) {
	t.Helper()

	vectors := []struct {
		name  string
		first float64
		last  float64
		tol   float64
	}{
		{"z1", 0.00025600000000000004, 0.5249560000000002, 1e-15},
		{"Ubar1", 5.721017219268166, 19.97597724735843, 1e-13},
		{"Urms1", 1.8135102270826953, 0.043377689886724406, 1e-14},
	}
	for _, tc := range vectors {
		v := mat.GetVariable(tc.name)
		if v == nil {
			t.Errorf("GetVariable(%q) = nil", tc.name)
			continue
		}
		data, ok := v.Data.([]float64)
		if !ok {
			t.Errorf("%s: Data = %T, want []float64", tc.name, v.Data)
			continue
		}
		if len(data) != 40 {
			t.Errorf("%s: len = %d, want 40", tc.name, len(data))
			continue
		}
		if math.Abs(data[0]-tc.first) > tc.tol {
			t.Errorf("%s[0] = %.17g, want %.17g", tc.name, data[0], tc.first)
		}
		if math.Abs(data[39]-tc.last) > tc.tol {
			t.Errorf("%s[39] = %.17g, want %.17g", tc.name, data[39], tc.last)
		}
	}
}

// dimsEqual returns true when two dimension slices have identical length and values.
func dimsEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
