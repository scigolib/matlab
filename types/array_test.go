package types

import "testing"

func TestNumericArray_Dims(t *testing.T) {
	tests := []struct {
		name string
		dims []int
		want []int
	}{
		{name: "1D", dims: []int{5}, want: []int{5}},
		{name: "2D", dims: []int{3, 4}, want: []int{3, 4}},
		{name: "3D", dims: []int{2, 3, 4}, want: []int{2, 3, 4}},
		{name: "nil dims", dims: nil, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			na := NumericArray{Dimensions: tt.dims}
			got := na.Dims()
			if len(got) != len(tt.want) {
				t.Fatalf("Dims() length = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Dims()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestNumericArray_Size(t *testing.T) {
	tests := []struct {
		name string
		dims []int
		want int
	}{
		{name: "1D", dims: []int{3}, want: 3},
		{name: "2D", dims: []int{2, 3}, want: 6},
		{name: "3D", dims: []int{2, 3, 4}, want: 24},
		{name: "empty dims", dims: []int{}, want: 0},
		{name: "nil dims", dims: nil, want: 0},
		{name: "scalar 1x1", dims: []int{1, 1}, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			na := NumericArray{Dimensions: tt.dims}
			if got := na.Size(); got != tt.want {
				t.Errorf("Size() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNumericArray_ElementType(t *testing.T) {
	tests := []struct {
		name     string
		dataType DataType
	}{
		{name: "double", dataType: Double},
		{name: "single", dataType: Single},
		{name: "int32", dataType: Int32},
		{name: "uint8", dataType: Uint8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			na := NumericArray{Type: tt.dataType}
			if got := na.ElementType(); got != tt.dataType {
				t.Errorf("ElementType() = %v, want %v", got, tt.dataType)
			}
		})
	}
}

func TestCharArray_Dims(t *testing.T) {
	tests := []struct {
		name string
		dims []int
		want []int
	}{
		{name: "1D", dims: []int{5}, want: []int{5}},
		{name: "2D", dims: []int{2, 3}, want: []int{2, 3}},
		{name: "nil dims", dims: nil, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := CharArray{Dimensions: tt.dims}
			got := ca.Dims()
			if len(got) != len(tt.want) {
				t.Fatalf("Dims() length = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Dims()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCharArray_Size(t *testing.T) {
	tests := []struct {
		name string
		dims []int
		want int
	}{
		{name: "1D", dims: []int{5}, want: 5},
		{name: "2D", dims: []int{2, 3}, want: 6},
		{name: "empty dims", dims: []int{}, want: 0},
		{name: "nil dims", dims: nil, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := CharArray{Dimensions: tt.dims}
			if got := ca.Size(); got != tt.want {
				t.Errorf("Size() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCharArray_ElementType(t *testing.T) {
	ca := CharArray{}
	if got := ca.ElementType(); got != Char {
		t.Errorf("ElementType() = %v, want %v", got, Char)
	}
}

func TestNumElements(t *testing.T) {
	tests := []struct {
		name string
		dims []int
		want int
	}{
		{name: "nil", dims: nil, want: 0},
		{name: "empty", dims: []int{}, want: 0},
		{name: "single element", dims: []int{3}, want: 3},
		{name: "2D", dims: []int{2, 3}, want: 6},
		{name: "3D", dims: []int{2, 3, 4}, want: 24},
		{name: "scalar", dims: []int{1}, want: 1},
		{name: "1x1", dims: []int{1, 1}, want: 1},
		{name: "contains zero", dims: []int{2, 0, 3}, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := numElements(tt.dims); got != tt.want {
				t.Errorf("numElements(%v) = %d, want %d", tt.dims, got, tt.want)
			}
		})
	}
}

// Verify interface compliance.
func TestArrayInterface(_ *testing.T) {
	var _ Array = NumericArray{}
	var _ Array = CharArray{}
}
