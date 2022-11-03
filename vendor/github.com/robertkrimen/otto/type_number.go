package otto

func (runtime *_runtime) newNumberObject(value Value) *_object {
	return runtime.newPrimitiveObject(classNumber, value.numberValue())
}
