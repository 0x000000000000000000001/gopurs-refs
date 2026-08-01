package Effect_Ref

func _New(val interface{}, _ interface{}) interface{} {
	return map[string]interface{}{"value": val}
}
func NewWithSelf(f func(interface{}) interface{}, _ interface{}) interface{} {
	ref := map[string]interface{}{}
	ref["value"] = f(ref)
	return ref
}
func Read(ref interface{}, _ interface{}) interface{} {
	return ref.(map[string]interface{})["value"]
}
func ModifyImpl(f func(interface{}) interface{}, ref interface{}, _ interface{}) interface{} {
	t := f(ref.(map[string]interface{})["value"]).(map[string]interface{})
	ref.(map[string]interface{})["value"] = t["state"]
	return t["value"]
}
func Write(val interface{}, ref interface{}, _ interface{}) interface{} {
	ref.(map[string]interface{})["value"] = val
	return nil
}
