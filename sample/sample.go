package sample

type TestSampleStruct struct {
	SimpleFloat64 float64 `json:"simple_float64"`
	SimpleBool    bool    `json:"simple_bool"`

	Sub      TestSubStruct   `json:"sub"`
	SubSlice []TestSubStruct `json:"sub_slice"`
}

type TestSubStruct struct {
	SubInt int `json:"sample_int"`
}
