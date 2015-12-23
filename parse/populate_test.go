package parse
import (
	"testing"
	"fmt"
	"os"
)

type MyConfig struct {
	SectionA struct {
				 Int      int `config:"int,-9"` // test the default val
				 Uint     uint
				 Int32    int32
				 Int64    int64
				 Float32  float32
				 Float64  float64
				 GoodEnv  bool
				 IntEnv  int
				 BackedupEnv  int
				 BadInt      int
				 BadUint     uint
				 BadInt32    int32
				 BadInt64    int64
				 BadFloat32  float32
				 BadFloat64  float64
				 BadString   string

				 Float32Slice []float32
				 Float64Slice []float64

				 IntSlice []int
				 Int8Slice []int8
				 Int16Slice []int16
				 Int32Slice []int32
				 Int64Slice []int64

				 UintSlice []uint
				 Uint8Slice []uint8
				 Uint16Slice []uint16
				 Uint32Slice []uint32
				 Uint64Slice []uint64
				 SectionB struct {
							  Int     int
							  Uint    uint
							  Int32   int32
							  Int64   int64
							  Float32 float32
							  Float64 float64
							  BadInt      int
							  BadUint     uint
							  BadInt32    int32
							  BadInt64    int64
							  BadFloat32  float32
							  BadFloat64  float64
							  BadString   string
							  String  string `config:"strong-string"`  // test the renaming
						  }
			 }
}

func TestPopulate(t *testing.T) {

	var (
		tree *Tree
		err  error
	)

	os.Setenv("TEST", "true")
	os.Setenv("TEST-INT", "888")
	if tree, err = ParseFile("./test.conf"); err != nil {
		t.Error("Failed to read ./test.conf:" + err.Error())
	}

	// test existing value stays if no default or value in config file
	testStruct := &MyConfig{}
	testStruct.SectionA.Int64 = 88888

	Populate(testStruct, tree.GetConfig(), "")

	t.Logf("After populate: %+v", testStruct)
	if fmt.Sprintf("%+v", testStruct) != "&{SectionA:{Int:-9 Uint:999 Int32:-999 Int64:88888 Float32:999.999 Float64:999.999 GoodEnv:true IntEnv:888 BackedupEnv:999 BadInt:0 BadUint:0 BadInt32:0 BadInt64:0 BadFloat32:0 BadFloat64:0 BadString: Float32Slice:[0.1 2 0.3 4] Float64Slice:[0.1 2 0.3 4] IntSlice:[1 2 3 4] Int8Slice:[1 2 3 4] Int16Slice:[1 2 3 4] Int32Slice:[1 2 3 4] Int64Slice:[1 2 3 4] UintSlice:[1 2 3 4] Uint8Slice:[1 2 3 4] Uint16Slice:[1 2 3 4] Uint32Slice:[1 2 3 4] Uint64Slice:[1 2 3 4] SectionB:{Int:-999 Uint:999 Int32:-999 Int64:999 Float32:999.999 Float64:999.999 BadInt:0 BadUint:0 BadInt32:0 BadInt64:0 BadFloat32:0 BadFloat64:0 BadString: String:lalala}}}" {
		t.Logf("Got: %+v", testStruct)
		t.Error("Not as expected.")
	}

}
