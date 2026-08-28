package contentplan

import (
	"reflect"
	"testing"
)

func TestDecodeItemsNormalizesMixedModelOutput(t *testing.T) {
	input := []byte(`[
		"直接文本",
		{"title":"增长","body":"同比提升 20%"},
		{"label":"效率","value":3.5},
		42,
		true,
		null,
		{"z":"末项","a":"首项"}
	]`)
	want := []string{
		"直接文本",
		"增长: 同比提升 20%",
		"效率: 3.5",
		"42",
		"true",
		"a: 首项; z: 末项",
	}

	got, err := DecodeItems(input)
	if err != nil {
		t.Fatalf("DecodeItems() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DecodeItems() = %#v, want %#v", got, want)
	}
}

func TestDecodeItemsAcceptsSingleObject(t *testing.T) {
	got, err := DecodeItems([]byte(`{"name":"结论","text":"优先解决契约问题"}`))
	if err != nil {
		t.Fatalf("DecodeItems() error = %v", err)
	}
	want := []string{"结论: 优先解决契约问题"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DecodeItems() = %#v, want %#v", got, want)
	}
}

func TestDecodeItemsDoesNotTreatDescriptionAsBody(t *testing.T) {
	got, err := DecodeItems([]byte(`{"title":"增长","description":"旧字段不再作为正文"}`))
	if err != nil {
		t.Fatalf("DecodeItems() error = %v", err)
	}
	want := []string{"增长"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DecodeItems() = %#v, want %#v", got, want)
	}
}
