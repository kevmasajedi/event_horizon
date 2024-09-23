package adapter

import (
	"event_horizon/system/hub"
)

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
func MeetAllConditions(hub *hub.Hub, conditions []string, emission string) {
	stack := make(map[string]struct{})
	for msg := range hub.DownLink() {
		stack[msg] = struct{}{}
		if stack_contains_all_conditions(stack, conditions) {
			for _, c := range conditions {
				hub.LogLink() <- c + "->" + emission
			}
			hub.UpLink() <- emission
			remove_items_from_stack(stack, conditions)
		}
	}
}
func MeetSomeConditions(hub *hub.Hub, conditions []string, emission string) {
	for msg := range hub.DownLink() {
		for _, c := range conditions {
			if msg == c {
				hub.LogLink() <- c + "->" + emission
				hub.UpLink() <- emission
				break
			}
		}
	}
}
