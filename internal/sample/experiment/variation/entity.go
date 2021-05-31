// Package variation is a test package.
package variation

type (
	// Entity defines experiment variation data.
	Entity struct {
		Metadata Metadata `json:"metadata"`
	}

	// Metadata defines experiment variation entity metadata.
	Metadata struct {
		Courses []int `json:"courses"`
	}
)
