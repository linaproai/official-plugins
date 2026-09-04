// This file verifies department leader selector projections and merge rules.

package dept

import (
	"testing"

	"lina-core/pkg/plugin/capability/usercap"
)

func TestToDeptUsersConvertsNumericIDs(t *testing.T) {
	t.Parallel()

	rows := []*usercap.UserInfo{
		{ID: "5", Username: "user004", Nickname: "Liu Yang"},
		{ID: "bad", Username: "invalid", Nickname: "Invalid"},
		nil,
	}
	got := toDeptUsers(rows, 10)
	if len(got) != 1 {
		t.Fatalf("expected 1 selectable user, got %d", len(got))
	}
	if got[0].Id != 5 || got[0].Username != "user004" || got[0].Nickname != "Liu Yang" {
		t.Fatalf("unexpected selectable user: %+v", got[0])
	}
}

func TestMergeLeaderUserPrependsMissingLeader(t *testing.T) {
	t.Parallel()

	users := []*DeptUser{{Id: 1, Username: "member", Nickname: "Member"}}
	leader := &DeptUser{Id: 5, Username: "user004", Nickname: "Liu Yang"}
	got := mergeLeaderUser(users, leader, 10)
	if len(got) != 2 {
		t.Fatalf("expected leader plus member, got %d users", len(got))
	}
	if got[0].Id != 5 || got[1].Id != 1 {
		t.Fatalf("expected leader first, got %+v", got)
	}
}

func TestMergeLeaderUserDoesNotDuplicate(t *testing.T) {
	t.Parallel()

	users := []*DeptUser{{Id: 5, Username: "user004", Nickname: "Liu Yang"}}
	leader := &DeptUser{Id: 5, Username: "user004", Nickname: "Liu Yang"}
	got := mergeLeaderUser(users, leader, 10)
	if len(got) != 1 {
		t.Fatalf("expected existing leader to stay unique, got %d users", len(got))
	}
}

func TestMergeLeaderUserTrimsToLimit(t *testing.T) {
	t.Parallel()

	users := []*DeptUser{{Id: 1}, {Id: 2}}
	leader := &DeptUser{Id: 9, Username: "leader", Nickname: "Leader"}
	got := mergeLeaderUser(users, leader, 2)
	if len(got) != 2 {
		t.Fatalf("expected bounded list of 2, got %d", len(got))
	}
	if got[0].Id != 9 || got[1].Id != 1 {
		t.Fatalf("expected leader to replace the last candidate, got %+v", got)
	}
}
