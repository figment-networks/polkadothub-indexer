package iterators

import (
	"github.com/figment-networks/polkadothub-indexer/types"
	"strconv"
	"testing"
)

func Test_HeightIterator(t *testing.T) {
	tests := []struct {
		name   string
		start  types.Height
		end    types.Height
		next   types.Height
		cont   bool
		length int64
		err    bool
	}{
		{"1-10", types.Height(1), types.Height(10), types.Height(2), true, 10, false},
		{"2-10", types.Height(2), types.Height(10), types.Height(3), true, 9, false},
		{"3-3", types.Height(3), types.Height(3), types.Height(4), false, 1, false},
		{"5-3", types.Height(5), types.Height(3), types.Height(5), false, -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewHeightIterator(tt.start, tt.end)
			cont := i.Next()
			if cont != tt.cont {
				t.Errorf("[Next()] exp: %s got: %s", strconv.FormatBool(tt.cont), strconv.FormatBool(cont))
			}
			if i.Value() != tt.next {
				t.Errorf("[Value()] exp: %d got: %d", tt.next, i.Value())
			}
			if i.Length() != tt.length {
				t.Errorf("[Length()] exp: %d got: %d", tt.length, i.Length())
			}
			if tt.err && i.Error() == nil {
				t.Errorf("[Error()] should be populated")
			}
		})
	}
}
