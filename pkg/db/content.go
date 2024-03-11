package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/gohugonet/hugoverse/internal/domain/content/entity"
	"github.com/gorilla/schema"
	bolt "go.etcd.io/bbolt"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Content retrives one item from the database. Non-existent values will return an empty []byte
// The `target` argument is a string made up of namespace:id (string:int)
func Content(target string) ([]byte, error) {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	val := &bytes.Buffer{}
	err := store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ns))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		_, err := val.Write(b.Get([]byte(id)))
		if err != nil {
			log.Println(err)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return val.Bytes(), nil
}

// SetContent inserts/replaces values in the database.
// The `target` argument is a string made up of namespace:id (string:int)
func SetContent(target string, data url.Values) (int, error) {
	t := strings.Split(target, ":")
	ns, id := t[0], t[1]

	// check if content id == -1 (indicating new post).
	// if so, run an insert which will assign the next auto incremented int.
	// this is done because bbolt begins its bucket auto increment value at 0,
	// which is the zero-value of an int in the Item struct field for ID.
	// this is a problem when the original first post (with auto ID = 0) gets
	// overwritten by any new post, originally having no ID, defauting to 0.
	if id == "-1" {
		return insert(ns, data)
	}

	return 0, nil

	//return update(ns, id, data, nil)
}

func insert(ns string, data url.Values) (int, error) {
	var effectedID int
	var specifier string // i.e. __pending, __sorted, etc.
	if strings.Contains(ns, "__") {
		spec := strings.Split(ns, "__")
		ns = spec[0]
		specifier = "__" + spec[1]
	}

	fmt.Printf("insert: ns: %s, specifier: %s\n", ns, specifier)

	var j []byte
	var cid string
	err := store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(ns + specifier))
		if err != nil {
			return err
		}

		// get the next available ID and convert to string
		// also set effectedID to int of ID
		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		cid = strconv.FormatUint(id, 10)
		effectedID, err = strconv.Atoi(cid)
		if err != nil {
			return err
		}
		data.Set("id", cid)

		// add UUID to data for use in embedded Item
		uid, err := uuid.NewV4()
		if err != nil {
			return err
		}

		data.Set("uuid", uid.String())

		// if type has a specifier, add it to data for downstream processing
		if specifier != "" {
			data.Set("__specifier", specifier)
		}

		j, err = postToJSON(ns, data)
		if err != nil {
			return err
		}

		err = b.Put([]byte(cid), j)
		if err != nil {
			return err
		}

		//// store the slug,type:id in contentIndex if public content
		//if specifier == "" {
		//	ci := tx.Bucket([]byte("__contentIndex"))
		//	if ci == nil {
		//		return bolt.ErrBucketNotFound
		//	}
		//
		//	k := []byte(data.Get("slug"))
		//	v := []byte(fmt.Sprintf("%s:%d", ns, effectedID))
		//	err := ci.Put(k, v)
		//	if err != nil {
		//		return err
		//	}
		//}

		return nil
	})
	if err != nil {
		fmt.Printf("??? error: %v\n", err)
		return 0, err
	}

	if specifier == "" {
		go SortContent(ns)
	}

	// todo

	//// insert changes data, so invalidate client caching
	//err = InvalidateCache()
	//if err != nil {
	//	return 0, err
	//}
	//
	//go func() {
	//	// add data to search index
	//	target := fmt.Sprintf("%s:%s", ns, cid)
	//	err = search.UpdateIndex(target, j)
	//	if err != nil {
	//		log.Println("[search] UpdateIndex Error:", err)
	//	}
	//}()

	return effectedID, nil
}

func postToJSON(ns string, data url.Values) ([]byte, error) {
	// find the content type and decode values into it
	t := func() interface{} { return new(entity.Demo) }
	post := t()

	// check for any multi-value fields (ex. checkbox fields)
	// and correctly format for db storage. Essentially, we need
	// fieldX.0: value1, fieldX.1: value2 => fieldX: []string{value1, value2}
	fieldOrderValue := make(map[string]map[string][]string)
	for k, v := range data {
		if strings.Contains(k, ".") {
			fo := strings.Split(k, ".")

			// put the order and the field value into map
			field := string(fo[0])
			order := string(fo[1])
			if len(fieldOrderValue[field]) == 0 {
				fieldOrderValue[field] = make(map[string][]string)
			}

			// orderValue is 0:[?type=Thing&id=1]
			orderValue := fieldOrderValue[field]
			orderValue[order] = v
			fieldOrderValue[field] = orderValue

			// discard the post form value with name.N
			data.Del(k)
		}
	}

	// add/set the key & value to the post form in order
	for f, ov := range fieldOrderValue {
		for i := 0; i < len(ov); i++ {
			position := fmt.Sprintf("%d", i)
			fieldValue := ov[position]

			if data.Get(f) == "" {
				for i, fv := range fieldValue {
					if i == 0 {
						data.Set(f, fv)
					} else {
						data.Add(f, fv)
					}
				}
			} else {
				for _, fv := range fieldValue {
					data.Add(f, fv)
				}
			}
		}
	}

	dec := schema.NewDecoder()
	dec.SetAliasTag("json")     // allows simpler struct tagging when creating a content type
	dec.IgnoreUnknownKeys(true) // will skip over form values submitted, but not in struct
	err := dec.Decode(post, data)
	if err != nil {
		return nil, err
	}

	//// if the content has no slug, and has no specifier, create a slug, check it
	//// for duplicates, and add it to our values
	//if data.Get("slug") == "" && data.Get("__specifier") == "" {
	//	slug, err := item.Slug(post.(item.Identifiable))
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	slug, err = checkSlugForDuplicate(slug)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	post.(item.Sluggable).SetSlug(slug)
	//	data.Set("slug", slug)
	//}

	// marshall content struct to json for db storage
	j, err := json.Marshal(post)
	if err != nil {
		return nil, err
	}

	return j, nil
}

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

	// only sort main content types i.e. Post
	if strings.Contains(namespace, "__") {
		return
	}

	all := ContentAll(namespace)

	var posts sortableContent
	// decode each (json) into type to then sort
	for i := range all {
		j := all[i]

		t := func() interface{} { return new(entity.Demo) }
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
