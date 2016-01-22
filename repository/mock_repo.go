/*
Copyright 2015 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package repository

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"strings"
)

// Constants used for testing.
// We initialize our mock repo with two branches (one of which holds a pending review),
// and commit history that looks like this:
//
//  Master Branch:    A--B--D--E--F--J
//                     \   /    \  \
//                       C       \  \
//                                \  \
//  Review Branch:                 G--H--I
//
// Where commits "B" and "D" represent reviews that have been submitted, and "G"
// is a pending review.
const (
	TestTargetRef   = "refs/heads/master"
	TestReviewRef   = "refs/heads/ojarjur/mychange"
	TestRequestsRef = "refs/notes/devtools/reviews"
	TestCommentsRef = "refs/notes/devtools/discuss"

	TestCommitA = "A"
	TestCommitB = "B"
	TestCommitC = "C"
	TestCommitD = "D"
	TestCommitE = "E"
	TestCommitF = "F"
	TestCommitG = "G"
	TestCommitH = "H"
	TestCommitI = "I"
	TestCommitJ = "J"

	TestRequestB = `{"timestamp": "0000000001", "reviewRef": "refs/heads/ojarjur/mychange", "targetRef": "refs/heads/master", "requester": "ojarjur", "reviewers": ["ojarjur"], "description": "B"}`
	TestRequestD = `{"timestamp": "0000000002", "reviewRef": "refs/heads/ojarjur/mychange", "targetRef": "refs/heads/master", "requester": "ojarjur", "reviewers": ["ojarjur"], "description": "D"}`
	TestRequestG = `{"timestamp": "0000000004", "reviewRef": "refs/heads/ojarjur/mychange", "targetRef": "refs/heads/master", "requester": "ojarjur", "reviewers": ["ojarjur"], "description": "G"}

{"timestamp": "0000000005", "reviewRef": "refs/heads/ojarjur/mychange", "targetRef": "refs/heads/master", "requester": "ojarjur", "reviewers": ["ojarjur"], "description": "Updated description of G"}

{"timestamp": "0000000005", "reviewRef": "refs/heads/ojarjur/mychange", "targetRef": "refs/heads/master", "requester": "ojarjur", "reviewers": ["ojarjur"], "description": "Final description of G"}`

	TestDiscussB = `{"timestamp": "0000000001", "author": "ojarjur", "location": {"commit": "B"}, "resolved": true}`
	TestDiscussD = `{"timestamp": "0000000003", "author": "ojarjur", "location": {"commit": "E"}, "resolved": true}`
)

type mockCommit struct {
	Message string   `json:"message,omitempty"`
	Time    string   `json:"time,omitempty"`
	Parents []string `json:"parents,omitempty"`
}

// mockRepoForTest defines an instance of Repo that can be used for testing.
type mockRepoForTest struct {
	Head    string
	Refs    map[string]string            `json:"refs,omitempty"`
	Commits map[string]mockCommit        `json:"commits,omitempty"`
	Notes   map[string]map[string]string `json:"notes,omitempty"`
}

// NewMockRepoForTest returns a mocked-out instance of the Repo interface that has been pre-populated with test data.
func NewMockRepoForTest() Repo {
	commitA := mockCommit{
		Message: "First commit",
		Time:    "0",
		Parents: nil,
	}
	commitB := mockCommit{
		Message: "Second commit",
		Time:    "1",
		Parents: []string{TestCommitA},
	}
	commitC := mockCommit{
		Message: "No, I'm the second commit",
		Time:    "1",
		Parents: []string{TestCommitA},
	}
	commitD := mockCommit{
		Message: "Fourth commit",
		Time:    "2",
		Parents: []string{TestCommitB, TestCommitC},
	}
	commitE := mockCommit{
		Message: "Fifth commit",
		Time:    "3",
		Parents: []string{TestCommitD},
	}
	commitF := mockCommit{
		Message: "Sixth commit",
		Time:    "4",
		Parents: []string{TestCommitE},
	}
	commitG := mockCommit{
		Message: "No, I'm the sixth commit",
		Time:    "4",
		Parents: []string{TestCommitE},
	}
	commitH := mockCommit{
		Message: "Seventh commit",
		Time:    "5",
		Parents: []string{TestCommitG, TestCommitF},
	}
	commitI := mockCommit{
		Message: "Eighth commit",
		Time:    "6",
		Parents: []string{TestCommitH},
	}
	commitJ := mockCommit{
		Message: "No, I'm the eighth commit",
		Time:    "6",
		Parents: []string{TestCommitF},
	}
	return mockRepoForTest{
		Head: TestTargetRef,
		Refs: map[string]string{
			TestTargetRef: TestCommitJ,
			TestReviewRef: TestCommitI,
		},
		Commits: map[string]mockCommit{
			TestCommitA: commitA,
			TestCommitB: commitB,
			TestCommitC: commitC,
			TestCommitD: commitD,
			TestCommitE: commitE,
			TestCommitF: commitF,
			TestCommitG: commitG,
			TestCommitH: commitH,
			TestCommitI: commitI,
			TestCommitJ: commitJ,
		},
		Notes: map[string]map[string]string{
			TestRequestsRef: map[string]string{
				TestCommitB: TestRequestB,
				TestCommitD: TestRequestD,
				TestCommitG: TestRequestG,
			},
			TestCommentsRef: map[string]string{
				TestCommitB: TestDiscussB,
				TestCommitD: TestDiscussD,
			},
		},
	}
}

// GetPath returns the path to the repo.
func (r mockRepoForTest) GetPath() string { return "~/mockRepo/" }

// GetRepoStateHash returns a hash which embodies the entire current state of a repository.
func (r mockRepoForTest) GetRepoStateHash() (string, error) {
	repoJSON, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha1.Sum([]byte(repoJSON))), nil
}

// GetUserEmail returns the email address that the user has used to configure git.
func (r mockRepoForTest) GetUserEmail() (string, error) { return "user@example.com", nil }

// GetCoreEditor returns the name of the editor that the user has used to configure git.
func (r mockRepoForTest) GetCoreEditor() (string, error) { return "vi", nil }

// HasUncommittedChanges returns true if there are local, uncommitted changes.
func (r mockRepoForTest) HasUncommittedChanges() (bool, error) { return false, nil }

func (r mockRepoForTest) resolveLocalRef(ref string) (string, error) {
	if commit, ok := r.Refs[ref]; ok {
		return commit, nil
	}
	if _, ok := r.Commits[ref]; ok {
		return ref, nil
	}
	return "", fmt.Errorf("The ref %q does not exist", ref)
}

// VerifyCommit verifies that the supplied hash points to a known commit.
func (r mockRepoForTest) VerifyCommit(hash string) error {
	if _, ok := r.Commits[hash]; !ok {
		return fmt.Errorf("The given hash %q is not a known commit", hash)
	}
	return nil
}

// VerifyGitRef verifies that the supplied ref points to a known commit.
func (r mockRepoForTest) VerifyGitRef(ref string) error {
	_, err := r.resolveLocalRef(ref)
	return err
}

// GetHeadRef returns the ref that is the current HEAD.
func (r mockRepoForTest) GetHeadRef() (string, error) { return r.Head, nil }

// GetCommitHash returns the hash of the commit pointed to by the given ref.
func (r mockRepoForTest) GetCommitHash(ref string) (string, error) {
	err := r.VerifyGitRef(ref)
	if err != nil {
		return "", err
	}
	return r.Refs[ref], nil
}

// ResolveRefCommit returns the commit pointed to by the given ref, which may be a remote ref.
//
// This differs from GetCommitHash which only works on exact matches, in that it will try to
// intelligently handle the scenario of a ref not existing locally, but being known to exist
// in a remote repo.
//
// This method should be used when a command may be performed by either the reviewer or the
// reviewee, while GetCommitHash should be used when the encompassing command should only be
// performed by the reviewee.
func (r mockRepoForTest) ResolveRefCommit(ref string) (string, error) {
	if commit, err := r.resolveLocalRef(ref); err == nil {
		return commit, err
	}
	return r.resolveLocalRef(strings.Replace(ref, "refs/heads/", "refs/remotes/origin/", 1))
}

func (r mockRepoForTest) getCommit(ref string) (mockCommit, error) {
	commit, err := r.resolveLocalRef(ref)
	return r.Commits[commit], err
}

// GetCommitMessage returns the message stored in the commit pointed to by the given ref.
func (r mockRepoForTest) GetCommitMessage(ref string) (string, error) {
	commit, err := r.getCommit(ref)
	if err != nil {
		return "", err
	}
	return commit.Message, nil
}

// GetCommitTime returns the commit time of the commit pointed to by the given ref.
func (r mockRepoForTest) GetCommitTime(ref string) (string, error) {
	commit, err := r.getCommit(ref)
	if err != nil {
		return "", err
	}
	return commit.Time, nil
}

// GetLastParent returns the last parent of the given commit (as ordered by git).
func (r mockRepoForTest) GetLastParent(ref string) (string, error) {
	commit, err := r.getCommit(ref)
	if len(commit.Parents) > 0 {
		return commit.Parents[len(commit.Parents)-1], err
	}
	return "", err
}

// GetCommitDetails returns the details of a commit's metadata.
func (r mockRepoForTest) GetCommitDetails(ref string) (*CommitDetails, error) {
	commit, err := r.getCommit(ref)
	if err != nil {
		return nil, err
	}
	var details CommitDetails
	details.Author = "Test Author"
	details.AuthorEmail = "author@example.com"
	details.Summary = commit.Message
	details.Time = commit.Time
	details.Parents = commit.Parents
	return &details, nil
}

// ancestors returns the breadth-first traversal of a commit's ancestors
func (r mockRepoForTest) ancestors(commit string) ([]string, error) {
	queue := []string{commit}
	var ancestors []string
	for queue != nil {
		var nextQueue []string
		for _, c := range queue {
			commit, err := r.getCommit(c)
			if err != nil {
				return nil, err
			}
			parents := commit.Parents
			nextQueue = append(nextQueue, parents...)
			ancestors = append(ancestors, parents...)
		}
		queue = nextQueue
	}
	return ancestors, nil
}

// IsAncestor determines if the first argument points to a commit that is an ancestor of the second.
func (r mockRepoForTest) IsAncestor(ancestor, descendant string) (bool, error) {
	if ancestor == descendant {
		return true, nil
	}
	descendantCommit, err := r.getCommit(descendant)
	if err != nil {
		return false, err
	}
	for _, parent := range descendantCommit.Parents {
		if t, e := r.IsAncestor(ancestor, parent); e == nil && t {
			return true, nil
		}
	}
	return false, nil
}

// MergeBase determines if the first commit that is an ancestor of the two arguments.
func (r mockRepoForTest) MergeBase(a, b string) (string, error) {
	ancestors, err := r.ancestors(a)
	if err != nil {
		return "", err
	}
	for _, ancestor := range ancestors {
		if t, e := r.IsAncestor(ancestor, b); e == nil && t {
			return ancestor, nil
		}
	}
	return "", nil
}

// Diff computes the diff between two given commits.
func (r mockRepoForTest) Diff(left, right string, diffArgs ...string) (string, error) {
	return fmt.Sprintf("Diff between %q and %q", left, right), nil
}

// Show returns the contents of the given file at the given commit.
func (r mockRepoForTest) Show(commit, path string) (string, error) {
	return fmt.Sprintf("%s:%s", commit, path), nil
}

// SwitchToRef changes the currently-checked-out ref.
func (r mockRepoForTest) SwitchToRef(ref string) error {
	r.Head = ref
	return nil
}

// MergeRef merges the given ref into the current one.
//
// The ref argument is the ref to merge, and fastForward indicates that the
// current ref should only move forward, as opposed to creating a bubble merge.
func (r mockRepoForTest) MergeRef(ref string, fastForward bool, messages ...string) error { return nil }

// RebaseRef rebases the given ref into the current one.
func (r mockRepoForTest) RebaseRef(ref string) error { return nil }

// ListCommitsBetween returns the list of commits between the two given revisions.
//
// The "from" parameter is the starting point (exclusive), and the "to" parameter
// is the ending point (inclusive). If the commit pointed to by the "from" parameter
// is not an ancestor of the commit pointed to by the "to" parameter, then the
// merge base of the two is used as the starting point.
//
// The generated list is in chronological order (with the oldest commit first).
func (r mockRepoForTest) ListCommitsBetween(from, to string) ([]string, error) { return nil, nil }

// GetNotes reads the notes from the given ref that annotate the given revision.
func (r mockRepoForTest) GetNotes(notesRef, revision string) []Note {
	notesText := r.Notes[notesRef][revision]
	var notes []Note
	for _, line := range strings.Split(notesText, "\n") {
		notes = append(notes, Note(line))
	}
	return notes
}

// AppendNote appends a note to a revision under the given ref.
func (r mockRepoForTest) AppendNote(ref, revision string, note Note) error {
	existingNotes := r.Notes[ref][revision]
	newNotes := existingNotes + "\n" + string(note)
	r.Notes[ref][revision] = newNotes
	return nil
}

// ListNotedRevisions returns the collection of revisions that are annotated by notes in the given ref.
func (r mockRepoForTest) ListNotedRevisions(notesRef string) []string {
	var revisions []string
	for revision := range r.Notes[notesRef] {
		if _, ok := r.Commits[revision]; ok {
			revisions = append(revisions, revision)
		}
	}
	return revisions
}

// PushNotes pushes git notes to a remote repo.
func (r mockRepoForTest) PushNotes(remote, notesRefPattern string) error { return nil }

// PullNotes fetches the contents of the given notes ref from a remote repo,
// and then merges them with the corresponding local notes using the
// "cat_sort_uniq" strategy.
func (r mockRepoForTest) PullNotes(remote, notesRefPattern string) error { return nil }
