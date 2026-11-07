package quota

import "testing"

func TestThresholdCrossing(t *testing.T) {
	for _, tc := range []struct {
		old, next int
		want      bool
	}{{21, 19, true}, {20, 19, true}, {19, 18, false}, {19, 25, false}, {25, 20, false}} {
		if got := CrossedThreshold([]Window{{Minutes: 300, Remaining: tc.old}}, []Window{{Minutes: 300, Remaining: tc.next}}, 20); got != tc.want {
			t.Fatalf("%+v: %v", tc, got)
		}
	}
	if CrossedThreshold(nil, []Window{{Remaining: 1}}, 20) {
		t.Fatal("startup alert")
	}
	if CrossedThreshold([]Window{{Minutes: 300, Remaining: 90}}, []Window{{Minutes: 10080, Remaining: 1}}, 20) {
		t.Fatal("different window")
	}
}
