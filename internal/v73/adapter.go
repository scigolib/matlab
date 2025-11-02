package v73

import (
	"github.com/scigolib/hdf5"
	"github.com/scigolib/matlab/types"
)

const (
	// MATLAB class type identifiers.
	matlabClassDouble = "double"
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
	case "single":
		dataType = types.Single
	case "int8":
		dataType = types.Int8
	case "uint8":
		dataType = types.Uint8
	case "int16":
		dataType = types.Int16
	case "uint16":
		dataType = types.Uint16
	case "int32":
		dataType = types.Int32
	case "uint32":
		dataType = types.Uint32
	case "int64":
		dataType = types.Int64
	case "uint64":
		dataType = types.Uint64
	case "char":
		dataType = types.Char
	case "struct":
		dataType = types.Struct
	case "cell":
		dataType = types.CellArray
	}

	// Read data - try numeric first, then strings
	// TODO: Handle string datasets with ReadStrings()
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
