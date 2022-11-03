package otto

func (runtime *_runtime) newBooleanObject(value Value) *_object {
	return runtime.newPrimitiveObject(classBoolean, toValue_bool(value.bool()))
}
