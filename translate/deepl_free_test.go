package translate

import (
	"testing"
)

func TestDeeplFreeTranslate(t *testing.T) {
	for _, unit := range []struct {
		text, from, to string
	}{
		{"美形崩壊拘束輪カン窒息汁!", "", "zh-CN"},
	} {
		result, err := DeeplFreeTranslate(unit.text, unit.from, unit.to)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(result)
	}
}
