package Effect_Ref

import (
	r "gopurs/output/gopurs_runtime"
	"sync"
	"testing"
)

type api struct{ new, read, write, modify, modifyImpl, self r.Value }

var unitABI = api{_Gopurs_Effect_Ref__New, _Gopurs_Effect_Ref_Read, _Gopurs_Effect_Ref_Write, _Gopurs_Effect_Ref_Modify_, _Gopurs_Effect_Ref_ModifyImpl, _Gopurs_Effect_Ref_NewWithSelf}
var sink r.Value

func run(action r.Value) r.Value { return r.Apply(action, r.Value{}) }

func TestEffectSemantics(t *testing.T) {
	for _, tc := range []struct {
		name string
		api  api
	}{{"unit", unitABI}} {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.api
			fresh := r.Apply(a.new, r.Int(0))
			ref := run(fresh)
			other := run(fresh)
			if ref.AnyVal() == other.AnyVal() {
				t.Fatal("new action replay reused a reference")
			}
			read := r.Apply(a.read, ref)
			write := r.Apply2(a.write, r.Int(8), ref)
			if run(read).IntVal != 0 {
				t.Fatal("write ran before Effect application")
			}
			run(write)
			if run(read).IntVal != 8 {
				t.Fatal("read replay did not observe write")
			}
			order := []int64{}
			callback := r.Func(func(v r.Value) r.Value { order = append(order, v.IntVal); return r.Int(v.IntVal + 1) })
			modify := r.Apply2(a.modify, callback, ref)
			if len(order) != 0 {
				t.Fatal("callback ran while constructing action")
			}
			run(modify)
			run(modify)
			if len(order) != 2 || order[0] != 8 || order[1] != 9 || run(read).IntVal != 10 {
				t.Fatal("callback order/replay", order)
			}
			run(write)
			if run(read).IntVal != 8 {
				t.Fatal("write replay failed")
			}
			update := r.Func(func(v r.Value) r.Value { return r.RecordDict2("state", "value", r.Int(v.IntVal+4), v) })
			if got := run(r.Apply2(a.modifyImpl, update, ref)).IntVal; got != 8 || run(read).IntVal != 12 {
				t.Fatal("modifyImpl state/value order")
			}
			selfCalls := 0
			self := r.Apply(a.self, r.Func(func(v r.Value) r.Value { selfCalls++; return v }))
			if selfCalls != 0 {
				t.Fatal("newWithSelf ran eagerly")
			}
			selfRef := run(self)
			if selfCalls != 1 || run(r.Apply(a.read, selfRef)).AnyVal() != selfRef.AnyVal() {
				t.Fatal("self reference")
			}
			run(self)
			if selfCalls != 2 {
				t.Fatal("self replay")
			}
		})
	}
}

func TestConcurrentModify(t *testing.T) {
	for _, a := range []api{unitABI} {
		ref := run(r.Apply(a.new, r.Int(0)))
		modify := r.Apply2(a.modify, r.Func(func(v r.Value) r.Value { return r.Int(v.IntVal + 1) }), ref)
		var wg sync.WaitGroup
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					run(modify)
				}
			}()
		}
		wg.Wait()
		if run(r.Apply(a.read, ref)).IntVal != 400 {
			t.Fatal("lost concurrent update")
		}
	}
}

func TestCallbackPanicIdentity(t *testing.T) {
	marker := &struct{ n int }{42}
	for _, a := range []api{unitABI} {
		ref := run(r.Apply(a.new, r.Int(0)))
		action := r.Apply2(a.modify, r.Func(func(r.Value) r.Value { panic(marker) }), ref)
		func() {
			defer func() {
				if recover() != marker {
					t.Fatal("panic identity changed")
				}
			}()
			run(action)
		}()
		// The existing FFI retains the lock if its callback panics. Do not reuse
		// this reference: the explicit Unit ABI leaves that policy unchanged.
	}
}
