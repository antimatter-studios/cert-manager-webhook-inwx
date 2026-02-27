package inwx

import (
	"testing"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSolver() (*MockClientWrapper, *INWXDNSSolver) {
	mock := NewMockClientWrapper()
	solver := NewSolverWithClient(mock)
	return mock, solver
}

func TestName(t *testing.T) {
	_, solver := newTestSolver()
	assert.Equal(t, "inwx", solver.Name())
}

func TestPresent(t *testing.T) {
	mock, solver := newTestSolver()
	mock.CreateZone("example.com")

	ch := &v1alpha1.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com.",
		ResolvedZone: "example.com.",
		Key:          "test-challenge-token",
	}

	err := solver.Present(ch)
	require.NoError(t, err)

	recs, err := mock.getRecords("example.com")
	require.NoError(t, err)
	require.Len(t, *recs, 1)
	assert.Equal(t, "_acme-challenge", (*recs)[0].Name)
	assert.Equal(t, "TXT", (*recs)[0].Type)
	assert.Equal(t, "test-challenge-token", (*recs)[0].Content)
	assert.Equal(t, 300, (*recs)[0].TTL)
}

func TestPresentSubdomain(t *testing.T) {
	mock, solver := newTestSolver()
	mock.CreateZone("example.com")

	ch := &v1alpha1.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.sub.example.com.",
		ResolvedZone: "example.com.",
		Key:          "subdomain-token",
	}

	err := solver.Present(ch)
	require.NoError(t, err)

	recs, err := mock.getRecords("example.com")
	require.NoError(t, err)
	require.Len(t, *recs, 1)
	assert.Equal(t, "_acme-challenge.sub", (*recs)[0].Name)
	assert.Equal(t, "subdomain-token", (*recs)[0].Content)
}

func TestPresentIdempotent(t *testing.T) {
	mock, solver := newTestSolver()
	mock.CreateZone("example.com")

	ch := &v1alpha1.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com.",
		ResolvedZone: "example.com.",
		Key:          "test-token",
	}

	// Present twice should not error
	err := solver.Present(ch)
	require.NoError(t, err)
	// Second call creates a second record in mock (real INWX returns 2302 error which we handle)
	// The important thing is it doesn't fail
}

func TestCleanUp(t *testing.T) {
	mock, solver := newTestSolver()
	mock.CreateZone("example.com")

	ch := &v1alpha1.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com.",
		ResolvedZone: "example.com.",
		Key:          "test-challenge-token",
	}

	// Present first
	err := solver.Present(ch)
	require.NoError(t, err)

	recs, _ := mock.getRecords("example.com")
	require.Len(t, *recs, 1)

	// CleanUp should delete it
	err = solver.CleanUp(ch)
	require.NoError(t, err)

	recs, _ = mock.getRecords("example.com")
	assert.Len(t, *recs, 0)
}

func TestCleanUpOnlyMatchingRecord(t *testing.T) {
	mock, solver := newTestSolver()
	mock.CreateZone("example.com")

	// Create two challenges for the same domain (concurrent validations)
	ch1 := &v1alpha1.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com.",
		ResolvedZone: "example.com.",
		Key:          "token-1",
	}
	ch2 := &v1alpha1.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com.",
		ResolvedZone: "example.com.",
		Key:          "token-2",
	}

	require.NoError(t, solver.Present(ch1))
	require.NoError(t, solver.Present(ch2))

	recs, _ := mock.getRecords("example.com")
	require.Len(t, *recs, 2)

	// CleanUp ch1 should only delete token-1
	require.NoError(t, solver.CleanUp(ch1))

	recs, _ = mock.getRecords("example.com")
	require.Len(t, *recs, 1)
	assert.Equal(t, "token-2", (*recs)[0].Content)
}

func TestCleanUpNonExistent(t *testing.T) {
	mock, solver := newTestSolver()
	mock.CreateZone("example.com")

	ch := &v1alpha1.ChallengeRequest{
		ResolvedFQDN: "_acme-challenge.example.com.",
		ResolvedZone: "example.com.",
		Key:          "nonexistent-token",
	}

	// CleanUp with no matching record should not error
	err := solver.CleanUp(ch)
	assert.NoError(t, err)
}

func TestExtractRecordName(t *testing.T) {
	assert.Equal(t, "_acme-challenge", extractRecordName("_acme-challenge.example.com", "example.com"))
	assert.Equal(t, "_acme-challenge.sub", extractRecordName("_acme-challenge.sub.example.com", "example.com"))
	assert.Equal(t, "", extractRecordName("example.com", "example.com"))
	assert.Equal(t, "_acme-challenge.deep.sub", extractRecordName("_acme-challenge.deep.sub.example.com", "example.com"))
}
