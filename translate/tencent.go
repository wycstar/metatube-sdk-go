package translate

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func sign(key []byte, msg string) []byte {
	hmac := hmac.New(sha256.New, key)
	hmac.Write([]byte(msg))
	return hmac.Sum(nil)
}

func buildCanonicalRequest(canonicalHeaders, signedHeaders, hashedRequestPayload string) string {
	return fmt.Sprintf("POST\n/\n\n%s\n%s\n%s", canonicalHeaders, signedHeaders, hashedRequestPayload)
}

func buildStringToSign(algorithm string, timestamp int64, credentialScope, hashedCanonicalRequest string) string {
	return fmt.Sprintf("%s\n%d\n%s\n%s", algorithm, timestamp, credentialScope, hashedCanonicalRequest)
}

func calculateSignature(secretKey []byte, stringToSign string) string {
	signature := hmac.New(sha256.New, secretKey)
	signature.Write([]byte(stringToSign))
	return hex.EncodeToString(signature.Sum(nil))
}

func buildAuthorization(algorithm, secretID, credentialScope, signedHeaders, signature string) string {
	return fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s", algorithm, secretID, credentialScope, signedHeaders, signature)
}

func TencentTranslate(q, source, target, key string) (string, error) {
	keys := strings.Split(key, "|")
	SECRET_ID, SECRET_KEY := keys[0], keys[1]
	payload := map[string]interface{}{
		"SourceText": q,
		"Source":     parseToTencentSupportedLanguage(source),
		"Target":     parseToTencentSupportedLanguage(target),
		"ProjectId":  1317866,
	}
	algorithm := "TC3-HMAC-SHA256"
	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).Format("2006-01-02")

	contentType := "application/json; charset=utf-8"
	signedHeaders := "content-type;host;x-tc-action"
	hashedRequestPayload := fmt.Sprintf("%x", sha256.Sum256([]byte(jsonToString(payload))))
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:tmt.tencentcloudapi.com\nx-tc-action:texttranslate\n", contentType)
	canonicalRequest := buildCanonicalRequest(canonicalHeaders, signedHeaders, hashedRequestPayload)

	credentialScope := fmt.Sprintf("%s/tmt/tc3_request", date)
	hashedCanonicalRequest := fmt.Sprintf("%x", sha256.Sum256([]byte(canonicalRequest)))
	stringToSign := buildStringToSign(algorithm, timestamp, credentialScope, hashedCanonicalRequest)

	secretDate := sign([]byte("TC3"+SECRET_KEY), date)
	secretService := sign(secretDate, "tmt")
	secretSigning := sign(secretService, "tc3_request")
	signature := calculateSignature(secretSigning, stringToSign)

	authorization := buildAuthorization(algorithm, SECRET_ID, credentialScope, signedHeaders, signature)

	url := "https://tmt.tencentcloudapi.com/"
	req, err := http.NewRequest("POST", url, strings.NewReader(jsonToString(payload)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", authorization)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Host", "tmt.tencentcloudapi.com")
	req.Header.Set("X-TC-Action", "TextTranslate")
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Version", "2018-03-21")
	req.Header.Set("X-TC-Region", "ap-beijing")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", err
	}

	targetText := result["Response"].(map[string]interface{})["TargetText"].(string)
	return targetText, nil
}

func jsonToString(data interface{}) string {
	bytes, _ := json.Marshal(data)
	return string(bytes)
}

func parseToTencentSupportedLanguage(lang string) string {
	if lang = strings.ToLower(lang); lang == "" || lang == "auto" /* auto detect */ {
		return "auto"
	}
	switch lang {
	case "zh", "chs", "zh-cn", "zh_cn", "zh-hans":
		return "zh"
	case "cht", "zh-tw", "zh_tw", "zh-hk", "zh_hk", "zh-hant":
		return "zh-TW"
	case "jp", "ja":
		return "ja"
	case "kor", "ko":
		return "ko"
	case "fra", "fr":
		return "fr"
	case "spa", "es":
		return "es"
	case "vie", "vi":
		return "vi"
	case "ara", "ar":
		return "ar"
	}
	return lang
}
