package frozencounter

// #cgo darwin CFLAGS: -I/System/Library/Frameworks/Accelerate.framework/Versions/A/Frameworks/vecLib.framework/Versions/A/Headers/
// #cgo darwin LDFLAGS: -L/System/Library/Frameworks/Accelerate.framework//Versions/A/Frameworks/vecLib.framework/Versions/A/ -lBLAS
// #include <cblas.h>
import "C"
import "unsafe"

type vector []float64

func (v vector) copy() vector {
	r := make(vector, len(v))

	C.cblas_dcopy(C.int(len(v)), (*C.double)(unsafe.Pointer(&v[0])), 1, (*C.double)(unsafe.Pointer(&r[0])), 1)

	return r
}

func (v vector) reset(val float64) {
	C.catlas_dset(C.int(len(v)), C.double(val), (*C.double)(unsafe.Pointer(&v[0])), 1)
}

// Return the sum of v
func (v vector) sum() float64 {
	return float64(C.cblas_dasum(C.int(len(v)), (*C.double)(unsafe.Pointer(&v[0])), 1))
}

// Dot product of v & o
func (v vector) dot(o vector) float64 {
	c1 := (*C.double)(unsafe.Pointer(&v[0]))
	c2 := (*C.double)(unsafe.Pointer(&o[0]))

	return float64(C.cblas_ddot(C.int(len(v)), c1, 1, c2, 1))
}

// v += o
func (v vector) add(o vector) {
	input := (*C.double)(unsafe.Pointer(&o[0]))
	output := (*C.double)(unsafe.Pointer(&v[0]))

	C.cblas_daxpy(C.int(len(v)), 1.0, input, 1, output, 1)
}

// v += scale * o
func (v vector) addScaled(scale float64, o vector) {
	input := (*C.double)(unsafe.Pointer(&o[0]))
	output := (*C.double)(unsafe.Pointer(&v[0]))

	C.cblas_daxpy(C.int(len(v)), C.double(scale), input, 1, output, 1)
}

// v -= o
func (v vector) subtract(o vector) {
	input := (*C.double)(unsafe.Pointer(&o[0]))
	output := (*C.double)(unsafe.Pointer(&v[0]))

	C.catlas_daxpby(C.int(len(v)), 1.0, input, 1, -1.0, output, 1)
}

// Scale the values in v by a (v *= a)
func (v vector) scale(a float64) {
	c := (*C.double)(unsafe.Pointer(&v[0]))

	C.cblas_dscal(C.int(len(v)), C.double(a), c, 1)
}

// Find the argmax of v
func (v vector) argmax() int {
	return int(C.cblas_idamax(C.int(len(v)), (*C.double)(unsafe.Pointer(&v[0])), 1))
}
