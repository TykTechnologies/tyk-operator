package controllers

import (
	"encoding/json"

	"github.com/buger/jsonparser"
)

var (
	OASClientCertAllowlistPath = []string{TykOASExtenstionStr, "server", "clientCertificates", "allowlist"}
	OASClientCertEnabledPath   = []string{TykOASExtenstionStr, "server", "clientCertificates", "enabled"}
)

func OASSetClientCertificatesAllowlist(data []byte, certid string) ([]byte, error) {
	val, _, _, err := jsonparser.Get(data, OASClientCertAllowlistPath...)
	if err != nil && err != jsonparser.KeyPathNotFoundError {
		return nil, err
	}

	allowlist := []string{}
	mCert := map[string]bool{}

	if val != nil {
		_, err = jsonparser.ArrayEach(val, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			item, pErr := jsonparser.ParseString(value)
			if pErr != nil {
				return
			}

			allowlist = append(allowlist, item)
			mCert[item] = true
		})
		if err != nil {
			return nil, err
		}
	}

	if _, ok := mCert[certid]; !ok {
		allowlist = append(allowlist, certid)
	}

	val, err = json.Marshal(allowlist)
	if err != nil {
		return nil, err
	}

	return jsonparser.Set(data, val, OASClientCertAllowlistPath...)
}
