package adapter

func stack_contains_all_conditions(stack map[string]struct{}, conditions []string) bool {
	for _, trigger := range conditions {
		if _, found := stack[trigger]; !found {
			return false
		}
	}
	return true
}
func remove_items_from_stack(stack map[string]struct{}, items []string) {
	for _, item := range items {
		delete(stack, item)
	}
}
func MeetAllConditions(downlink chan string, uplink chan string, loglink chan string, conditions []string, emission string) {
	stack := make(map[string]struct{})
	for msg := range downlink {
		stack[msg] = struct{}{}
		if stack_contains_all_conditions(stack, conditions) {
			for _, c := range conditions {
				loglink <- c + "->" + emission
			}
			uplink <- emission
			remove_items_from_stack(stack, conditions)
		}
	}
}
func MeetSomeConditions(downlink chan string, uplink chan string, loglink chan string, conditions []string, emission string) {
	for msg := range downlink {
		for _, c := range conditions {
			if msg == c {
				loglink <- c + "->" + emission
				uplink <- emission
				break
			}
		}
	}
}
