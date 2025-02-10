package database

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/mdfriday/hugoverse/internal/interfaces/api/token"
	"github.com/mdfriday/hugoverse/pkg/db"
	"net/http"
)

func (d *Database) Open(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		email, err := token.GetEmail(req)
		if err != nil {
			d.log.Errorf("Error getting email: %v", err)
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := d.StartUserDatabase(email); err != nil {
			d.log.Errorf("Error starting user database: %v", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(res, req)
	})
}

func (d *Database) StartUserDatabase(email string) error {
	var buckets []string
	buckets = append(buckets, d.contentBuckets...)
	buckets = append(buckets, userBuckets...)

	ud := hashEmailMD5(email)
	s, err := db.OpenUserStore(ud, d.dataDir, buckets)
	if err != nil {
		return err
	}

	d.userStore = s
	d.userDir = ud

	d.log.Debugf("Started user database: %s", ud)

	return nil
}

func hashEmailMD5(email string) string {
	hash := md5.New()
	hash.Write([]byte(email))

	return hex.EncodeToString(hash.Sum(nil))
}
