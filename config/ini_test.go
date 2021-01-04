package config

import (
	"testing"
)

func TestNew(t *testing.T) {
	cfg, err := New("testdata/test.conf")
	assert(t, err == nil)
	for _, tc := range testcases {
		sec, err := cfg.GetSection(tc.sec)
		assert(t, err == nil)
		val1, err := sec.GetString(tc.strKey)
		assert(t, err == nil)
		assert(t, val1 == tc.strVal)
		val2, err := sec.GetInt(tc.intKey)
		assert(t, err == nil)
		assert(t, val2 == tc.intVal)
		val3, err := sec.GetBool(tc.boolKey)
		assert(t, err == nil)
		assert(t, val3 == tc.boolVal)
		val4, err := sec.GetFloat(tc.floatKey)
		assert(t, err == nil)
		assert(t, val4 == tc.floatVal)

		notExist := "@not_exist@"
		_, err = sec.GetString(notExist)
		assert(t, err != nil)
		_, err = sec.GetInt(notExist)
		assert(t, err != nil)
		_, err = sec.GetBool(notExist)
		assert(t, err != nil)
		_, err = sec.GetFloat(notExist)
		assert(t, err != nil)
	}

	_, err = New("testdata/test1.conf")
	assert(t, err != nil)
}

func TestGetSetting(t *testing.T) {
	cfg, err := New("testdata/test.conf")
	assert(t, err == nil)

	for _, tc := range testcases {
		_, err = cfg.GetSetting(tc.sec, tc.strKey)
		assert(t, err == nil)
		_, err = cfg.GetIntSetting(tc.sec, tc.intKey)
		assert(t, err == nil)
		_, err = cfg.GetBoolSetting(tc.sec, tc.boolKey)
		assert(t, err == nil)
		_, err = cfg.GetFloatSetting(tc.sec, tc.floatKey)
		assert(t, err == nil)

		notExist := "@not_exist@"
		_, err = cfg.GetSetting(tc.sec, notExist)
		assert(t, err != nil)
		_, err = cfg.GetIntSetting(tc.sec, notExist)
		assert(t, err != nil)
		_, err = cfg.GetBoolSetting(tc.sec, notExist)
		assert(t, err != nil)
		_, err = cfg.GetFloatSetting(tc.sec, notExist)
		assert(t, err != nil)

		_, err = cfg.GetSetting(notExist, tc.strKey)
		assert(t, err != nil)
		_, err = cfg.GetIntSetting(notExist, tc.intKey)
		assert(t, err != nil)
		_, err = cfg.GetBoolSetting(notExist, tc.boolKey)
		assert(t, err != nil)
		_, err = cfg.GetFloatSetting(notExist, tc.floatKey)
		assert(t, err != nil)
	}
}

func TestGetSettingMust(t *testing.T) {
	cfg, err := New("testdata/test.conf")
	assert(t, err == nil)
	for _, tc := range testcases {
		sec, err := cfg.GetSection(tc.sec)
		assert(t, err == nil)
		assert(t, "abc" == sec.GetStringMust("not_exist", "abc"))
		assert(t, 123 == sec.GetIntMust("not_exist", 123))
		assert(t, true == sec.GetBoolMust("not_exist", true))
	}
}

func TestNewConfigWithFormatType(t *testing.T) {
	_, err := NewConfigWithFormatType(ConfFormatTypeIni, "testdata/test.conf")
	assert(t, err == nil)
	_, err = NewConfigWithFormatType(ConfFormatTypeToml, "testdata/test.toml")
	assert(t, err == nil)
}

func TestGetAllSections(t *testing.T) {
	cfg, err := NewConfigWithFormatType(ConfFormatTypeIni, "testdata/test.conf")
	assert(t, err == nil)
	secs := cfg.GetAllSections()
	assert(t, secs != nil && len(secs) > 0)
}

func TestLoadAndLastMod(t *testing.T) {
	cfg, err := New("testdata/test.conf")
	assert(t, err == nil)
	err = cfg.Load()
	assert(t, err == nil)
	last, err := cfg.LastModify()
	assert(t, err == nil && last.Unix() > 0)
}
