package orchestration

import "testing"

func TestResultValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		result  Result
		wantErr bool
	}{
		{name: "completed", result: Result{Status: ResultCompleted, Summary: "done"}},
		{name: "needs advice", result: Result{Status: ResultNeedsAdvice, Summary: "blocked", AdviceRequest: &AdviceRequest{Question: "which?", Reason: "ambiguous"}}},
		{name: "missing advice", result: Result{Status: ResultNeedsAdvice, Summary: "blocked"}, wantErr: true},
		{name: "unknown", result: Result{Status: "wat", Summary: "bad"}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := test.result.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestValidateTasksRejectsCycle(t *testing.T) {
	t.Parallel()
	tasks := []Task{
		{ID: "a", Objective: "a", Role: RoleWorker, Permission: PermissionRead, DependsOn: []string{"b"}},
		{ID: "b", Objective: "b", Role: RoleWorker, Permission: PermissionRead, DependsOn: []string{"a"}},
	}
	if err := ValidateTasks(tasks); err == nil {
		t.Fatal("ValidateTasks() accepted a cycle")
	}
}
