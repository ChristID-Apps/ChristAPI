package news

import "testing"

func TestMemberNewsStatusFilter(t *testing.T) {
	if got := newsStatusFilter(false, "draft"); got != "published" {
		t.Fatalf("member status filter = %q, want published", got)
	}
	if got := newsStatusFilter(true, "draft"); got != "draft" {
		t.Fatalf("admin status filter = %q, want draft", got)
	}
	if got := newsStatusFilter(true, ""); got != "" {
		t.Fatalf("empty admin status filter = %q, want empty", got)
	}
}
