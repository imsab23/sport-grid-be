package bracket

import "testing"

func TestNextPowerOfTwo(t *testing.T) {
	cases := map[int]int{1: 1, 2: 2, 3: 4, 4: 4, 5: 8, 7: 8, 8: 8, 9: 16}
	for n, want := range cases {
		if got := nextPowerOfTwo(n); got != want {
			t.Errorf("nextPowerOfTwo(%d) = %d, want %d", n, got, want)
		}
	}
}

func TestSeedOrder(t *testing.T) {
	got := seedOrder(8)
	want := []int{1, 8, 4, 5, 2, 7, 3, 6}
	if len(got) != len(want) {
		t.Fatalf("seedOrder(8) length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("seedOrder(8)[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

// TestSeedOrderByesNeverPair proves the invariant Generate() relies on: for any n
// participants seeded into the next-power-of-two bracket, a bye slot (seed number > n)
// never pairs against another bye slot in round 1, so a single bye-resolution pass suffices.
func TestSeedOrderByesNeverPair(t *testing.T) {
	for n := 2; n <= 64; n++ {
		size := nextPowerOfTwo(n)
		order := seedOrder(size)

		for i := 0; i < size; i += 2 {
			aBye := order[i] > n
			bBye := order[i+1] > n
			if aBye && bBye {
				t.Fatalf("n=%d size=%d: both slots %d (seed %d) and %d (seed %d) are byes",
					n, size, i, order[i], i+1, order[i+1])
			}
		}
	}
}
