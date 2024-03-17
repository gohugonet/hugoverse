package db

import (
	"encoding/json"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/content/entity"
	bolt "go.etcd.io/bbolt"
	"log"
	"sort"
	"sync"
	"time"
)

// ContentAll retrives all items from the database within the provided namespace
func ContentAll(namespace string) [][]byte {
	var posts [][]byte
	store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(namespace))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		numKeys := b.Stats().KeyN
		posts = make([][]byte, 0, numKeys)

		b.ForEach(func(k, v []byte) error {
			posts = append(posts, v)

			return nil
		})

		return nil
	})

	return posts
}

// Sortable ensures data is sortable by time
type Sortable interface {
	Time() int64
	Touch() int64
}

type sortableContent []Sortable

func (s sortableContent) Len() int {
	return len(s)
}

func (s sortableContent) Less(i, j int) bool {
	return s[i].Time() > s[j].Time()
}

func (s sortableContent) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// SortContent sorts all content of the type supplied as the namespace by time,
// in descending order, from most recent to least recent
// Should be called from a goroutine after SetContent is successful
func SortContent(namespace string) {
	// wait if running too frequently per namespace
	if !enoughTime(namespace) {
		return
	}

	all := ContentAll(namespace)

	var posts sortableContent
	// decode each (json) into type to then sort
	for i := range all {
		j := all[i]

		t := func() interface{} { return new(entity.Song) }
		post := t()

		err := json.Unmarshal(j, &post)
		if err != nil {
			log.Println("Error decoding json while sorting", namespace, ":", err)
			return
		}

		posts = append(posts, post.(Sortable))
	}

	// sort posts
	sort.Sort(posts)

	// marshal posts to json
	var bb [][]byte
	for i := range posts {
		j, err := json.Marshal(posts[i])
		if err != nil {
			// log error and kill sort so __sorted is not in invalid state
			log.Println("Error marshal post to json in SortContent:", err)
			return
		}

		bb = append(bb, j)
	}

	// store in <namespace>_sorted bucket, first delete existing
	err := store.Update(func(tx *bolt.Tx) error {
		bname := []byte(namespace + "__sorted")
		err := tx.DeleteBucket(bname)
		if err != nil && err != bolt.ErrBucketNotFound {
			return err
		}

		b, err := tx.CreateBucketIfNotExists(bname)
		if err != nil {
			return err
		}

		// encode to json and store as 'post.Time():i':post
		for i := range bb {
			cid := fmt.Sprintf("%d:%d", posts[i].Time(), i)
			err = b.Put([]byte(cid), bb[i])
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Println("Error while updating db with sorted", namespace, err)
	}

}

func enoughTime(key string) bool {
	last, ok := lastInvocation(key)
	if !ok {
		// no invocation yet
		// track next invocation
		setLastInvocation(key)
		return true
	}

	// if our required wait time has been met, return true
	if time.Now().After(last.Add(waitDuration)) {
		setLastInvocation(key)
		return true
	}

	// dispatch a delayed invocation in case no additional one follows
	go func() {
		lastInvocationBeforeTimer, _ := lastInvocation(key) // zero value can be handled, no need for ok
		enoughTimer := time.NewTimer(waitDuration)
		<-enoughTimer.C
		lastInvocationAfterTimer, _ := lastInvocation(key)
		if !lastInvocationAfterTimer.After(lastInvocationBeforeTimer) {
			SortContent(key)
		}
	}()

	return false
}

var sortContentCalls = make(map[string]time.Time)
var waitDuration = time.Millisecond * 2000
var sortMutex = &sync.Mutex{}

func setLastInvocation(key string) {
	sortMutex.Lock()
	sortContentCalls[key] = time.Now()
	sortMutex.Unlock()
}

func lastInvocation(key string) (time.Time, bool) {
	sortMutex.Lock()
	last, ok := sortContentCalls[key]
	sortMutex.Unlock()
	return last, ok
}
