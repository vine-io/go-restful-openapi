package internal

type Model struct {
	ID string `json:"id"`
}

type UserS struct {
	// hello
	Model

	Name string `json:"name"`
	Age  int    `json:"age"`
}
