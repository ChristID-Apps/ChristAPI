package activities

import "testing"

func TestMemberActivityVisibility(t *testing.T) {
	if got := activityStatusFilter(false, "draft"); got != "published" {
		t.Fatalf("member status filter = %q, want published", got)
	}
	if activityVisibleToUser(false, "draft") {
		t.Fatal("member should not see draft activity")
	}
	if !activityVisibleToUser(false, "published") {
		t.Fatal("member should see published activity")
	}
	if !activityVisibleToUser(true, "draft") {
		t.Fatal("admin should see draft activity")
	}
}
