package feed

import "proj1/lock"

//Feed represents a user's twitter feed
// You will add to this interface the implementations as you complete them.
type Feed interface {
	Add(body string, timestamp float64)
	Remove(timestamp float64) bool
	Contains(timestamp float64) bool
	CurrentPosts() ([]string, []float64)
}

//feed is the internal representation of a user's twitter feed (hidden from outside packages)
// You CAN add to this structure but you cannot remove any of the original fields. You must use
// the original fields in your implementation. You can assume the feed will not have duplicate posts
type feed struct {
	start *post // a pointer to the beginning post
	rwLock *lock.RWLock	//  a read/write lock
}

//post is the internal representation of a post on a user's twitter feed (hidden from outside packages)
// You CAN add to this structure but you cannot remove any of the original fields. You must use
// the original fields in your implementation.
type post struct {
	body      string // the text of the post
	timestamp float64  // Unix timestamp of the post
	next      *post  // the next post in the feed
}

//NewPost creates and returns a new post value given its body and timestamp
func newPost(body string, timestamp float64, next *post) *post {
	return &post{body, timestamp, next}
}

//NewFeed creates a empy user feed
func NewFeed() Feed {
	lock := lock.NewRWLock()
	return &feed{start: nil, rwLock: lock}
}

// Add inserts a new post to the feed. The feed is always ordered by the timestamp where
// the most recent timestamp is at the beginning of the feed followed by the second most
// recent timestamp, etc. You may need to insert a new post somewhere in the feed because
// the given timestamp may not be the most recent.
func (f *feed) Add(body string, timestamp float64) {
	f.rwLock.Lock()
	newFeed := newPost(body, timestamp, nil)

	pred := f.start

	if pred == nil{
		f.start = newFeed
		f.rwLock.Unlock()
		return
	}
	if pred.timestamp <= newFeed.timestamp{
		newFeed.next = pred
		f.start = newFeed
		f.rwLock.Unlock()
		return
	}

	curr := pred
	for curr != nil && curr.timestamp > newFeed.timestamp{
		pred = curr
		curr = curr.next
	}

	pred.next = newFeed
	newFeed.next = curr
	f.rwLock.Unlock()
}

// Remove deletes the post with the given timestamp. If the timestamp
// is not included in a post of the feed then the feed remains
// unchanged. Return true if the deletion was a success, otherwise return false
func (f *feed) Remove(timestamp float64) bool {

	f.rwLock.Lock()
	if f.start == nil{
		f.rwLock.Unlock()
		return false
	}
	if f.start.timestamp == timestamp{
		f.start = f.start.next
		f.rwLock.Unlock()
		return true
	}
	pred := f.start
	curr := pred

	for curr != nil && curr.timestamp >= timestamp{
		if curr.timestamp == timestamp{
			pred.next = curr.next
			f.rwLock.Unlock()
			return true
		}
		pred = curr
		curr = curr.next
	}
	f.rwLock.Unlock()
	return false

}

// Contains determines whether a post with the given timestamp is
// inside a feed. The function returns true if there is a post
// with the timestamp, otherwise, false.
func (f *feed) Contains(timestamp float64) bool {

	f.rwLock.RLock()
	curr := f.start

	for curr != nil{
		if curr.timestamp == timestamp{
			f.rwLock.RUnlock()
			return true
		}
		curr = curr.next
	}
	f.rwLock.RUnlock()
	return false
}

// return the current contents in feed in separated slices
func (f *feed) CurrentPosts() ([]string, []float64){
	body := make([]string, 0)
	timestamps := make([]float64, 0)

	f.rwLock.RLock()

	curr := f.start
	for curr != nil{
		body = append(body, curr.body)
		timestamps = append(timestamps, curr.timestamp)
		curr = curr.next
	}

	f.rwLock.RUnlock()
	return body, timestamps
}


