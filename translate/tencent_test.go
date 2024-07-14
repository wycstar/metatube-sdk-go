package translate

import (
	"fmt"
	"testing"
)

func TestTencentTranslate(t *testing.T) {
	text := "ニュースキャスターは生ハメ本番中"
	translation, err := TencentTranslate(text, "auto", "zh", "")
	if err != nil {
		fmt.Println("Translation Error:", err)
		return
	}
	fmt.Println(translation)
}
