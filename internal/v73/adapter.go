package v73

import (
	"fmt"

	"github.com/scigolib/hdf5"
	"github.com/scigolib/matlab/types"
)

const (
	// MATLAB class type identifiers.
	matlabClassDouble = "double"
	matlabClassSingle = "single"
	matlabClassInt8   = "int8"
	matlabClassUint8  = "uint8"
	matlabClassInt16  = "int16"
	matlabClassUint16 = "uint16"
	matlabClassInt32  = "int32"
	matlabClassUint32 = "uint32"
	matlabClassInt64  = "int64"
	matlabClassUint64 = "uint64"
	matlabClassChar   = "char"
	matlabClassStruct = "struct"
	matlabClassCell   = "cell"
)

// HDF5Adapter adapts HDF5 structures to MATLAB types.
type HDF5Adapter struct {
	file *hdf5.File
}

// NewHDF5Adapter creates a new adapter.
func NewHDF5Adapter(file *hdf5.File) *HDF5Adapter {
	return &HDF5Adapter{file: file}
}

// ConvertToMatlab converts HDF5 file to MATLAB variables.
func (a *HDF5Adapter) ConvertToMatlab() ([]*types.Variable, error) {
	var variables []*types.Variable

	// Traverse the root group
	root := a.file.Root()
	a.traverseGroup(root, "", &variables)

	return variables, nil
}

// traverseGroup recursively processes groups and datasets.
func (a *HDF5Adapter) traverseGroup(group *hdf5.Group, path string, variables *[]*types.Variable) {
	// Check if this is a complex number group by looking for MATLAB_complex attribute
	// Complex groups have structure: group -> real/imag datasets
	isComplexGroup := false
	attrs, err := group.Attributes()
	if err == nil {
		for _, attr := range attrs {
			if attr.Name == "MATLAB_complex" {
				// Found MATLAB_complex attribute - this is a complex variable
				isComplexGroup = true
				break
			}
		}
	}

	if isComplexGroup {
		// This IS a complex variable - convert it and don't traverse children
		variable, err := a.convertComplexGroup(group, path)
		if err == nil {
			*variables = append(*variables, variable)
			return // Don't traverse children (real/imag datasets)
		}
		// If conversion failed, fall through to normal traversal
	}

	// Process all children (datasets and subgroups)
	for _, child := range group.Children() {
		switch obj := child.(type) {
		case *hdf5.Dataset:
			variable := a.convertDataset(obj, path)
			*variables = append(*variables, variable)
		case *hdf5.Group:
			newPath := path + "/" + obj.Name()
			a.traverseGroup(obj, newPath, variables)
		}
	}
}

// convertDataset converts HDF5 dataset to MATLAB variable.
func (a *HDF5Adapter) convertDataset(dataset *hdf5.Dataset, path string) *types.Variable {
	name := path + "/" + dataset.Name()
	if path == "" {
		name = dataset.Name()
	}

	// Determine MATLAB class from attributes
	matlabClass := matlabClassDouble
	if val, err := dataset.ReadAttribute("MATLAB_class"); err == nil {
		if strVal, ok := val.(string); ok {
			matlabClass = strVal
		}
	}

	// Determine data type
	dataType := types.Unknown
	switch matlabClass {
	case matlabClassDouble:
		dataType = types.Double
	case matlabClassSingle:
		dataType = types.Single
	case matlabClassInt8:
		dataType = types.Int8
	case matlabClassUint8:
		dataType = types.Uint8
	case matlabClassInt16:
		dataType = types.Int16
	case matlabClassUint16:
		dataType = types.Uint16
	case matlabClassInt32:
		dataType = types.Int32
	case matlabClassUint32:
		dataType = types.Uint32
	case matlabClassInt64:
		dataType = types.Int64
	case matlabClassUint64:
		dataType = types.Uint64
	case matlabClassChar:
		dataType = types.Char
	case matlabClassStruct:
		dataType = types.Struct
	case matlabClassCell:
		dataType = types.CellArray
	}

	// Read data - try numeric first, then strings as fallback.
	// This handles both numeric arrays and character/string datasets.
	var data interface{}
	var dims []int

	numData, err := dataset.Read()
	if err == nil {
		data = numData
		dims = []int{len(numData)}
	} else {
		// If numeric read fails, try string read
		strData, strErr := dataset.ReadStrings()
		if strErr == nil {
			data = strData
			dims = []int{len(strData)}
		} else {
			// Return empty data on error
			data = []float64{}
			dims = []int{0}
		}
	}

	// Create variable
	variable := &types.Variable{
		Name:       name,
		Dimensions: dims,
		DataType:   dataType,
		Data:       data,
	}

	// Add attributes
	attrs := make(map[string]interface{})
	attrList, err := dataset.Attributes()
	if err == nil {
		for _, attr := range attrList {
			// Note: core.Attribute has Name field, not Name() method
			attrs[attr.Name] = attr
		}
	}
	variable.Attributes = attrs

	return variable
}

// convertComplexGroup converts an HDF5 group representing a complex MATLAB variable.
//
// The group contains "real" and "imag" datasets with a MATLAB_complex attribute set to 1.
// This matches the structure created by the v7.3 writer for complex numbers.
func (a *HDF5Adapter) convertComplexGroup(group *hdf5.Group, name string) (*types.Variable, error) {
	// Find real and imag datasets in children
	var realDS, imagDS *hdf5.Dataset
	for _, child := range group.Children() {
		if ds, ok := child.(*hdf5.Dataset); ok {
			switch ds.Name() {
			case "real":
				realDS = ds
			case "imag":
				imagDS = ds
			}
		}
	}

	if realDS == nil {
		return nil, fmt.Errorf("complex group missing 'real' dataset")
	}
	if imagDS == nil {
		return nil, fmt.Errorf("complex group missing 'imag' dataset")
	}

	// Read real data (this also gives us dimensions)
	realData, err := realDS.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read real data: %w", err)
	}

	// Read imag data
	imagData, err := imagDS.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read imag data: %w", err)
	}

	// Determine dimensions from data length
	// Read() returns []float64, so use len() directly
	dimensions := []int{len(realData)}

	// Get MATLAB_class from real dataset (groups may not support ReadAttribute)
	var dataType types.DataType
	if val, err := realDS.ReadAttribute("MATLAB_class"); err == nil {
		if classStr, ok := val.(string); ok {
			dataType = a.matlabClassToDataType(classStr)
		}
	}

	// Default to Double if unknown
	if dataType == types.Unknown {
		dataType = types.Double
	}

	// Strip leading slash from name if present
	if name != "" && name[0] == '/' {
		name = name[1:]
	}

	// Create Variable
	return &types.Variable{
		Name:       name,
		IsComplex:  true,
		Dimensions: dimensions,
		DataType:   dataType,
		Data: &types.NumericArray{
			Real: realData,
			Imag: imagData,
		},
	}, nil
}

// matlabClassToDataType converts MATLAB class string to DataType.
func (a *HDF5Adapter) matlabClassToDataType(matlabClass string) types.DataType {
	switch matlabClass {
	case matlabClassDouble:
		return types.Double
	case matlabClassSingle:
		return types.Single
	case matlabClassInt8:
		return types.Int8
	case matlabClassUint8:
		return types.Uint8
	case matlabClassInt16:
		return types.Int16
	case matlabClassUint16:
		return types.Uint16
	case matlabClassInt32:
		return types.Int32
	case matlabClassUint32:
		return types.Uint32
	case matlabClassInt64:
		return types.Int64
	case matlabClassUint64:
		return types.Uint64
	case matlabClassChar:
		return types.Char
	case matlabClassStruct:
		return types.Struct
	case matlabClassCell:
		return types.CellArray
	default:
		return types.Unknown
	}
}
