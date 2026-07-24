func _New(val any) func() any {
	return func() any {
		return map[string]any{"value": val}
	}
}
func NewWithSelf(f func(any) any) func() any {
	return func() any {
		ref := map[string]any{}
		ref["value"] = f(ref)
		return ref
	}
}
func Read(ref map[string]any) func() any {
	return func() any {
		return ref["value"]
	}
}
func ModifyImpl(f func(any) map[string]any, ref map[string]any) func() any {
	return func() any {
		t := f(ref["value"])
		ref["value"] = t["state"]
		return t["value"]
	}
}
func Write(val any, ref map[string]any) func() {
	return func() {
		ref["value"] = val
	}
}
