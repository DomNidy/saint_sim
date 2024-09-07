package utils

// todo: Does go gc will clean this up, right?: https://tip.golang.org/doc/gc-guide
// todo: also, is this safe? aren't we assigning i to the memory that gets allocated
// todo: for the function params, which, is local to the stack frame of this func?
// todo: idk, seems like the returned values are valid anyway.
// Helper function used to concisely 'inline' an int pointer
func IntPtr(i int) *int {
	return &i
}
