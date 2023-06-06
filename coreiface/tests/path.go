package tests

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/ipfs/boxo/coreiface/options"
	"github.com/ipfs/boxo/path"
	ipldcbor "github.com/ipfs/go-ipld-cbor"
	"github.com/stretchr/testify/require"
)

func (tp *TestSuite) TestPath(t *testing.T) {
	t.Run("TestMutablePath", tp.TestMutablePath)
	t.Run("TestPathRemainder", tp.TestPathRemainder)
	t.Run("TestEmptyPathRemainder", tp.TestEmptyPathRemainder)
	t.Run("TestInvalidPathRemainder", tp.TestInvalidPathRemainder)
	t.Run("TestPathRoot", tp.TestPathRoot)
	t.Run("TestPathJoin", tp.TestPathJoin)
}

func (tp *TestSuite) TestMutablePath(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	api, err := tp.makeAPI(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	blk, err := api.Block().Put(ctx, strings.NewReader(`foo`))
	if err != nil {
		t.Fatal(err)
	}

	if blk.Path().Mutable() {
		t.Error("expected /ipld path to be immutable")
	}

	// get self /ipns path

	if api.Key() == nil {
		t.Fatal(".Key not implemented")
	}

	keys, err := api.Key().List(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !keys[0].Path().Mutable() {
		t.Error("expected self /ipns path to be mutable")
	}
}

func (tp *TestSuite) TestPathRemainder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	api, err := tp.makeAPI(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	if api.Dag() == nil {
		t.Fatal(".Dag not implemented")
	}

	nd, err := ipldcbor.FromJSON(strings.NewReader(`{"foo": {"bar": "baz"}}`), math.MaxUint64, -1)
	if err != nil {
		t.Fatal(err)
	}

	if err := api.Dag().Add(ctx, nd); err != nil {
		t.Fatal(err)
	}

	p, err := path.NewPath(nd.String() + "/foo/bar")
	if err != nil {
		t.Fatal(err)
	}

	rp1, err := api.ResolvePath(ctx, p)
	if err != nil {
		t.Fatal(err)
	}

	if rp1.Remainder() != "foo/bar" {
		t.Error("expected to get path remainder")
	}
}

func (tp *TestSuite) TestEmptyPathRemainder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	api, err := tp.makeAPI(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	if api.Dag() == nil {
		t.Fatal(".Dag not implemented")
	}

	nd, err := ipldcbor.FromJSON(strings.NewReader(`{"foo": {"bar": "baz"}}`), math.MaxUint64, -1)
	if err != nil {
		t.Fatal(err)
	}

	if err := api.Dag().Add(ctx, nd); err != nil {
		t.Fatal(err)
	}

	rp1, err := api.ResolvePath(ctx, path.NewIPFSPath(nd.Cid()))
	if err != nil {
		t.Fatal(err)
	}

	if rp1.Remainder() != "" {
		t.Error("expected the resolved path to not have a remainder")
	}
}

func (tp *TestSuite) TestInvalidPathRemainder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	api, err := tp.makeAPI(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	if api.Dag() == nil {
		t.Fatal(".Dag not implemented")
	}

	nd, err := ipldcbor.FromJSON(strings.NewReader(`{"foo": {"bar": "baz"}}`), math.MaxUint64, -1)
	if err != nil {
		t.Fatal(err)
	}

	if err := api.Dag().Add(ctx, nd); err != nil {
		t.Fatal(err)
	}

	p, err := path.Join(path.NewIPLDPath(nd.Cid()), "/bar/baz")
	if err != nil {
		t.Fatal(err)
	}

	_, err = api.ResolvePath(ctx, p)
	if err == nil || !strings.Contains(err.Error(), `no link named "bar"`) {
		t.Fatalf("unexpected error: %s", err)
	}
}

func (tp *TestSuite) TestPathRoot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	api, err := tp.makeAPI(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	if api.Block() == nil {
		t.Fatal(".Block not implemented")
	}

	blk, err := api.Block().Put(ctx, strings.NewReader(`foo`), options.Block.Format("raw"))
	if err != nil {
		t.Fatal(err)
	}

	if api.Dag() == nil {
		t.Fatal(".Dag not implemented")
	}

	nd, err := ipldcbor.FromJSON(strings.NewReader(`{"foo": {"/": "`+blk.Path().Cid().String()+`"}}`), math.MaxUint64, -1)
	if err != nil {
		t.Fatal(err)
	}

	if err := api.Dag().Add(ctx, nd); err != nil {
		t.Fatal(err)
	}

	p, err := path.Join(path.NewIPLDPath(nd.Cid()), "/foo")
	if err != nil {
		t.Fatal(err)
	}

	rp, err := api.ResolvePath(ctx, p)
	if err != nil {
		t.Fatal(err)
	}

	if rp.Root().String() != nd.Cid().String() {
		t.Error("unexpected path root")
	}

	if rp.Cid().String() != blk.Path().Cid().String() {
		t.Error("unexpected path cid")
	}
}

func (tp *TestSuite) TestPathJoin(t *testing.T) {
	p1, err := path.NewPath("/ipfs/QmYNmQKp6SuaVrpgWRsPTgCQCnpxUYGq76YEKBXuj2N4H6/bar/baz")
	require.NoError(t, err)

	p2, err := path.Join(p1, "foo")
	require.NoError(t, err)

	require.Equal(t, "/ipfs/QmYNmQKp6SuaVrpgWRsPTgCQCnpxUYGq76YEKBXuj2N4H6/bar/baz/foo", p2.String())
}
