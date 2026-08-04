package reposerverextract

import (
	"testing"

	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
)

// An OCI source may carry the oci:// scheme while the credential is registered
// under the bare host. Both forms must resolve to the same entry, otherwise the
// chart pull falls back to an anonymous request and the registry answers
// "basic credential not found".
func TestGetRepoIgnoresOCIScheme(t *testing.T) {
	const bare = "682033484092.dkr.ecr.eu-central-1.amazonaws.com/moia/platform-engineering-infra"
	withCreds := &v1alpha1.Repository{Repo: bare, Username: "AWS", Password: "token"}

	for _, tc := range []struct{ name, registered, looked string }{
		{"bare registered, schemed lookup", bare, "oci://" + bare},
		{"schemed registered, bare lookup", "oci://" + bare, bare},
		{"both schemed", "oci://" + bare, "oci://" + bare},
		{"mixed case", "OCI://" + bare, bare},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rc := &RepoCreds{reposByURL: map[string]*v1alpha1.Repository{
				normalizeRepoURL(tc.registered): withCreds,
			}}
			got := rc.GetRepo(tc.looked)
			if got.Username == "" || got.Password == "" {
				t.Fatalf("credentials lost: registered %q, looked up %q -> %+v",
					tc.registered, tc.looked, got)
			}
		})
	}
}

func TestNormalizeRepoURLStripsScheme(t *testing.T) {
	if a, b := normalizeRepoURL("oci://Ghcr.io/Org"), normalizeRepoURL("ghcr.io/org"); a != b {
		t.Fatalf("scheme not normalised away: %q != %q", a, b)
	}
	if got := normalizeRepoURL("https://github.com/moia-dev/x.git"); got != "https://github.com/moia-dev/x" {
		t.Fatalf("git URL handling changed: %q", got)
	}
}
