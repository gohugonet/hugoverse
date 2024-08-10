package entity

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/domain/resources"
	"github.com/gohugonet/hugoverse/internal/domain/resources/valueobject"
	"github.com/gohugonet/hugoverse/pkg/constants"
	"hash"
	"io"
)

const defaultHashAlgo = "sha256"

type IntegrityClient struct{}

type fingerprintTransformation struct {
	algo string
}

func (t *fingerprintTransformation) Key() valueobject.ResourceTransformationKey {
	return valueobject.NewResourceTransformationKey(constants.ResourceTransformationFingerprint, t.algo)
}

// Transform creates a MD5 hash of the Resource content and inserts that hash before
// the extension in the filename.
func (t *fingerprintTransformation) Transform(ctx *valueobject.ResourceTransformationCtx) error {
	h, err := newHash(t.algo)
	if err != nil {
		return err
	}

	var w io.Writer
	if rc, ok := ctx.Source.From.(io.ReadSeeker); ok {
		// This transformation does not change the content, so try to
		// avoid writing to To if we can.
		defer rc.Seek(0, 0)
		w = h
	} else {
		w = io.MultiWriter(h, ctx.Target.To)
	}

	io.Copy(w, ctx.Source.From)
	d, err := digest(h)
	if err != nil {
		return err
	}

	ctx.Data["Integrity"] = integrity(t.algo, d)
	ctx.AddOutPathIdentifier("." + hex.EncodeToString(d[:]))
	return nil
}

func newHash(algo string) (hash.Hash, error) {
	switch algo {
	case "md5":
		return md5.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha384":
		return sha512.New384(), nil
	case "sha512":
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("unsupported hash algorithm: %q, use either md5, sha256, sha384 or sha512", algo)
	}
}

// Fingerprint applies fingerprinting of the given resource and hash algorithm.
// It defaults to sha256 if none given, and the options are md5, sha256 or sha512.
// The same algo is used for both the fingerprinting part (aka cache busting) and
// the base64-encoded Subresource Integrity hash, so you will have to stay away from
// md5 if you plan to use both.
// See https://developer.mozilla.org/en-US/docs/Web/Security/Subresource_Integrity
func (c *IntegrityClient) Fingerprint(res resources.Resource, algo string) (resources.Resource, error) {
	if algo == "" {
		algo = defaultHashAlgo
	}
	transRes := res.(Transformer)
	return transRes.Transform(&fingerprintTransformation{algo: algo})
}

func integrity(algo string, sum []byte) string {
	encoded := base64.StdEncoding.EncodeToString(sum)
	return algo + "-" + encoded
}

func digest(h hash.Hash) ([]byte, error) {
	sum := h.Sum(nil)
	return sum, nil
}
