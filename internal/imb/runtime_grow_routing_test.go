package imb

import "testing"

// TestGrowRoutingIndependence is a regression guard for the write-through
// aliasing bug where GrowRouting reused dst's backing array when it had spare
// capacity. The result must be backed by its own memory so writes into it can
// never reach the caller's routing buffer.
func TestGrowRoutingIndependence(t *testing.T) {
	t.Run("spare capacity does not alias dst", func(t *testing.T) {
		dst := make([]byte, 2, 8)
		copy(dst, []byte("AB"))
		got := GrowRouting(dst, 'Z')
		if string(got) != "ABZ" {
			t.Fatalf("got %q want %q", got, "ABZ")
		}
		if len(got) != 3 {
			t.Fatalf("len(got)=%d want 3", len(got))
		}
		// Writing anywhere into the result must not touch dst.
		for i := range got {
			backup := dst[0]
			got[i] = 'X'
			if dst[0] != backup {
				t.Fatalf("write to got[%d] wrote through into dst", i)
			}
		}
	})

	t.Run("full capacity result still independent", func(t *testing.T) {
		// When dst is already at capacity append() would be forced to grow,
		// but the contract must hold regardless of the input's capacity.
		dst := []byte("AB")
		got := GrowRouting(dst, 'Z')
		if string(got) != "ABZ" {
			t.Fatalf("got %q want %q", got, "ABZ")
		}
		got[0] = 'X'
		if dst[0] != 'A' {
			t.Fatal("GrowRouting aliased dst at full capacity")
		}
	})

	t.Run("empty dst is independent", func(t *testing.T) {
		var dst []byte
		got := GrowRouting(dst, 'Z')
		if string(got) != "Z" {
			t.Fatalf("got %q want %q", got, "Z")
		}
		got[0] = 'X'
		// dst is nil/empty; nothing to assert about its backing store, but
		// the result must be its own 1-element slice.
		if len(got) != 1 || cap(got) < 1 {
			t.Fatalf("got len=%d cap=%d", len(got), cap(got))
		}
	})
}
