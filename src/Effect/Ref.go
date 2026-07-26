func _New(val interface{}) func() interface{} {
	return func() interface{} {
		return map[string]interface{}{"value": val}
	}
}
func NewWithSelf(f func(interface{}) interface{}) func() interface{} {
	return func() interface{} {
		ref := map[string]interface{}{}
		ref["value"] = f(ref)
		return ref
	}
}
func Read(ref map[string]interface{}) func() interface{} {
	return func() interface{} {
		return ref["value"]
	}
}
func ModifyImpl(f func(interface{}) map[string]interface{}, ref map[string]interface{}) func() interface{} {
	return func() interface{} {
		t := f(ref["value"])
		ref["value"] = t["state"]
		return t["value"]
	}
}
func Write(val interface{}, ref map[string]interface{}) func() {
	return func() {
		ref["value"] = val
	}
}
